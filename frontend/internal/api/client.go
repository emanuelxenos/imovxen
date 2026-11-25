package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RequestOption permite passar tokens e refresh tokens para revalidação silenciosa
type RequestOption func(*requestConfig)

type requestConfig struct {
	token        string
	refreshToken string
	onRefresh    func(newAccess, newRefresh string)
}

func WithAuth(token, refreshToken string, onRefresh func(newAccess, newRefresh string)) RequestOption {
	return func(rc *requestConfig) {
		rc.token = token
		rc.refreshToken = refreshToken
		rc.onRefresh = onRefresh
	}
}

func (c *Client) Get(endpoint string, token string, out interface{}) error {
	return c.doRequest("GET", endpoint, nil, token, "", nil, out)
}

func (c *Client) GetWithAutoRefresh(endpoint string, token, refreshToken string, onRefresh func(newAccess, newRefresh string), out interface{}) error {
	return c.doRequest("GET", endpoint, nil, token, refreshToken, onRefresh, out)
}

func (c *Client) Post(endpoint string, body interface{}, token string, out interface{}) error {
	return c.doRequest("POST", endpoint, body, token, "", nil, out)
}

func (c *Client) Put(endpoint string, body interface{}, token string, out interface{}) error {
	return c.doRequest("PUT", endpoint, body, token, "", nil, out)
}

func (c *Client) Patch(endpoint string, body interface{}, token string, out interface{}) error {
	return c.doRequest("PATCH", endpoint, body, token, "", nil, out)
}

func (c *Client) Delete(endpoint string, token string, out interface{}) error {
	return c.doRequest("DELETE", endpoint, nil, token, "", nil, out)
}

func (c *Client) doRequest(method, endpoint string, body interface{}, token, refreshToken string, onRefresh func(newAccess, newRefresh string), out interface{}) error {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erro ao serializar body: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Se retornou 401 e temos refresh token, tenta renovar silenciosamente
	if resp.StatusCode == http.StatusUnauthorized && refreshToken != "" {
		newAccess, newRefresh, refErr := c.RefreshToken(refreshToken)
		if refErr == nil && newAccess != "" {
			if onRefresh != nil {
				onRefresh(newAccess, newRefresh)
			}
			// Repete a requisição original com o novo token
			retryReq, retryErr := http.NewRequest(method, c.baseURL+endpoint, bytes.NewReader(bodyBytes))
			if retryErr == nil {
				if body != nil {
					retryReq.Header.Set("Content-Type", "application/json")
				}
				retryReq.Header.Set("Authorization", "Bearer "+newAccess)
				retryResp, doErr := c.httpClient.Do(retryReq)
				if doErr == nil {
					defer retryResp.Body.Close()
					if retryResp.StatusCode < 400 && out != nil {
						return json.NewDecoder(retryResp.Body).Decode(out)
					}
					return nil
				}
			}
		}
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("erro da API HTTP %d", resp.StatusCode)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) RefreshToken(refreshToken string) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	payload := map[string]string{"refresh_token": refreshToken}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/api/auth/refresh", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("falha ao renovar token: HTTP %d", resp.StatusCode)
	}

	var res RefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", "", err
	}
	return res.AccessToken, res.RefreshToken, nil
}
