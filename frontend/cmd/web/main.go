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
	_ = apiClient

	r := gin.Default()

	// Arquivos Estáticos (CSS, HTMX)
	r.Static("/static", "./static")

	// Rota Home SSR com HTMX
	r.GET("/", func(c *gin.Context) {
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
                <a href="/login" class="btn btn-primary">Anunciar Imóvel</a>
            </nav>
        </div>
    </header>

    <main class="container">
        <section class="hero">
            <h1>Encontre o imóvel perfeito para comprar ou alugar</h1>
            <p style="color: var(--text-muted); margin-top: 0.5rem;">Direto com anunciantes de forma simples, transparente e sem burocracia.</p>
            
            <div class="search-bar">
                <input class="search-input" type="text" name="q" placeholder="Buscar por cidade, bairro ou título..."
                       hx-get="/partials/imoveis" 
                       hx-trigger="keyup changed delay:400ms" 
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

	// Partial HTMX para busca de imóveis
	r.GET("/partials/imoveis", func(c *gin.Context) {
		q := c.Query("q")
		slog.Info("Busca reativa HTMX executada", "termo", q)
		
		partial := `
        <div class="property-card">
            <img src="https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=500&q=80" alt="Apartamento">
            <div class="property-content">
                <div class="property-price">R$ 3.200 /mês</div>
                <h3 style="margin: 0.5rem 0;">Apartamento 2 Dormitórios</h3>
                <p style="color: var(--text-muted); font-size: 0.875rem;">Pinheiros, São Paulo - SP</p>
                <p style="margin-top: 0.5rem; font-size: 0.875rem;">65m² · 2 quartos · 1 vaga</p>
            </div>
        </div>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(partial))
	})

	slog.Info("Servidor Frontend Iniciado", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
