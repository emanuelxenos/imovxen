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
	// PÁGINA: HOME COM BUSCA HTMX
	// ==========================================
	r.GET("/", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		userBtn := `<a href="/login" class="btn btn-outline">Entrar</a>
                    <a href="/cadastro" class="btn btn-primary" style="margin-left: 0.5rem;">Cadastre-se</a>`
		if token != "" {
			userBtn = `<a href="/painel/imoveis" class="btn btn-primary">Meus Imóveis</a>
                       <a href="/logout" class="btn btn-outline" style="margin-left: 0.5rem;">Sair</a>`
		}

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Imovxen — Marketplace Imobiliário</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo">Imovxen</a>
            <nav class="nav-links">
                <a href="/">Explorar</a>
                ` + userBtn + `
            </nav>
        </div>
    </header>

    <main class="container">
        <section class="hero">
            <h1>Encontre o imóvel perfeito para comprar ou alugar</h1>
            <p style="color: var(--text-muted); margin-top: 0.5rem;">Direto com anunciantes de forma simples, transparente e sem intermediários.</p>
            
            <div class="search-bar">
                <input class="search-input" type="text" name="q" placeholder="Buscar por cidade, bairro ou título..."
                       hx-get="/partials/imoveis" 
                       hx-trigger="keyup changed delay:350ms" 
                       hx-target="#lista-imoveis">
                <button class="btn btn-primary">Buscar</button>
            </div>
        </section>

        <section id="lista-imoveis" class="property-grid">
            <div class="property-card">
                <img src="https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=500&q=80" alt="Apartamento">
                <div class="property-content">
                    <div class="property-price">R$ 3.200 /mês</div>
                    <h3 style="margin: 0.5rem 0;">Apartamento 2 Dormitórios</h3>
                    <p style="color: var(--text-muted); font-size: 0.875rem;">Pinheiros, São Paulo - SP</p>
                    <p style="margin-top: 0.5rem; font-size: 0.875rem;">65m² · 2 quartos · 1 vaga</p>
                </div>
            </div>

            <div class="property-card">
                <img src="https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=500&q=80" alt="Casa">
                <div class="property-content">
                    <div class="property-price">R$ 850.000</div>
                    <h3 style="margin: 0.5rem 0;">Casa em Condomínio Fechado</h3>
                    <p style="color: var(--text-muted); font-size: 0.875rem;">Barra da Tijuca, Rio de Janeiro - RJ</p>
                    <p style="margin-top: 0.5rem; font-size: 0.875rem;">180m² · 3 suítes · 2 vagas</p>
                </div>
            </div>
        </section>
    </main>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// ==========================================
	// PÁGINA: CADASTRO
	// ==========================================
	r.GET("/cadastro", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Cadastro — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo">Imovxen</a>
            <a href="/login" class="btn btn-outline">Já tenho conta</a>
        </div>
    </header>

    <div class="form-container">
        <h2>Crie sua conta grátis</h2>
        <p style="color: var(--text-muted); margin-bottom: 1.5rem;">Publique até 2 imóveis gratuitamente no plano Freemium.</p>
        
        <div id="cadastro-feedback"></div>

        <form hx-post="/auth/register" hx-target="#cadastro-feedback">
            <div class="form-group">
                <label>Nome Completo</label>
                <input class="form-control" type="text" name="name" required placeholder="Seu nome">
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
            <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1rem;">Criar Conta</button>
        </form>
    </div>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// Handler de formulário: Registro
	r.POST("/auth/register", func(c *gin.Context) {
		name := c.PostForm("name")
		email := c.PostForm("email")
		password := c.PostForm("password")
		phone := c.PostForm("phone")

		payload := map[string]interface{}{
			"name":     name,
			"email":    email,
			"password": password,
			"phone":    phone,
		}

		var apiRes struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Error        string `json:"error"`
		}

		err := apiClient.Post("/api/auth/register", payload, "", &apiRes)
		if err != nil || apiRes.AccessToken == "" {
			c.Header("HX-Retarget", "#cadastro-feedback")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Erro ao criar conta. E-mail já em uso ou dados inválidos.</div>`))
			return
		}

		// Armazena cookie de sessão seguro
		c.SetCookie("access_token", apiRes.AccessToken, 900, "/", "", false, true)
		c.SetCookie("refresh_token", apiRes.RefreshToken, 604800, "/", "", false, true)

		// Redireciona via HTMX para o painel de imóveis
		c.Header("HX-Redirect", "/painel/imoveis")
		c.Status(http.StatusOK)
	})

	// ==========================================
	// PÁGINA: LOGIN
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
            <a href="/" class="logo">Imovxen</a>
            <a href="/cadastro" class="btn btn-outline">Criar conta</a>
        </div>
    </header>

    <div class="form-container">
        <h2>Entrar na sua conta</h2>
        <p style="color: var(--text-muted); margin-bottom: 1.5rem;">Gerencie seus anúncios e visualize os contatos recebidos.</p>
        
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
            <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1rem;">Entrar</button>
        </form>
    </div>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// Handler de formulário: Login
	r.POST("/auth/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}

		var apiRes struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}

		err := apiClient.Post("/api/auth/login", payload, "", &apiRes)
		if err != nil || apiRes.AccessToken == "" {
			c.Header("HX-Retarget", "#login-feedback")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">E-mail ou senha incorretos.</div>`))
			return
		}

		c.SetCookie("access_token", apiRes.AccessToken, 900, "/", "", false, true)
		c.SetCookie("refresh_token", apiRes.RefreshToken, 604800, "/", "", false, true)

		c.Header("HX-Redirect", "/painel/imoveis")
		c.Status(http.StatusOK)
	})

	// Logout
	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	})

	// ==========================================
	// PÁGINA: PAINEL MEUS IMÓVEIS (PROTEGIDA)
	// ==========================================
	r.GET("/painel/imoveis", func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		var meRes struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
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
    <title>Meus Imóveis — Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <header class="navbar">
        <div class="container nav-content">
            <a href="/" class="logo">Imovxen</a>
            <div class="nav-links">
                <span style="color: var(--text-muted); font-size: 0.875rem; margin-right: 1rem;">Olá, ` + meRes.User.Name + `</span>
                <a href="/logout" class="btn btn-outline">Sair</a>
            </div>
        </div>
    </header>

    <main class="container" style="margin-top: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
            <div>
                <h2>Painel do Anunciante</h2>
                <p style="color: var(--text-muted); font-size: 0.875rem;">Cota de anúncios: <strong>` + string(rune('0'+meRes.Quota.Used)) + ` de ` + string(rune('0'+meRes.Quota.Limit)) + `</strong> usados</p>
            </div>
            <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary">+ Novo Anúncio</button>
        </div>

        <div class="property-grid">
            <div class="property-card" style="padding: 2rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1;">
                <p>Nenhum anúncio publicado ainda.</p>
                <button onclick="document.getElementById('modal-novo-imovel').style.display='block'" class="btn btn-primary" style="margin-top: 1rem;">Publicar seu primeiro imóvel</button>
            </div>
        </div>

        <!-- MODAL NOVO ANÚNCIO -->
        <div id="modal-novo-imovel" style="display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 50; overflow-y: auto; padding: 2rem 0;">
            <div class="form-container" style="margin: auto; max-width: 650px;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                    <h3>Cadastrar Novo Imóvel</h3>
                    <button onclick="document.getElementById('modal-novo-imovel').style.display='none'" style="background: none; border: none; font-size: 1.5rem; cursor: pointer;">&times;</button>
                </div>

                <div id="novo-imovel-feedback"></div>

                <form hx-post="/painel/imoveis/novo" hx-target="#novo-imovel-feedback">
                    <div class="form-group">
                        <label>Título do Anúncio</label>
                        <input class="form-control" type="text" name="title" required placeholder="Ex: Apartamento reformado em Pinheiros">
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
                            <input class="form-control" type="number" step="0.01" name="price" required placeholder="3200.00">
                        </div>
                        <div class="form-group">
                            <label>Área Útil (m²)</label>
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
                    <button type="submit" class="btn btn-primary" style="width: 100%; margin-top: 1rem;">Publicar Anúncio</button>
                </form>
            </div>
        </div>
    </main>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	slog.Info("Servidor Frontend Iniciado", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
