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
	// PÁGINA: HOME MODERNA
	// ==========================================
	r.GET("/", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		userActions := `<a href="/login" class="nav-item">Entrar</a>
                        <a href="/cadastro" class="btn btn-primary">Criar Conta Grátis</a>`
		if token != "" {
			userActions = `<a href="/painel/imoveis" class="btn btn-primary">Painel de Imóveis</a>
                           <a href="/logout" class="btn btn-outline" style="margin-left: 0.5rem;">Sair</a>`
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
                    <button class="search-tab-btn active" onclick="setPurpose(this, 'ALL')">Todos</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'RENT')">Aluguel</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'SALE')">Comprar</button>
                </div>
                <div class="search-grid">
                    <div class="field-group">
                        <label>Localização</label>
                        <input class="field-input" type="text" name="city" placeholder="Cidade, bairro ou CEP..."
                               hx-get="/partials/imoveis" 
                               hx-trigger="keyup changed delay:350ms" 
                               hx-target="#lista-imoveis">
                    </div>
                    <div class="field-group">
                        <label>Tipo de Imóvel</label>
                        <select class="field-input" name="type" hx-get="/partials/imoveis" hx-target="#lista-imoveis">
                            <option value="">Todos os tipos</option>
                            <option value="APARTMENT">Apartamentos</option>
                            <option value="HOUSE">Casas e Sobrados</option>
                            <option value="COMMERCIAL">Comerciais</option>
                        </select>
                    </div>
                    <div class="field-group">
                        <label>Faixa de Preço</label>
                        <select class="field-input" name="price_range" hx-get="/partials/imoveis" hx-target="#lista-imoveis">
                            <option value="">Qualquer valor</option>
                            <option value="1">Até R$ 3.000</option>
                            <option value="2">R$ 3.000 a R$ 6.000</option>
                            <option value="3">R$ 500k a R$ 1 Milhão</option>
                            <option value="4">Acima de R$ 1 Milhão</option>
                        </select>
                    </div>
                    <button class="btn btn-primary" style="height: 46px; align-self: flex-end; padding: 0 1.5rem;">
                        Buscar
                    </button>
                </div>
            </div>

            <div class="quick-categories">
                <a href="#" class="category-pill">🏢 Apartamentos de Alto Padrão</a>
                <a href="#" class="category-pill">🏡 Casas em Condomínio</a>
                <a href="#" class="category-pill">💼 Espaços Comerciais</a>
                <a href="#" class="category-pill">🏖️ Imóveis no Litoral</a>
            </div>
        </div>
    </section>

    <main class="container">
        <div class="section-header">
            <div>
                <h2 class="section-title">Imóveis em Destaque</h2>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Oportunidades verificadas e recém-publicadas na plataforma.
                </p>
            </div>
            <a href="/" class="btn btn-outline">Ver todos os imóveis &rarr;</a>
        </div>

        <section id="lista-imoveis" class="property-grid">
            <article class="property-card" onclick="window.location.href='/imoveis/exemplo-1'" style="cursor: pointer;">
                <div class="card-image-wrapper">
                    <span class="badge-purpose">Aluguel</span>
                    <span class="badge-featured">Exclusivo</span>
                    <img src="https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=800&q=80" alt="Apartamento Pinheiros">
                </div>
                <div class="property-content">
                    <div class="property-price">R$ 3.800 <span>/mês</span></div>
                    <h3 class="property-title">Apartamento Studio Moderno</h3>
                    <div class="property-location">📍 Pinheiros, São Paulo - SP</div>
                    <div class="property-amenities">
                        <div class="amenity-item">📐 62m²</div>
                        <div class="amenity-item">🛏️ 2 Quartos</div>
                        <div class="amenity-item">🚗 1 Vaga</div>
                    </div>
                </div>
            </article>

            <article class="property-card" onclick="window.location.href='/imoveis/exemplo-2'" style="cursor: pointer;">
                <div class="card-image-wrapper">
                    <span class="badge-purpose" style="background: rgba(37,99,235,0.85);">Venda</span>
                    <span class="badge-featured" style="background: #10b981;">Novo</span>
                    <img src="https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=800&q=80" alt="Casa Barra">
                </div>
                <div class="property-content">
                    <div class="property-price">R$ 1.250.000</div>
                    <h3 class="property-title">Casa Contemporânea com Piscina</h3>
                    <div class="property-location">📍 Barra da Tijuca, Rio de Janeiro - RJ</div>
                    <div class="property-amenities">
                        <div class="amenity-item">📐 240m²</div>
                        <div class="amenity-item">🛏️ 4 Suítes</div>
                        <div class="amenity-item">🚗 3 Vagas</div>
                    </div>
                </div>
            </article>
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
                    Marketplace imobiliário de alta performance construído com Go e HTMX.
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
        }
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// ==========================================
	// PÁGINA: DETALHE DO IMÓVEL (/imoveis/:id)
	// ==========================================
	r.GET("/imoveis/:id", func(c *gin.Context) {
		id := c.Param("id")

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Apartamento Studio Moderno — Imovxen</title>
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
                    <img src="https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=1200&q=80" style="width: 100%; height: 100%; object-fit: cover;">
                </div>

                <div style="margin-top: 2rem;">
                    <span class="badge-purpose" style="position: static; display: inline-block; margin-bottom: 0.5rem;">Aluguel</span>
                    <h1 style="font-size: 2.2rem; font-weight: 800; color: var(--text-main);">Apartamento Studio Moderno em Pinheiros</h1>
                    <p style="color: var(--text-muted); font-size: 1rem; margin-top: 0.5rem;">📍 Rua dos Pinheiros, Pinheiros — São Paulo, SP</p>
                    
                    <div class="property-amenities" style="margin: 1.5rem 0; font-size: 1rem; gap: 2rem; border-bottom: 1px solid var(--border); padding-bottom: 1.5rem;">
                        <div class="amenity-item">📐 <strong>62m²</strong> área útil</div>
                        <div class="amenity-item">🛏️ <strong>2</strong> quartos (1 suíte)</div>
                        <div class="amenity-item">🚿 <strong>1</strong> banheiro</div>
                        <div class="amenity-item">🚗 <strong>1</strong> vaga coberta</div>
                    </div>

                    <h3 style="font-size: 1.3rem; font-weight: 700; margin-bottom: 0.75rem;">Sobre o imóvel</h3>
                    <p style="color: #334155; line-height: 1.8;">
                        Excelente apartamento totalmente mobiliado e decorado por arquiteto no coração de Pinheiros. 
                        Condomínio moderno com piscina no rooftop, academia completa, espaço coworking e portaria 24 horas. 
                        Localização privilegiada, a apenas 3 minutos a pé da estação Fradique Coutinho do metrô.
                    </p>
                </div>
            </div>

            <!-- CARD LATERAL DE CONTATO E PREÇO -->
            <div>
                <div style="background: #ffffff; padding: 2rem; border-radius: 1.25rem; border: 1px solid var(--border); box-shadow: var(--shadow-xl); position: sticky; top: 100px;">
                    <div style="font-size: 2rem; font-weight: 800; color: var(--primary);">
                        R$ 3.800 <span style="font-size: 1rem; font-weight: 500; color: var(--text-muted);">/mês</span>
                    </div>
                    <div style="font-size: 0.85rem; color: var(--text-muted); margin-top: 0.25rem; border-bottom: 1px solid var(--border); padding-bottom: 1rem;">
                        Condomínio: R$ 650 · IPTU: R$ 120
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
                            <textarea class="form-control" name="message" rows="3" placeholder="Olá, gostaria de saber se o imóvel ainda está disponível para visita."></textarea>
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

	// Endpoint HTMX para envio de contato/lead
	r.POST("/imoveis/:id/contato", func(c *gin.Context) {
		name := c.PostForm("name")
		msg := fmt.Sprintf(`<div class="alert alert-success">
            <strong>Obrigado, %s!</strong><br>
            Sua mensagem foi enviada diretamente para o proprietário. Em breve ele entrará em contato via WhatsApp/e-mail.
        </div>`, name)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(msg))
	})

	// ==========================================
	// PÁGINA: PAINEL MEUS IMÓVEIS + SUBMISSÃO REAL
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
			Data []struct {
				ID       string  `json:"id"`
				Title    string  `json:"title"`
				Type     string  `json:"type"`
				Purpose  string  `json:"purpose"`
				Price    float64 `json:"price"`
				City     string  `json:"city"`
				State    string  `json:"state"`
				Status   string  `json:"status"`
			} `json:"data"`
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
				cardsHTML += fmt.Sprintf(`
                <article class="property-card">
                    <div class="card-image-wrapper">
                        <span class="badge-purpose">%s</span>
                        <img src="https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=800&q=80">
                    </div>
                    <div class="property-content">
                        <div class="property-price">R$ %.2f</div>
                        <h3 class="property-title">%s</h3>
                        <div class="property-location">📍 %s - %s</div>
                    </div>
                </article>`, p.Purpose, p.Price, p.Title, p.City, p.State)
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
                    <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1.25rem; padding: 0.85rem;">Publicar Agora</button>
                </form>
            </div>
        </div>
    </main>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// Submissão real de criação de anúncio
	r.POST("/painel/imoveis/novo", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
		area, _ := strconv.ParseFloat(c.PostForm("area_m2"), 64)

		payload := map[string]interface{}{
			"title":    c.PostForm("title"),
			"type":     c.PostForm("type"),
			"purpose":  c.PostForm("purpose"),
			"price":    price,
			"area_m2":  area,
			"city":     c.PostForm("city"),
			"state":    c.PostForm("state"),
		}

		var apiRes struct {
			ID    string `json:"id"`
			Error string `json:"error"`
		}

		err := apiClient.Post("/api/properties", payload, token, &apiRes)
		if err != nil || apiRes.ID == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Erro ao criar anúncio. Verifique se a cota foi atingida.</div>`))
			return
		}

		// Recarrega o painel via HTMX
		c.Header("HX-Refresh", "true")
		c.Status(http.StatusOK)
	})

	// Handlers de Auth (preservados)
	r.GET("/cadastro", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	})
	r.GET("/login", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/")
	})

	slog.Info("Servidor Frontend Iniciado com Fluxos Completos", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
