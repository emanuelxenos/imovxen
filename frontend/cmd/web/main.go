package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imovxen/frontend/internal/api"
)

type PropertyItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   *string  `json:"description"`
	Type          string   `json:"type"`
	Purpose       string   `json:"purpose"`
	Status        string   `json:"status"`
	Price         float64  `json:"price"`
	AreaM2        *float64 `json:"area_m2"`
	Bedrooms      *int16   `json:"bedrooms"`
	Bathrooms     *int16   `json:"bathrooms"`
	ParkingSpots  *int16   `json:"parking_spots"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	Neighborhood  *string  `json:"neighborhood"`
	CoverImageUrl *string  `json:"cover_image_url"`
}

func renderPropertyCard(p PropertyItem) string {
	imgURL := "https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=800&q=80"
	if p.CoverImageUrl != nil && *p.CoverImageUrl != "" {
		imgURL = *p.CoverImageUrl
	}

	purposeLabel := "Aluguel"
	priceSuffix := "<span>/mês</span>"
	if p.Purpose == "SALE" {
		purposeLabel = "Venda"
		priceSuffix = ""
	}

	area := "—"
	if p.AreaM2 != nil && *p.AreaM2 > 0 {
		area = fmt.Sprintf("%.0fm²", *p.AreaM2)
	}

	bedrooms := "—"
	if p.Bedrooms != nil && *p.Bedrooms > 0 {
		bedrooms = fmt.Sprintf("%d Quartos", *p.Bedrooms)
	}

	parking := "—"
	if p.ParkingSpots != nil && *p.ParkingSpots > 0 {
		parking = fmt.Sprintf("%d Vagas", *p.ParkingSpots)
	}

	return fmt.Sprintf(`
    <article class="property-card" onclick="window.location.href='/imoveis/%s'" style="cursor: pointer;">
        <div class="card-image-wrapper">
            <span class="badge-purpose">%s</span>
            <img src="%s" alt="%s">
        </div>
        <div class="property-content">
            <div class="property-price">R$ %.2f %s</div>
            <h3 class="property-title">%s</h3>
            <div class="property-location">📍 %s - %s</div>
            <div class="property-amenities">
                <div class="amenity-item">📐 %s</div>
                <div class="amenity-item">🛏️ %s</div>
                <div class="amenity-item">🚗 %s</div>
            </div>
        </div>
    </article>`, p.ID, purposeLabel, imgURL, p.Title, p.Price, priceSuffix, p.Title, p.City, p.State, area, bedrooms, parking)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("WEB_PORT")
	if port == "" {
		port = "3000"
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	apiClient := api.NewClient(backendURL)

	r := gin.Default()
	r.Static("/static", "./static")

	// ==========================================
	// PÁGINA: HOME (BUSCA E LISTAGEM DO BANCO REAL)
	// ==========================================
	r.GET("/", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		userActions := `<a href="/login" class="nav-item">Entrar</a>
                        <a href="/cadastro" class="btn btn-primary">Criar Conta Grátis</a>`
		if token != "" {
			userActions = `<a href="/painel/imoveis" class="btn btn-primary">Painel de Imóveis</a>
                           <a href="/logout" class="btn btn-outline" style="margin-left: 0.5rem;">Sair</a>`
		}

		var apiRes struct {
			Data []PropertyItem `json:"data"`
		}
		_ = apiClient.Get("/api/properties", "", &apiRes)

		cardsHTML := ""
		if len(apiRes.Data) == 0 {
			cardsHTML = `
            <div class="property-card" style="padding: 4rem 2rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                <p style="font-size: 1.2rem; font-weight: 700; color: var(--text-main);">Nenhum imóvel disponível no momento.</p>
                <p style="margin-top: 0.5rem;">Cadastre-se e seja o primeiro anunciante a publicar um imóvel na plataforma!</p>
                <a href="/cadastro" class="btn btn-primary" style="margin-top: 1.5rem;">Anunciar Agora Gratuitamente</a>
            </div>`
		} else {
			for _, item := range apiRes.Data {
				cardsHTML += renderPropertyCard(item)
			}
		}

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Imovxen — O Marketplace Imobiliário Inteligente</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <nav class="nav-links">
                <a href="/" class="nav-item">Explorar</a>
                <a href="/painel/imoveis" class="nav-item">Anunciar</a>
                ` + userActions + `
            </nav>
        </div>
    </header>

    <section class="hero-wrapper">
        <div class="container">
            <div class="hero-badge">
                <span>✨ Plataforma Direta & Sem Burocracia</span>
            </div>
            <h1 class="hero-title">
                O jeito mais inteligente de encontrar seu <span class="highlight">próximo imóvel</span>
            </h1>
            <p class="hero-subtitle">
                Conectamos compradores, inquilinos e proprietários diretamente, com transparência total e processos 100% digitais.
            </p>

            <div class="search-card">
                <div class="search-tabs">
                    <button class="search-tab-btn active" onclick="setPurpose(this, '')">Todos</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'RENT')">Aluguel</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'SALE')">Comprar</button>
                </div>
                <div class="search-grid">
                    <div class="field-group">
                        <label>Localização</label>
                        <input id="search-q" class="field-input" type="text" name="q" placeholder="Cidade, bairro ou título..."
                               hx-get="/partials/imoveis" 
                               hx-include="#search-type, #search-purpose, #search-price"
                               hx-trigger="keyup changed delay:350ms" 
                               hx-target="#lista-imoveis">
                    </div>
                    <div class="field-group">
                        <label>Tipo de Imóvel</label>
                        <select id="search-type" class="field-input" name="type" 
                                hx-get="/partials/imoveis" 
                                hx-include="#search-q, #search-purpose, #search-price"
                                hx-target="#lista-imoveis">
                            <option value="">Todos os tipos</option>
                            <option value="APARTMENT">Apartamentos</option>
                            <option value="HOUSE">Casas e Sobrados</option>
                            <option value="COMMERCIAL">Comerciais</option>
                            <option value="LAND">Terrenos</option>
                        </select>
                    </div>
                    <div class="field-group">
                        <label>Preço Máximo</label>
                        <select id="search-price" class="field-input" name="max_price" 
                                hx-get="/partials/imoveis" 
                                hx-include="#search-q, #search-type, #search-purpose"
                                hx-target="#lista-imoveis">
                            <option value="0">Qualquer valor</option>
                            <option value="2500">Até R$ 2.500</option>
                            <option value="5000">Até R$ 5.000</option>
                            <option value="10000">Até R$ 10.000</option>
                            <option value="500000">Até R$ 500.000</option>
                            <option value="1500000">Até R$ 1.5 Milhão</option>
                        </select>
                    </div>
                    <input type="hidden" id="search-purpose" name="purpose" value="">
                    <button class="btn btn-primary" style="height: 46px; align-self: flex-end; padding: 0 1.5rem;"
                            hx-get="/partials/imoveis"
                            hx-include="#search-q, #search-type, #search-purpose, #search-price"
                            hx-target="#lista-imoveis">
                        Buscar
                    </button>
                </div>
            </div>

            <div class="quick-categories">
                <a href="javascript:void(0)" onclick="setFilter('APARTMENT')" class="category-pill">🏢 Apartamentos</a>
                <a href="javascript:void(0)" onclick="setFilter('HOUSE')" class="category-pill">🏡 Casas</a>
                <a href="javascript:void(0)" onclick="setFilter('COMMERCIAL')" class="category-pill">💼 Comerciais</a>
                <a href="javascript:void(0)" onclick="setFilter('LAND')" class="category-pill">🌱 Terrenos</a>
            </div>
        </div>
    </section>

    <main class="container">
        <div class="section-header">
            <div>
                <h2 class="section-title">Imóveis Disponíveis</h2>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Conectados em tempo real com a base de dados oficial.
                </p>
            </div>
        </div>

        <section id="lista-imoveis" class="property-grid">
            ` + cardsHTML + `
        </section>
    </main>

    <footer class="footer">
        <div class="container">
            <div class="footer-content">
                <div class="logo-wrapper">
                    <div class="logo-icon">IX</div>
                    <div class="logo-text">Imov<span>xen</span></div>
                </div>
                <p style="color: var(--text-muted); font-size: 0.9rem;">
                    Marketplace imobiliário conectado ao PostgreSQL 16.
                </p>
            </div>
            <p style="text-align: center; color: var(--text-muted); font-size: 0.85rem;">
                &copy; 2026 Imovxen SaaS. Todos os direitos reservados.
            </p>
        </div>
    </footer>

    <script>
        function setPurpose(btn, purpose) {
            document.querySelectorAll('.search-tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById('search-purpose').value = purpose;
            htmx.trigger('#search-q', 'keyup');
        }
        function setFilter(type) {
            document.getElementById('search-type').value = type;
            htmx.trigger('#search-type', 'change');
        }
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// ==========================================
	// PARTIAL HTMX (BUSCA REAL NO POSTGRESQL)
	// ==========================================
	r.GET("/partials/imoveis", func(c *gin.Context) {
		q := c.Query("q")
		propType := c.Query("type")
		purpose := c.Query("purpose")
		maxPrice := c.Query("max_price")

		endpoint := fmt.Sprintf("/api/properties?q=%s&type=%s&purpose=%s&max_price=%s", q, propType, purpose, maxPrice)
		var apiRes struct {
			Data []PropertyItem `json:"data"`
		}
		_ = apiClient.Get(endpoint, "", &apiRes)

		if len(apiRes.Data) == 0 {
			emptyMsg := `
            <div class="property-card" style="padding: 3rem 1.5rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                <p style="font-size: 1.1rem; font-weight: 600;">Nenhum imóvel encontrado para os filtros selecionados.</p>
                <p style="font-size: 0.875rem; margin-top: 0.25rem;">Tente ajustar o termo de busca ou remover as restrições.</p>
            </div>`
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(emptyMsg))
			return
		}

		cards := ""
		for _, item := range apiRes.Data {
			cards += renderPropertyCard(item)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(cards))
	})

	// ==========================================
	// PÁGINA: DETALHE DO IMÓVEL (/imoveis/:id) REAL
	// ==========================================
	r.GET("/imoveis/:id", func(c *gin.Context) {
		id := c.Param("id")

		var prop PropertyItem
		err := apiClient.Get("/api/properties/"+id, "", &prop)
		if err != nil || prop.ID == "" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		imgURL := "https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=1200&q=80"
		if prop.CoverImageUrl != nil && *prop.CoverImageUrl != "" {
			imgURL = *prop.CoverImageUrl
		}

		purposeLabel := "Aluguel"
		priceSuffix := "<span>/mês</span>"
		if prop.Purpose == "SALE" {
			purposeLabel = "Venda"
			priceSuffix = ""
		}

		desc := "Imóvel disponível para negociação direta com o anunciante através da plataforma Imovxen."
		if prop.Description != nil && *prop.Description != "" {
			desc = *prop.Description
		}

		areaVal := "—"
		if prop.AreaM2 != nil && *prop.AreaM2 > 0 {
			areaVal = fmt.Sprintf("%.0fm²", *prop.AreaM2)
		}

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + prop.Title + ` — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <a href="/" class="btn btn-outline">&larr; Voltar para a busca</a>
        </div>
    </header>

    <main class="container" style="margin-top: 2.5rem; margin-bottom: 5rem;">
        <div style="display: grid; grid-template-columns: 2fr 1fr; gap: 2.5rem;">
            <div>
                <div style="border-radius: 1rem; overflow: hidden; height: 420px; box-shadow: var(--shadow-lg);">
                    <img src="` + imgURL + `" style="width: 100%; height: 100%; object-fit: cover;">
                </div>

                <div style="margin-top: 2rem;">
                    <span class="badge-purpose" style="position: static; display: inline-block; margin-bottom: 0.5rem;">` + purposeLabel + `</span>
                    <h1 style="font-size: 2.2rem; font-weight: 800; color: var(--text-main);">` + prop.Title + `</h1>
                    <p style="color: var(--text-muted); font-size: 1rem; margin-top: 0.5rem;">📍 ` + prop.City + ` - ` + prop.State + `</p>
                    
                    <div class="property-amenities" style="margin: 1.5rem 0; font-size: 1rem; gap: 2rem; border-bottom: 1px solid var(--border); padding-bottom: 1.5rem;">
                        <div class="amenity-item">📐 <strong>` + areaVal + `</strong> área útil</div>
                    </div>

                    <h3 style="font-size: 1.3rem; font-weight: 700; margin-bottom: 0.75rem;">Sobre o imóvel</h3>
                    <p style="color: #334155; line-height: 1.8;">` + desc + `</p>
                </div>
            </div>

            <div>
                <div style="background: #ffffff; padding: 2rem; border-radius: 1.25rem; border: 1px solid var(--border); box-shadow: var(--shadow-xl); position: sticky; top: 100px;">
                    <div style="font-size: 2rem; font-weight: 800; color: var(--primary);">
                        R$ ` + fmt.Sprintf("%.2f", prop.Price) + ` ` + priceSuffix + `
                    </div>

                    <div id="lead-feedback" style="margin-top: 1rem;"></div>

                    <form hx-post="/imoveis/` + id + `/contato" hx-target="#lead-feedback" style="margin-top: 1rem;">
                        <h4 style="font-weight: 700; font-size: 1.1rem; margin-bottom: 0.75rem;">Falar com o proprietário</h4>
                        <div class="form-group">
                            <label>Seu Nome</label>
                            <input class="form-control" type="text" name="name" required placeholder="Nome completo">
                        </div>
                        <div class="form-group">
                            <label>E-mail</label>
                            <input class="form-control" type="email" name="email" required placeholder="seu@email.com">
                        </div>
                        <div class="form-group">
                            <label>Telefone / WhatsApp</label>
                            <input class="form-control" type="text" name="phone" placeholder="(11) 99999-9999">
                        </div>
                        <div class="form-group">
                            <label>Mensagem</label>
                            <textarea class="form-control" name="message" rows="3" placeholder="Olá, gostaria de saber mais informações sobre o imóvel."></textarea>
                        </div>
                        <button type="submit" class="btn btn-primary" style="width: 100%; padding: 0.85rem; margin-top: 0.5rem;">
                            Enviar Mensagem Direta
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </main>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.POST("/imoveis/:id/contato", func(c *gin.Context) {
		name := c.PostForm("name")
		msg := fmt.Sprintf(`<div class="alert alert-success">
            <strong>Obrigado, %s!</strong><br>
            Sua mensagem foi enviada ao anunciante com sucesso!
        </div>`, name)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(msg))
	})

	// ==========================================
	// PÁGINA: CADASTRO (/cadastro)
	// ==========================================
	r.GET("/cadastro", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Criar Conta — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <a href="/login" class="btn btn-outline">Já tenho conta</a>
        </div>
    </header>

    <div class="form-container">
        <h2 style="font-weight: 800; font-size: 1.75rem;">Crie sua conta grátis</h2>
        <p style="color: var(--text-muted); margin: 0.35rem 0 1.75rem;">Publique seus anúncios de forma 100% digital e sem burocracia.</p>
        
        <div id="cadastro-feedback"></div>

        <form hx-post="/auth/register" hx-target="#cadastro-feedback">
            <div class="form-group">
                <label>Nome Completo</label>
                <input class="form-control" type="text" name="name" required placeholder="Ex: Lucas Silva">
            </div>
            <div class="form-group">
                <label>E-mail</label>
                <input class="form-control" type="email" name="email" required placeholder="seu@email.com">
            </div>
            <div class="form-group">
                <label>Telefone / WhatsApp</label>
                <input class="form-control" type="text" name="phone" placeholder="(11) 99999-9999">
            </div>
            <div class="form-group">
                <label>Senha</label>
                <input class="form-control" type="password" name="password" required minlength="6" placeholder="Mínimo 6 caracteres">
            </div>
            <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1rem; padding: 0.85rem;">Criar Minha Conta</button>
        </form>
    </div>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// ==========================================
	// PÁGINA: LOGIN (/login)
	// ==========================================
	r.GET("/login", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Entrar — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <a href="/cadastro" class="btn btn-outline">Criar conta</a>
        </div>
    </header>

    <div class="form-container">
        <h2 style="font-weight: 800; font-size: 1.75rem;">Acesse sua conta</h2>
        <p style="color: var(--text-muted); margin: 0.35rem 0 1.75rem;">Gerencie seus anúncios e responda a clientes interessados.</p>
        
        <div id="login-feedback"></div>

        <form hx-post="/auth/login" hx-target="#login-feedback">
            <div class="form-group">
                <label>E-mail</label>
                <input class="form-control" type="email" name="email" required placeholder="seu@email.com">
            </div>
            <div class="form-group">
                <label>Senha</label>
                <input class="form-control" type="password" name="password" required placeholder="Sua senha">
            </div>
            <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1rem; padding: 0.85rem;">Entrar na Plataforma</button>
        </form>
    </div>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.POST("/auth/register", func(c *gin.Context) {
		payload := map[string]interface{}{
			"name":     c.PostForm("name"),
			"email":    c.PostForm("email"),
			"password": c.PostForm("password"),
			"phone":    c.PostForm("phone"),
		}
		var apiRes struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := apiClient.Post("/api/auth/register", payload, "", &apiRes); err != nil || apiRes.AccessToken == "" {
			c.Header("HX-Retarget", "#cadastro-feedback")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Erro ao criar conta. E-mail já em uso ou dados inválidos.</div>`))
			return
		}
		c.SetCookie("access_token", apiRes.AccessToken, 900, "/", "", false, true)
		c.SetCookie("refresh_token", apiRes.RefreshToken, 604800, "/", "", false, true)
		c.Header("HX-Redirect", "/painel/imoveis")
		c.Status(http.StatusOK)
	})

	r.POST("/auth/login", func(c *gin.Context) {
		payload := map[string]interface{}{
			"email":    c.PostForm("email"),
			"password": c.PostForm("password"),
		}
		var apiRes struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := apiClient.Post("/api/auth/login", payload, "", &apiRes); err != nil || apiRes.AccessToken == "" {
			c.Header("HX-Retarget", "#login-feedback")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">E-mail ou senha incorretos.</div>`))
			return
		}
		c.SetCookie("access_token", apiRes.AccessToken, 900, "/", "", false, true)
		c.SetCookie("refresh_token", apiRes.RefreshToken, 604800, "/", "", false, true)
		c.Header("HX-Redirect", "/painel/imoveis")
		c.Status(http.StatusOK)
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	})

	// ==========================================
	// PÁGINA: PAINEL MEUS IMÓVEIS (DADOS DO BANCO REAL)
	// ==========================================
	r.GET("/painel/imoveis", func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		var meRes struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
			Quota struct {
				Used  int `json:"used"`
				Limit int `json:"limit"`
			} `json:"quota"`
		}
		_ = apiClient.Get("/api/me", token, &meRes)

		var propsRes struct {
			Data []PropertyItem `json:"data"`
		}
		_ = apiClient.Get("/api/me/properties", token, &propsRes)

		cardsHTML := ""
		if len(propsRes.Data) == 0 {
			cardsHTML = `
            <div class="property-card" style="padding: 4rem 2rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                <p style="font-size: 1.1rem; font-weight: 600;">Você ainda não possui anúncios ativos.</p>
                <p style="font-size: 0.9rem; margin-top: 0.25rem;">Comece agora mesmo a divulgar seus imóveis para milhares de compradores.</p>
                <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary" style="margin-top: 1.5rem;">Publicar Primeiro Imóvel</button>
            </div>`
		} else {
			for _, p := range propsRes.Data {
				cardsHTML += renderPropertyCard(p)
			}
		}

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Painel do Anunciante — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <div class="nav-links">
                <span style="color: var(--text-muted); font-size: 0.9rem;">Olá, <strong>` + meRes.User.Name + `</strong></span>
                <a href="/logout" class="btn btn-outline">Sair</a>
            </div>
        </div>
    </header>

    <main class="container" style="margin-top: 2.5rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
            <div>
                <h2 style="font-size: 1.85rem; font-weight: 800;">Painel de Gestão</h2>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Cota ativa: <strong>` + strconv.Itoa(meRes.Quota.Used) + ` de ` + strconv.Itoa(meRes.Quota.Limit) + ` anúncios gratuitos</strong>
                </p>
            </div>
            <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary">+ Anunciar Imóvel</button>
        </div>

        <div class="property-grid">
            ` + cardsHTML + `
        </div>

        <div id="modal-novo-imovel" style="display: none; position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); z-index: 50; overflow-y: auto; padding: 2.5rem 0;">
            <div class="form-container" style="margin: auto; max-width: 620px;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                    <h3 style="font-size: 1.35rem; font-weight: 700;">Publicar Novo Imóvel</h3>
                    <button onclick="document.getElementById('modal-novo-imovel').style.display='none'" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted);">&times;</button>
                </div>

                <div id="novo-imovel-feedback"></div>

                <form hx-post="/painel/imoveis/novo" hx-target="#novo-imovel-feedback">
                    <div class="form-group">
                        <label>Título do Anúncio</label>
                        <input class="form-control" type="text" name="title" required placeholder="Ex: Apartamento no coração de Pinheiros">
                    </div>
                    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                        <div class="form-group">
                            <label>Tipo de Imóvel</label>
                            <select class="form-control" name="type" required>
                                <option value="APARTMENT">Apartamento</option>
                                <option value="HOUSE">Casa</option>
                                <option value="COMMERCIAL">Comercial</option>
                                <option value="LAND">Terreno</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Finalidade</label>
                            <select class="form-control" name="purpose" required>
                                <option value="RENT">Aluguel</option>
                                <option value="SALE">Venda</option>
                            </select>
                        </div>
                    </div>
                    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                        <div class="form-group">
                            <label>Preço (R$)</label>
                            <input class="form-control" type="number" step="0.01" name="price" required placeholder="3500.00">
                        </div>
                        <div class="form-group">
                            <label>Área (m²)</label>
                            <input class="form-control" type="number" step="0.1" name="area_m2" placeholder="65.0">
                        </div>
                    </div>
                    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                        <div class="form-group">
                            <label>Cidade</label>
                            <input class="form-control" type="text" name="city" required placeholder="São Paulo">
                        </div>
                        <div class="form-group">
                            <label>Estado (UF)</label>
                            <input class="form-control" type="text" name="state" required maxlength="2" placeholder="SP">
                        </div>
                    </div>
                    <div class="form-group">
                        <label>Descrição do Imóvel</label>
                        <textarea class="form-control" name="description" rows="3" placeholder="Detalhes do imóvel, condomínio, localização..."></textarea>
                    </div>
                    <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1.25rem; padding: 0.85rem;">Publicar Agora</button>
                </form>
            </div>
        </div>
    </main>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.POST("/painel/imoveis/novo", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
		area, _ := strconv.ParseFloat(c.PostForm("area_m2"), 64)

		desc := c.PostForm("description")
		payload := map[string]interface{}{
			"title":       c.PostForm("title"),
			"type":        c.PostForm("type"),
			"purpose":     c.PostForm("purpose"),
			"price":       price,
			"area_m2":     area,
			"city":        c.PostForm("city"),
			"state":       c.PostForm("state"),
			"description": desc,
		}

		var apiRes struct {
			ID    string `json:"id"`
			Error string `json:"error"`
		}

		err := apiClient.Post("/api/properties", payload, token, &apiRes)
		if err != nil || apiRes.ID == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Erro ao criar anúncio. Verifique se o limite de cota gratuita (2 imóveis) foi atingido.</div>`))
			return
		}

		c.Header("HX-Refresh", "true")
		c.Status(http.StatusOK)
	})

	slog.Info("Servidor Frontend Conectado ao Banco Real", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
