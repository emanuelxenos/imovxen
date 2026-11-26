package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AsaasClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewAsaasClient(apiKey string, isProduction bool) *AsaasClient {
	baseURL := "https://sandbox.asaas.com/api/v3"
	if isProduction {
		baseURL = "https://api.asaas.com/v3"
	}

	return &AsaasClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type AsaasCustomerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	CpfCnpj  string `json:"cpfCnpj,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type AsaasCustomerResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AsaasPaymentRequest struct {
	Customer            string  `json:"customer"`
	BillingType         string  `json:"billingType"` // PIX, BOLETO, CREDIT_CARD
	Value               float64 `json:"value"`
	DueDate             string  `json:"dueDate"`
	Description         string  `json:"description"`
	PostalService       bool    `json:"postalService"`
}

type AsaasPaymentResponse struct {
	ID          string  `json:"id"`
	Customer    string  `json:"customer"`
	Value       float64 `json:"value"`
	Status      string  `json:"status"` // PENDING, RECEIVED, CONFIRMED, OVERDUE
	BillingType string  `json:"billingType"`
	InvoiceURL  string  `json:"invoiceUrl"`
}

type AsaasPixQrCodeResponse struct {
	EncodedImage string `json:"encodedImage"`
	Payload      string `json:"payload"`
	ExpirationDate string `json:"expirationDate"`
}

// CreateCustomer cria ou busca um cliente no Asaas
func (c *AsaasClient) CreateCustomer(req AsaasCustomerRequest) (*AsaasCustomerResponse, error) {
	if c.apiKey == "" {
		// Mock local simulado caso nenhuma chave Asaas esteja configurada
		return &AsaasCustomerResponse{
			ID:    "cus_mock_" + fmt.Sprintf("%d", time.Now().Unix()),
			Name:  req.Name,
			Email: req.Email,
		}, nil
	}

	bodyBytes, _ := json.Marshal(req)
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/customers", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("access_token", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("asaas customer error (%d): %s", resp.StatusCode, string(respBody))
	}

	var res AsaasCustomerResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CreatePixPayment cria uma cobrança via Pix
func (c *AsaasClient) CreatePixPayment(customerID string, value float64, description string) (*AsaasPaymentResponse, *AsaasPixQrCodeResponse, error) {
	if c.apiKey == "" {
		// Simulação completa para testes locais caso não haja credencial Asaas
		mockID := "pay_mock_" + fmt.Sprintf("%d", time.Now().Unix())
		mockPay := &AsaasPaymentResponse{
			ID:          mockID,
			Customer:    customerID,
			Value:       value,
			Status:      "PENDING",
			BillingType: "PIX",
			InvoiceURL:  "https://sandbox.asaas.com/i/" + mockID,
		}
		mockPix := &AsaasPixQrCodeResponse{
			EncodedImage: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			Payload:      "00020126580014BR.GOV.BCB.PIX0136imovxen-pix-mock-payload520400005303986540" + fmt.Sprintf("%.2f", value) + "5802BR5913IMOVXEN SAAS6009SAO PAULO62070503***6304ABCD",
			ExpirationDate: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}
		return mockPay, mockPix, nil
	}

	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	payReq := AsaasPaymentRequest{
		Customer:      customerID,
		BillingType:   "PIX",
		Value:         value,
		DueDate:       dueDate,
		Description:   description,
		PostalService: false,
	}

	bodyBytes, _ := json.Marshal(payReq)
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/payments", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("access_token", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("asaas payment error (%d): %s", resp.StatusCode, string(respBody))
	}

	var payRes AsaasPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payRes); err != nil {
		return nil, nil, err
	}

	// Buscar QR Code Pix da cobrança
	qrReq, err := http.NewRequest(http.MethodGet, c.baseURL+"/payments/"+payRes.ID+"/pixQrCode", nil)
	if err != nil {
		return &payRes, nil, nil
	}
	qrReq.Header.Set("access_token", c.apiKey)

	qrResp, err := c.httpClient.Do(qrReq)
	if err != nil {
		return &payRes, nil, nil
	}
	defer qrResp.Body.Close()

	var qrRes AsaasPixQrCodeResponse
	if err := json.NewDecoder(qrResp.Body).Decode(&qrRes); err != nil {
		return &payRes, nil, nil
	}

	return &payRes, &qrRes, nil
}
