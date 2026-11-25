package main

import (
	"log/slog"
	"net/http"
	"os"

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
	// PÁGINA: HOME MODERNA E IMPACTANTE
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
    <!-- NAVBAR ELEGANTE -->
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo-wrapper">
                <div class="logo-icon">IX</div>
                <div class="logo-text">Imov<span>xen</span></div>
            </a>
            <nav class="nav-links">
                <a href="/" class="nav-item">Comprar</a>
                <a href="/" class="nav-item">Alugar</a>
                <a href="/painel/imoveis" class="nav-item">Anunciar</a>
                ` + userActions + `
            </nav>
        </div>
    </header>

    <!-- HERO SECTION MODERNA -->
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

            <!-- CARD DE BUSCA PREMIUM -->
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

            <!-- CATEGORIAS EM DESTAQUE -->
            <div class="quick-categories">
                <a href="#" class="category-pill">🏢 Apartamentos de Alto Padrão</a>
                <a href="#" class="category-pill">🏡 Casas em Condomínio</a>
                <a href="#" class="category-pill">💼 Espaços Comerciais</a>
                <a href="#" class="category-pill">🏖️ Imóveis no Litoral</a>
            </div>
        </div>
    </section>

    <!-- VITRINE DE IMÓVEIS -->
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
            <!-- CARD 1 -->
            <article class="property-card">
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

            <!-- CARD 2 -->
            <article class="property-card">
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

            <!-- CARD 3 -->
            <article class="property-card">
                <div class="card-image-wrapper">
                    <span class="badge-purpose">Aluguel</span>
                    <img src="https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=800&q=80" alt="Cobertura">
                </div>
                <div class="property-content">
                    <div class="property-price">R$ 7.500 <span>/mês</span></div>
                    <h3 class="property-title">Cobertura Duplex com Vista Panorâmica</h3>
                    <div class="property-location">📍 Batel, Curitiba - PR</div>
                    <div class="property-amenities">
                        <div class="amenity-item">📐 185m²</div>
                        <div class="amenity-item">🛏️ 3 Suítes</div>
                        <div class="amenity-item">🚗 2 Vagas</div>
                    </div>
                </div>
            </article>
        </section>
    </main>

    <!-- FOOTER MODERNO -->
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

	// Partial HTMX para busca dinâmica
	r.GET("/partials/imoveis", func(c *gin.Context) {
		partial := `
        <article class="property-card">
            <div class="card-image-wrapper">
                <span class="badge-purpose">Aluguel</span>
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

        <article class="property-card">
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
        </article>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(partial))
	})

	// Handlers de Auth e Painel continuam preservados...
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
                    Cota ativa: <strong>` + string(rune('0'+meRes.Quota.Used)) + ` de ` + string(rune('0'+meRes.Quota.Limit)) + ` anúncios gratuitos</strong>
                </p>
            </div>
            <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary">+ Anunciar Imóvel</button>
        </div>

        <div class="property-grid">
            <div class="property-card" style="padding: 4rem 2rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                <p style="font-size: 1.1rem; font-weight: 600;">Você ainda não possui anúncios ativos.</p>
                <p style="font-size: 0.9rem; margin-top: 0.25rem;">Comece agora mesmo a divulgar seus imóveis para milhares de compradores.</p>
                <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary" style="margin-top: 1.5rem;">Publicar Primeiro Imóvel</button>
            </div>
        </div>

        <div id="modal-novo-imovel" style="display: none; position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); z-index: 50; overflow-y: auto; padding: 2.5rem 0;">
            <div class="form-container" style="margin: auto; max-width: 620px;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                    <h3 style="font-size: 1.35rem; font-weight: 700;">Publicar Novo Imóvel</h3>
                    <button onclick="document.getElementById('modal-novo-imovel').style.display='none'" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted);">&times;</button>
                </div>
                <form hx-post="/painel/imoveis/novo">
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

	slog.Info("Servidor Frontend Iniciado com Novo Design", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
