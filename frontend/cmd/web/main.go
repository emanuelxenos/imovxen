package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

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

type LeadItem struct {
	ID            string  `json:"id"`
	PropertyTitle string  `json:"property_title"`
	Name          string  `json:"name"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	Message       *string `json:"message"`
	CreatedAt     string  `json:"created_at"`
}

type UserProfile struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Phone        *string `json:"phone"`
	CpfCnpj      *string `json:"cpf_cnpj"`
	ListingQuota int     `json:"listing_quota"`
	ListingsUsed int     `json:"listings_used"`
}

type PaymentRecord struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Amount      float64 `json:"amount"`
	ItemsCount  int     `json:"items_count"`
	DueDate     string  `json:"due_date"`
	Status      string  `json:"status"`
	Method      string  `json:"method"`
	GatewayID   *string `json:"gateway_id"`
	PixCode     *string `json:"pix_code"`
	PixQrBase64 *string `json:"pix_qr_base64"`
	PaidAt      *string `json:"paid_at"`
	CreatedAt   string  `json:"created_at"`
}

type PlanPackageItem struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Listings       int     `json:"listings"`
	Price          float64 `json:"price"`
	PriceFormatted string  `json:"price_formatted"`
}

type ContractItem struct {
	ID             string  `json:"id"`
	PropertyID     string  `json:"property_id"`
	PropertyTitle  string  `json:"property_title"`
	PropertyCity   string  `json:"property_city"`
	PropertyState  string  `json:"property_state"`
	TenantName     string  `json:"tenant_name"`
	TenantPhone    *string `json:"tenant_phone"`
	RentAmount     float64 `json:"rent_amount"`
	DueDay         int     `json:"due_day"`
	DurationMonths int     `json:"duration_months"`
	Status         string  `json:"status"`
}

type InstallmentItem struct {
	ID                string  `json:"id"`
	InstallmentNumber int     `json:"installment_number"`
	DueDate           string  `json:"due_date"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
	ManualFeeAmount   float64 `json:"manual_fee_amount"`
	PaidAt            *string `json:"paid_at"`
	PaymentMethod     *string `json:"payment_method"`
	ReceiptCode       *string `json:"receipt_code"`
}

func parsePhotos(coverURL *string) []string {
	defaultPhotos := []string{
		"https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80",
	}

	if coverURL == nil || *coverURL == "" {
		return defaultPhotos
	}

	urls := strings.Split(*coverURL, ",")
	var cleaned []string
	for _, u := range urls {
		trim := strings.TrimSpace(u)
		if trim != "" {
			cleaned = append(cleaned, trim)
		}
	}
	if len(cleaned) == 0 {
		return defaultPhotos
	}
	return cleaned
}

func renderPropertyCard(p PropertyItem, isOwner bool) string {
	photos := parsePhotos(p.CoverImageUrl)
	coverImg := photos[0]

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

	actionButtons := ""
	if isOwner {
		actionButtons = fmt.Sprintf(`
        <div style="display: flex; gap: 0.5rem; margin-top: 1rem; border-top: 1px solid var(--border); padding-top: 0.75rem;">
            <button onclick="event.stopPropagation(); abrirModalEdicao('%s', '%s', '%s', '%s', %.2f, '%.1f', '%s', '%s', '%s')" class="btn btn-outline" style="flex: 1; padding: 0.4rem 0.5rem; font-size: 0.85rem;">✏️ Editar</button>
            <button onclick="event.stopPropagation(); if(confirm('Tem certeza que deseja excluir este anúncio?')) htmx.ajax('DELETE', '/painel/imoveis/%s', {target: 'body'})" class="btn btn-outline" style="color: #dc2626; border-color: #fecaca; padding: 0.4rem 0.5rem; font-size: 0.85rem;">🗑️ Excluir</button>
        </div>`, p.ID, p.Title, p.Type, p.Purpose, p.Price, func() float64 {
			if p.AreaM2 != nil {
				return *p.AreaM2
			}
			return 0
		}(), p.City, p.State, func() string {
			if p.CoverImageUrl != nil {
				return *p.CoverImageUrl
			}
			return ""
		}(), p.ID)
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
            %s
        </div>
    </article>`, p.ID, purposeLabel, coverImg, p.Title, p.Price, priceSuffix, p.Title, p.City, p.State, area, bedrooms, parking, actionButtons)
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
	// PÁGINA: HOME
	// ==========================================
	r.GET("/", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		userActions := `<a href="/login" class="nav-item">Entrar</a>
                        <a href="/cadastro" class="btn btn-primary">Criar Conta Grátis</a>`
		if token != "" {
			userActions = `<a href="/painel" class="btn btn-primary">Meu Painel</a>
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
                <p style="margin-top: 0.5rem;">Cadastre-se e publique seu primeiro imóvel para aluguel ou venda!</p>
                <a href="/cadastro" class="btn btn-primary" style="margin-top: 1.5rem;">Publicar Anúncio Agora</a>
            </div>`
		} else {
			for _, item := range apiRes.Data {
				cardsHTML += renderPropertyCard(item, false)
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
                <a href="/painel" class="nav-item">Painel</a>
                ` + userActions + `
            </nav>
        </div>
    </header>

    <section class="hero-wrapper">
        <div class="container">
            <div class="hero-badge">
                <span>✨ Aluguel e Venda Direto com Proprietários</span>
            </div>
            <h1 class="hero-title">
                Encontre o imóvel perfeito para <span class="highlight">alugar ou comprar</span>
            </h1>
            <p class="hero-subtitle">
                Negocie diretamente, envie propostas em tempo real e faça tudo de forma 100% digital.
            </p>

            <div class="search-card">
                <div class="search-tabs">
                    <button class="search-tab-btn active" onclick="setPurpose(this, '')">Todos</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'RENT')">Aluguel</button>
                    <button class="search-tab-btn" onclick="setPurpose(this, 'SALE')">Comprar</button>
                </div>
                <div class="search-grid">
                    <div class="field-group">
                        <label>Localização <span id="geo-status" style="color: var(--primary); font-size: 0.75rem; text-transform: none; cursor: pointer;" onclick="detectarLocalizacao()">📍 Usar meu local</span></label>
                        <input id="search-q" class="field-input" type="text" name="q" placeholder="Cidade, bairro ou título..."
                               hx-get="/partials/imoveis" 
                               hx-include="#search-type, #search-purpose, #search-price, #search-lat, #search-lng"
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
                    <input type="hidden" id="search-lat" name="lat" value="">
                    <input type="hidden" id="search-lng" name="lng" value="">
                    <button class="btn btn-primary" style="height: 46px; align-self: flex-end; padding: 0 1.5rem;"
                            hx-get="/partials/imoveis"
                            hx-include="#search-q, #search-type, #search-purpose, #search-price, #search-lat, #search-lng"
                            hx-target="#lista-imoveis">
                        Buscar
                    </button>
                </div>
            </div>
        </div>
    </section>

    <main class="container">
        <div class="section-header">
            <div>
                <h2 class="section-title">Imóveis Disponíveis</h2>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Oportunidades atualizadas no PostgreSQL em tempo real.
                </p>
            </div>
        </div>

        <section id="lista-imoveis" class="property-grid">
            ` + cardsHTML + `
        </section>
    </main>

    <script>
        function detectarLocalizacao() {
            const status = document.getElementById('geo-status');
            if (!navigator.geolocation) {
                status.innerText = 'GPS não suportado';
                return;
            }
            status.innerText = 'Obtendo local...';
            navigator.geolocation.getCurrentPosition(
                (pos) => {
                    const lat = pos.coords.latitude;
                    const lng = pos.coords.longitude;
                    document.getElementById('search-lat').value = lat;
                    document.getElementById('search-lng').value = lng;
                    status.innerText = '📍 Mais próximos de você';
                    
                    const q = document.getElementById('search-q').value;
                    const type = document.getElementById('search-type').value;
                    const purpose = document.getElementById('search-purpose').value;
                    const maxPrice = document.getElementById('search-price').value;

                    htmx.ajax('GET', '/partials/imoveis?q=' + encodeURIComponent(q) + '&type=' + encodeURIComponent(type) + '&purpose=' + encodeURIComponent(purpose) + '&max_price=' + encodeURIComponent(maxPrice) + '&lat=' + lat + '&lng=' + lng, {
                        target: '#lista-imoveis',
                        swap: 'innerHTML'
                    });
                },
                () => {
                    status.innerText = 'Permissão negada';
                }
            );
        }

        // Tenta detectar automaticamente ao entrar na página
        window.addEventListener('DOMContentLoaded', () => {
            if (navigator.geolocation) {
                navigator.geolocation.getCurrentPosition((pos) => {
                    document.getElementById('search-lat').value = pos.coords.latitude;
                    document.getElementById('search-lng').value = pos.coords.longitude;
                    document.getElementById('geo-status').innerText = '📍 Ordenado por proximidade';
                    htmx.ajax('GET', '/partials/imoveis?lat=' + pos.coords.latitude + '&lng=' + pos.coords.longitude, {
                        target: '#lista-imoveis',
                        swap: 'innerHTML'
                    });
                }, () => {});
            }
        });

        function setPurpose(btn, purpose) {
            document.querySelectorAll('.search-tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById('search-purpose').value = purpose;
            
            const q = document.getElementById('search-q').value;
            const type = document.getElementById('search-type').value;
            const maxPrice = document.getElementById('search-price').value;
            const lat = document.getElementById('search-lat').value;
            const lng = document.getElementById('search-lng').value;

            htmx.ajax('GET', '/partials/imoveis?q=' + encodeURIComponent(q) + '&type=' + encodeURIComponent(type) + '&purpose=' + encodeURIComponent(purpose) + '&max_price=' + encodeURIComponent(maxPrice) + '&lat=' + lat + '&lng=' + lng, {
                target: '#lista-imoveis',
                swap: 'innerHTML'
            });
        }
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.GET("/partials/imoveis", func(c *gin.Context) {
		q := c.Query("q")
		propType := c.Query("type")
		purpose := c.Query("purpose")
		maxPrice := c.Query("max_price")
		lat := c.Query("lat")
		lng := c.Query("lng")

		endpoint := fmt.Sprintf("/api/properties?q=%s&type=%s&purpose=%s&max_price=%s&lat=%s&lng=%s", q, propType, purpose, maxPrice, lat, lng)
		var apiRes struct {
			Data []PropertyItem `json:"data"`
		}
		_ = apiClient.Get(endpoint, "", &apiRes)

		if len(apiRes.Data) == 0 {
			emptyMsg := `
            <div class="property-card" style="padding: 3rem 1.5rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                <p style="font-size: 1.1rem; font-weight: 600;">Nenhum imóvel encontrado para os filtros selecionados.</p>
            </div>`
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(emptyMsg))
			return
		}

		cards := ""
		for _, item := range apiRes.Data {
			cards += renderPropertyCard(item, false)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(cards))
	})

	// ==========================================
	// PÁGINA: DETALHE DO IMÓVEL COM CARROSSEL REAL
	// ==========================================
	r.GET("/imoveis/:id", func(c *gin.Context) {
		id := c.Param("id")

		var prop PropertyItem
		err := apiClient.Get("/api/properties/"+id, "", &prop)
		if err != nil || prop.ID == "" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		photos := parsePhotos(prop.CoverImageUrl)

		purposeAction := "Fazer Proposta de Aluguel"
		purposeLabel := "Aluguel"
		priceSuffix := "<span>/mês</span>"
		if prop.Purpose == "SALE" {
			purposeAction = "Fazer Oferta de Compra"
			purposeLabel = "Venda"
			priceSuffix = ""
		}

		desc := "Imóvel pronto para morar/investir com negociação direta pela plataforma."
		if prop.Description != nil && *prop.Description != "" {
			desc = *prop.Description
		}

		areaVal := "—"
		if prop.AreaM2 != nil && *prop.AreaM2 > 0 {
			areaVal = fmt.Sprintf("%.0fm²", *prop.AreaM2)
		}

		slidesHTML := ""
		thumbsHTML := ""
		for i, img := range photos {
			activeClass := "display: none;"
			if i == 0 {
				activeClass = "display: block;"
			}
			slidesHTML += fmt.Sprintf(`<img id="slide-%d" src="%s" class="carousel-slide" style="%s; width: 100%%; height: 460px; object-fit: cover; border-radius: 1rem;">`, i, img, activeClass)
			thumbsHTML += fmt.Sprintf(`<img src="%s" onclick="goToSlide(%d)" style="width: 80px; height: 60px; object-fit: cover; border-radius: 0.5rem; cursor: pointer; border: 2px solid var(--border);" class="carousel-thumb" id="thumb-%d">`, img, i, i)
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
                <div style="position: relative; border-radius: 1rem; overflow: hidden; box-shadow: var(--shadow-lg);">
                    ` + slidesHTML + `
                    <button onclick="prevSlide()" style="position: absolute; left: 1rem; top: 50%; transform: translateY(-50%); background: rgba(0,0,0,0.6); color: #fff; border: none; width: 44px; height: 44px; border-radius: 50%; cursor: pointer; font-size: 1.5rem; display: flex; align-items: center; justify-content: center;">&#10094;</button>
                    <button onclick="nextSlide()" style="position: absolute; right: 1rem; top: 50%; transform: translateY(-50%); background: rgba(0,0,0,0.6); color: #fff; border: none; width: 44px; height: 44px; border-radius: 50%; cursor: pointer; font-size: 1.5rem; display: flex; align-items: center; justify-content: center;">&#10095;</button>
                </div>
                <div style="display: flex; gap: 0.75rem; margin-top: 1rem; overflow-x: auto; padding-bottom: 0.5rem;">
                    ` + thumbsHTML + `
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

                    <form hx-post="/imoveis/` + id + `/proposta" hx-target="#lead-feedback" style="margin-top: 1rem;">
                        <h4 style="font-weight: 700; font-size: 1.1rem; margin-bottom: 0.75rem;">` + purposeAction + `</h4>
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
                            <label>Mensagem / Termos da Proposta</label>
                            <textarea class="form-control" name="message" rows="3" placeholder="Tenho interesse em fechar o contrato. Gostaria de agendar visita e enviar meus documentos para análise de aluguel."></textarea>
                        </div>
                        <button type="submit" class="btn btn-primary" style="width: 100%; padding: 0.85rem; margin-top: 0.5rem;">
                            Enviar Proposta Oficial
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </main>

    <script>
        let currentSlide = 0;
        const totalSlides = ` + strconv.Itoa(len(photos)) + `;

        function showSlide(index) {
            document.querySelectorAll('.carousel-slide').forEach(s => s.style.display = 'none');
            document.querySelectorAll('.carousel-thumb').forEach(t => t.style.borderColor = 'var(--border)');
            const slide = document.getElementById('slide-' + index);
            const thumb = document.getElementById('thumb-' + index);
            if (slide) slide.style.display = 'block';
            if (thumb) thumb.style.borderColor = 'var(--primary)';
            currentSlide = index;
        }
        function nextSlide() { showSlide((currentSlide + 1) % totalSlides); }
        function prevSlide() { showSlide((currentSlide - 1 + totalSlides) % totalSlides); }
        function goToSlide(index) { showSlide(index); }
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.POST("/imoveis/:id/proposta", func(c *gin.Context) {
		name := c.PostForm("name")
		email := c.PostForm("email")
		phone := c.PostForm("phone")
		message := c.PostForm("message")
		id := c.Param("id")

		payload := map[string]interface{}{
			"name":    name,
			"email":   email,
			"phone":   phone,
			"message": message,
		}
		_ = apiClient.Post("/api/properties/"+id+"/leads", payload, "", nil)

		msg := fmt.Sprintf(`<div class="alert alert-success">
            <strong>Proposta enviada com sucesso, %s!</strong><br>
            O proprietário recebeu seu contato e a proposta já está visível no painel dele.
        </div>`, name)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(msg))
	})

	// ==========================================================
	// PAINEL PROFISSIONAL COM SIDEBAR LATERAL NATIVA (/painel)
	// ==========================================
	r.GET("/painel", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/painel/imoveis")
	})

	r.GET("/painel/:tab", func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		tab := c.Param("tab")
		if tab == "" {
			tab = "imoveis"
		}

		var meRes struct {
			User  UserProfile `json:"user"`
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

		var leadsRes struct {
			Data []LeadItem `json:"data"`
		}
		_ = apiClient.Get("/api/me/leads", token, &leadsRes)

		// Classes ativas para a Sidebar
		imoveisClass := ""
		propostasClass := ""
		contratosClass := ""
		financeiroClass := ""
		perfilClass := ""

		mainContent := ""
		pageTitle := ""

		if tab == "imoveis" {
			imoveisClass = "active"
			pageTitle = "Gestão de Anúncios"

			cardsHTML := ""
			if len(propsRes.Data) == 0 {
				cardsHTML = `
                <div class="property-card" style="padding: 4rem 2rem; text-align: center; color: var(--text-muted); grid-column: 1 / -1; border-style: dashed;">
                    <p style="font-size: 1.1rem; font-weight: 600;">Você ainda não possui anúncios ativos.</p>
                    <p style="margin-top: 0.25rem;">Comece agora mesmo a divulgar seus imóveis para milhares de compradores.</p>
                    <button onclick="abrirModalCriacao()" class="btn btn-primary" style="margin-top: 1.5rem;">+ Publicar Primeiro Imóvel</button>
                </div>`
			} else {
				for _, p := range propsRes.Data {
					cardsHTML += renderPropertyCard(p, true)
				}
			}

			mainContent = fmt.Sprintf(`
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
                <div>
                    <h1 style="font-size: 1.85rem; font-weight: 800; letter-spacing: -0.5px;">Meus Imóveis</h1>
                    <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                        Você está utilizando <strong>%d de %d anúncios gratuitos</strong> do seu plano Freemium.
                    </p>
                </div>
                <button onclick="abrirModalCriacao()" class="btn btn-primary">+ Novo Anúncio</button>
            </div>
            <div class="property-grid">%s</div>`, meRes.Quota.Used, meRes.Quota.Limit, cardsHTML)

		} else if tab == "propostas" {
			propostasClass = "active"
			pageTitle = "Propostas Recebidas"

			leadsHTML := ""
			if len(leadsRes.Data) == 0 {
				leadsHTML = `<div style="background: #fff; padding: 4rem; text-align: center; border-radius: 1rem; border: 1px dashed var(--border); color: var(--text-muted);">
                    <p style="font-size: 1.1rem; font-weight: 600;">Nenhuma proposta recebida até o momento.</p>
                    <p style="font-size: 0.9rem; margin-top: 0.25rem;">Assim que um interessado enviar uma oferta ou mensagem, ela aparecerá aqui.</p>
                </div>`
			} else {
				leadsHTML = `<div style="display: flex; flex-direction: column; gap: 1rem;">`
				for _, l := range leadsRes.Data {
					phoneVal := "Não informado"
					if l.Phone != nil && *l.Phone != "" {
						phoneVal = *l.Phone
					}
					emailVal := "Não informado"
					if l.Email != nil && *l.Email != "" {
						emailVal = *l.Email
					}
					msgVal := "Sem mensagem"
					if l.Message != nil && *l.Message != "" {
						msgVal = *l.Message
					}

					leadsHTML += fmt.Sprintf(`
                    <div style="background: #ffffff; padding: 1.5rem; border-radius: 0.75rem; border: 1px solid var(--border); box-shadow: var(--shadow-sm);">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem;">
                            <span style="font-weight: 700; font-size: 1.1rem; color: var(--text-main);">%s</span>
                            <span class="badge-purpose" style="position: static;">Imóvel: %s</span>
                        </div>
                        <p style="margin-bottom: 0.75rem; color: #334155; line-height: 1.6;"><strong>Mensagem:</strong> %s</p>
                        <div style="display: flex; align-items: center; gap: 1.5rem; font-size: 0.875rem; color: var(--text-muted); border-top: 1px solid #f1f5f9; padding-top: 0.75rem;">
                            <div>📧 %s</div>
                            <div>📱 %s</div>
                            <a href="https://wa.me/55%s" target="_blank" class="btn btn-primary" style="margin-left: auto; padding: 0.35rem 0.85rem; font-size: 0.85rem;">Chamar no WhatsApp</a>
                        </div>
                    </div>`, l.Name, l.PropertyTitle, msgVal, emailVal, phoneVal, strings.ReplaceAll(phoneVal, " ", ""))
				}
				leadsHTML += `</div>`
			}

			mainContent = fmt.Sprintf(`
            <div style="margin-bottom: 2rem;">
                <h1 style="font-size: 1.85rem; font-weight: 800; letter-spacing: -0.5px;">Propostas & Contatos</h1>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Responda aos clientes interessados em alugar ou comprar seus imóveis.
                </p>
            </div>
            %s`, leadsHTML)

		} else if tab == "contratos" {
			contratosClass = "active"
			pageTitle = "Contratos & Carnê de Locação"

			var contractsRes struct {
				Data []ContractItem `json:"data"`
			}
			_ = apiClient.Get("/api/me/contracts", token, &contractsRes)

			contractsListHTML := ""
			if len(contractsRes.Data) == 0 {
				contractsListHTML = `<div style="background: #fff; padding: 4rem; text-align: center; border-radius: 1rem; border: 1px dashed var(--border); color: var(--text-muted);">
                    <p style="font-size: 1.1rem; font-weight: 600;">Nenhum contrato de aluguel ativo.</p>
                    <p style="font-size: 0.9rem; margin-top: 0.25rem;">Vincule um inquilino a um de seus imóveis para gerar o carnê mensal de cobrança e recibos.</p>
                    <button onclick="abrirModalContrato()" class="btn btn-primary" style="margin-top: 1.5rem;">+ Cadastrar Novo Contrato</button>
                </div>`
			} else {
				contractsListHTML = `<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1.5rem;">`
				for _, c := range contractsRes.Data {
					phoneVal := "—"
					if c.TenantPhone != nil && *c.TenantPhone != "" {
						phoneVal = *c.TenantPhone
					}
					contractsListHTML += fmt.Sprintf(`
                    <div style="background: #ffffff; padding: 1.5rem; border-radius: 1rem; border: 1px solid var(--border); box-shadow: var(--shadow-sm); display: flex; flex-direction: column; justify-content: space-between;">
                        <div>
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">
                                <span class="badge-purpose" style="position: static;">Ativo</span>
                                <span style="font-size: 0.85rem; color: var(--text-muted);">Vencimento: Dia %d</span>
                            </div>
                            <h3 style="font-size: 1.2rem; font-weight: 700; margin-bottom: 0.25rem;">%s</h3>
                            <div style="font-size: 0.85rem; color: var(--text-muted); margin-bottom: 1rem;">📍 %s - %s</div>
                            <div style="background: #f8fafc; padding: 0.75rem 1rem; border-radius: 0.5rem; margin-bottom: 1rem; font-size: 0.9rem;">
                                <div><strong>Inquilino:</strong> %s</div>
                                <div><strong>WhatsApp:</strong> %s</div>
                                <div style="margin-top: 0.25rem; font-size: 1.1rem; font-weight: 800; color: var(--text-main);">R$ %.2f / mês</div>
                            </div>
                        </div>
                        <button onclick="carregarCarneContrato('%s', '%s', %.2f)" class="btn btn-primary" style="width: 100%%; padding: 0.65rem;">
                            📑 Abrir Carnê Mensal
                        </button>
                    </div>`, c.DueDay, c.PropertyTitle, c.PropertyCity, c.PropertyState, c.TenantName, phoneVal, c.RentAmount, c.ID, c.PropertyTitle, c.RentAmount)
				}
				contractsListHTML += `</div>`
			}

			// Renderizar opções de imóveis do usuário para o modal de contrato
			propsOptionsHTML := ""
			for _, p := range propsRes.Data {
				propsOptionsHTML += fmt.Sprintf(`<option value="%s">%s (R$ %.2f)</option>`, p.ID, p.Title, p.Price)
			}

			mainContent = fmt.Sprintf(`
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
                <div>
                    <h1 style="font-size: 1.85rem; font-weight: 800; letter-spacing: -0.5px;">Contratos & Carnê de Locação</h1>
                    <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                        Gerencie os aluguéis, dê baixa manual com recibo em PDF ou reemita cobranças com novo prazo.
                    </p>
                </div>
                <button onclick="abrirModalContrato()" class="btn btn-primary">+ Novo Contrato</button>
            </div>

            <!-- Tabela comparativa de taxa da baixa manual -->
            <div style="background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 0.75rem; padding: 1.25rem; margin-bottom: 2rem; display: flex; align-items: center; gap: 1.5rem; flex-wrap: wrap;">
                <div style="font-size: 1.5rem;">💡</div>
                <div style="flex: 1; font-size: 0.9rem; color: #1e3a8a;">
                    <strong>Taxa de Serviço para Baixa Manual & Emissão de Recibo:</strong><br>
                    Aluguel até R$ 500: <strong>R$ 3,00</strong> &bull; De R$ 501 a R$ 1.500: <strong>R$ 5,00</strong> &bull; Acima de R$ 1.500: <strong>R$ 7,00</strong>.
                </div>
            </div>

            <div id="contratos-grid">%s</div>

            <!-- Container dinâmico do Carnê do Contrato Selecionado -->
            <div id="carne-container" style="display: none; margin-top: 3rem;"></div>`, contractsListHTML)

		} else if tab == "financeiro" {
			financeiroClass = "active"
			pageTitle = "Planos & Cotas de Anúncio"

			var pkgsRes struct {
				Packages []PlanPackageItem `json:"packages"`
			}
			_ = apiClient.Get("/api/plans/packages", token, &pkgsRes)

			var payRes struct {
				Data []PaymentRecord `json:"data"`
			}
			_ = apiClient.Get("/api/me/payments", token, &payRes)

			packagesHTML := `<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 3rem;">`
			for _, pkg := range pkgsRes.Packages {
				packagesHTML += fmt.Sprintf(`
                <div style="background: #ffffff; border: 2px solid var(--border); border-radius: 1rem; padding: 2rem; display: flex; flex-direction: column; justify-content: space-between; position: relative; transition: all 0.2s;" onmouseover="this.style.borderColor='var(--primary)'" onmouseout="this.style.borderColor='var(--border)'">
                    <div>
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">
                            <h3 style="font-size: 1.3rem; font-weight: 800;">%s</h3>
                            <span style="background: #eff6ff; color: var(--primary); padding: 0.25rem 0.65rem; border-radius: 9999px; font-size: 0.8rem; font-weight: 700;">+%d Imóveis</span>
                        </div>
                        <p style="color: var(--text-muted); font-size: 0.9rem; margin-bottom: 1.5rem;">%s</p>
                        <div style="font-size: 2rem; font-weight: 800; color: var(--text-main); margin-bottom: 1.5rem;">%s</div>
                    </div>
                    <button onclick="abrirCheckoutPix('%s', '%s', '%s')" class="btn btn-primary" style="width: 100%%; padding: 0.75rem;">
                        Pagar via Pix Instantâneo
                    </button>
                </div>`, pkg.Title, pkg.Listings, pkg.Description, pkg.PriceFormatted, pkg.ID, pkg.Title, pkg.PriceFormatted)
			}
			packagesHTML += `</div>`

			historyHTML := ""
			if len(payRes.Data) == 0 {
				historyHTML = `<div style="padding: 2.5rem; text-align: center; color: var(--text-muted); background: #fff; border-radius: 0.75rem; border: 1px dashed var(--border);">Nenhum pagamento realizado ainda.</div>`
			} else {
				historyHTML = `<div style="background: #fff; border-radius: 0.75rem; border: 1px solid var(--border); overflow: hidden;">
                <table style="width: 100%%; border-collapse: collapse; text-align: left; font-size: 0.9rem;">
                    <thead>
                        <tr style="background: #f8fafc; border-bottom: 1px solid var(--border); color: var(--text-muted);">
                            <th style="padding: 1rem;">Data</th>
                            <th style="padding: 1rem;">Descrição</th>
                            <th style="padding: 1rem;">Valor</th>
                            <th style="padding: 1rem;">Método</th>
                            <th style="padding: 1rem;">Status</th>
                            <th style="padding: 1rem;">Ação</th>
                        </tr>
                    </thead>
                    <tbody>`
				for _, p := range payRes.Data {
					statusBadge := `<span style="background: #fef3c7; color: #b45309; padding: 0.2rem 0.5rem; border-radius: 4px; font-weight: 600; font-size: 0.75rem;">PENDENTE</span>`
					actionBtn := ""
					if p.Status == "PAID" {
						statusBadge = `<span style="background: #dcfce7; color: #15803d; padding: 0.2rem 0.5rem; border-radius: 4px; font-weight: 600; font-size: 0.75rem;">PAGO</span>`
					} else {
						actionBtn = fmt.Sprintf(`<button onclick="simularAprovacao('%s')" class="btn btn-outline" style="padding: 0.2rem 0.6rem; font-size: 0.75rem;">⚡ Simular Aprovação</button>`, p.ID)
					}

					historyHTML += fmt.Sprintf(`
                    <tr style="border-bottom: 1px solid var(--border);">
                        <td style="padding: 1rem;">%s</td>
                        <td style="padding: 1rem; font-weight: 600;">+%d Anúncios Adicionais</td>
                        <td style="padding: 1rem;">R$ %.2f</td>
                        <td style="padding: 1rem;">%s</td>
                        <td style="padding: 1rem;">%s</td>
                        <td style="padding: 1rem;">%s</td>
                    </tr>`, p.CreatedAt[:10], p.ItemsCount, p.Amount, p.Method, statusBadge, actionBtn)
				}
				historyHTML += `</tbody></table></div>`
			}

			mainContent = fmt.Sprintf(`
            <div style="margin-bottom: 2rem;">
                <h1 style="font-size: 1.85rem; font-weight: 800; letter-spacing: -0.5px;">Planos & Cotas de Anúncio</h1>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Seu limite atual é de <strong>%d anúncios</strong> (utilizando %d). Adquira pacotes permanentes e amplie seu catálogo.
                </p>
            </div>

            <div style="background: linear-gradient(135deg, #1e293b, #0f172a); color: #fff; padding: 2rem; border-radius: 1rem; margin-bottom: 2.5rem; display: flex; justify-content: space-between; align-items: center;">
                <div>
                    <div style="font-size: 0.85rem; text-transform: uppercase; letter-spacing: 1px; color: #94a3b8; font-weight: 700;">Capacidade de Anúncios</div>
                    <div style="font-size: 2.5rem; font-weight: 800; margin-top: 0.25rem;">%d / %d <span style="font-size: 1rem; font-weight: 500; color: #cbd5e1;">ativos</span></div>
                </div>
                <div style="text-align: right;">
                    <div style="font-size: 0.9rem; color: #94a3b8;">Status da Conta</div>
                    <div style="font-size: 1.1rem; font-weight: 700; color: #38bdf8;">Ativo &bull; Freemium / Recorrente</div>
                </div>
            </div>

            <h2 style="font-size: 1.35rem; font-weight: 800; margin-bottom: 1.25rem;">Escolha um Pacote de Anúncios</h2>
            %s

            <h2 style="font-size: 1.35rem; font-weight: 800; margin-bottom: 1.25rem; margin-top: 2rem;">Histórico de Pagamentos</h2>
            %s`, meRes.Quota.Limit, meRes.Quota.Used, meRes.Quota.Used, meRes.Quota.Limit, packagesHTML, historyHTML)

		} else if tab == "perfil" {
			perfilClass = "active"
			pageTitle = "Meu Perfil"

			phoneVal := ""
			if meRes.User.Phone != nil {
				phoneVal = *meRes.User.Phone
			}
			cpfVal := ""
			if meRes.User.CpfCnpj != nil {
				cpfVal = *meRes.User.CpfCnpj
			}

			mainContent = fmt.Sprintf(`
            <div style="margin-bottom: 2rem;">
                <h1 style="font-size: 1.85rem; font-weight: 800; letter-spacing: -0.5px;">Meu Perfil & Segurança</h1>
                <p style="color: var(--text-muted); font-size: 0.95rem; margin-top: 0.25rem;">
                    Gerencie seus dados pessoais e credenciais de acesso.
                </p>
            </div>

            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
                <div style="background: #ffffff; padding: 2rem; border-radius: 1rem; border: 1px solid var(--border); box-shadow: var(--shadow-sm);">
                    <h3 style="font-size: 1.25rem; font-weight: 700; margin-bottom: 1.25rem;">Dados Cadastrais</h3>
                    <div id="profile-feedback"></div>
                    <form hx-put="/painel/perfil/atualizar" hx-target="#profile-feedback">
                        <div class="form-group">
                            <label>Nome Completo</label>
                            <input class="form-control" type="text" name="name" required value="%s">
                        </div>
                        <div class="form-group">
                            <label>E-mail (Login)</label>
                            <input class="form-control" type="text" disabled value="%s" style="background: #e2e8f0; cursor: not-allowed;">
                        </div>
                        <div class="form-group">
                            <label>Telefone / WhatsApp</label>
                            <input class="form-control" type="text" name="phone" value="%s">
                        </div>
                        <div class="form-group">
                            <label>CPF ou CNPJ</label>
                            <input class="form-control" type="text" name="cpf_cnpj" value="%s" placeholder="000.000.000-00">
                        </div>
                        <button type="submit" class="btn btn-primary" style="width: 100%%; margin-top: 1rem;">Salvar Alterações</button>
                    </form>
                </div>

                <div style="background: #ffffff; padding: 2rem; border-radius: 1rem; border: 1px solid var(--border); box-shadow: var(--shadow-sm);">
                    <h3 style="font-size: 1.25rem; font-weight: 700; margin-bottom: 1.25rem;">Alterar Senha</h3>
                    <div id="password-feedback"></div>
                    <form hx-post="/painel/perfil/senha" hx-target="#password-feedback">
                        <div class="form-group">
                            <label>Senha Atual</label>
                            <input class="form-control" type="password" name="current_password" required placeholder="Digite sua senha atual">
                        </div>
                        <div class="form-group">
                            <label>Nova Senha</label>
                            <input class="form-control" type="password" name="new_password" required minlength="6" placeholder="Mínimo 6 caracteres">
                        </div>
                        <button type="submit" class="btn btn-outline" style="width: 100%%; margin-top: 1rem;">Atualizar Senha</button>
                    </form>
                </div>
            </div>`, meRes.User.Name, meRes.User.Email, phoneVal, cpfVal)
		}

		userInitial := "U"
		if len(meRes.User.Name) > 0 {
			userInitial = strings.ToUpper(string(meRes.User.Name[0]))
		}

		html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + pageTitle + ` — Painel Imovxen</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
</head>
<body>
    <div class="dashboard-layout">
        <!-- SIDEBAR LATERAL MODERNA -->
        <aside class="sidebar">
            <div class="sidebar-header">
                <a href="/" class="logo-wrapper">
                    <div class="logo-icon">IX</div>
                    <div class="logo-text">Imov<span>xen</span></div>
                </a>
            </div>

            <nav class="sidebar-nav">
                <a href="/painel/imoveis" class="sidebar-link ` + imoveisClass + `">
                    <span class="icon">🏠</span>
                    <span>Meus Imóveis</span>
                </a>
                <a href="/painel/propostas" class="sidebar-link ` + propostasClass + `">
                    <span class="icon">📬</span>
                    <span>Propostas</span>
                </a>
                <a href="/painel/contratos" class="sidebar-link ` + contratosClass + `">
                    <span class="icon">📜</span>
                    <span>Contratos & Carnê</span>
                </a>
                <a href="/painel/financeiro" class="sidebar-link ` + financeiroClass + `">
                    <span class="icon">💳</span>
                    <span>Planos & Cotas</span>
                </a>
                <a href="/painel/perfil" class="sidebar-link ` + perfilClass + `">
                    <span class="icon">👤</span>
                    <span>Meu Perfil</span>
                </a>
                <div style="height: 1px; background: var(--border); margin: 1rem 0;"></div>
                <a href="/" class="sidebar-link">
                    <span class="icon">🌐</span>
                    <span>Ver Vitrine Pública</span>
                </a>
            </nav>

            <div class="sidebar-footer">
                <div class="sidebar-user">
                    <div class="user-avatar">` + userInitial + `</div>
                    <div style="overflow: hidden; flex: 1;">
                        <div style="font-weight: 700; font-size: 0.9rem; white-space: nowrap; text-overflow: ellipsis; overflow: hidden;">` + meRes.User.Name + `</div>
                        <div style="font-size: 0.75rem; color: var(--text-muted); white-space: nowrap; text-overflow: ellipsis; overflow: hidden;">` + meRes.User.Email + `</div>
                    </div>
                    <a href="/logout" title="Sair" style="color: #dc2626; text-decoration: none; font-size: 1.1rem; padding: 0.25rem;">🚪</a>
                </div>
            </div>
        </aside>

        <!-- ÁREA PRINCIPAL DO DASHBOARD -->
        <div class="dashboard-main">
            <header class="dashboard-header">
                <div style="font-weight: 600; color: var(--text-muted); font-size: 0.9rem;">
                    Painel do Anunciante &bull; ` + pageTitle + `
                </div>
                <div style="display: flex; gap: 1rem; align-items: center;">
                    <a href="/" class="btn btn-outline" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">Explorar Marketplace</a>
                    <button onclick="abrirModalCriacao()" class="btn btn-primary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">+ Publicar Anúncio</button>
                </div>
            </header>

            <div class="dashboard-body">
                ` + mainContent + `
            </div>
        </div>
    </div>

    <!-- MODAL DE CRIAÇÃO / EDIÇÃO COM UPLOAD REAL -->
    <div id="modal-imovel" style="display: none; position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); z-index: 100; overflow-y: auto; padding: 2.5rem 0;">
        <div class="form-container" style="margin: auto; max-width: 650px;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                <h3 id="modal-title" style="font-size: 1.35rem; font-weight: 700;">Publicar Novo Imóvel</h3>
                <button onclick="fecharModal()" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted);">&times;</button>
            </div>

            <div id="modal-feedback"></div>

            <form id="form-imovel" onsubmit="enviarImovelComUpload(event)">
                <input type="hidden" id="form-id" name="id" value="">
                <input type="hidden" id="form-photos" name="photos" value="">
                
                <div class="form-group">
                    <label>Título do Anúncio</label>
                    <input id="form-title" class="form-control" type="text" name="title" required placeholder="Ex: Apartamento reformado em Pinheiros">
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                    <div class="form-group">
                        <label>Tipo de Imóvel</label>
                        <select id="form-type" class="form-control" name="type" required>
                            <option value="APARTMENT">Apartamento</option>
                            <option value="HOUSE">Casa</option>
                            <option value="COMMERCIAL">Comercial</option>
                            <option value="LAND">Terreno</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>Finalidade</label>
                        <select id="form-purpose" class="form-control" name="purpose" required>
                            <option value="RENT">Aluguel</option>
                            <option value="SALE">Venda</option>
                        </select>
                    </div>
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                    <div class="form-group">
                        <label>Preço (R$)</label>
                        <input id="form-price" class="form-control" type="number" step="0.01" name="price" required placeholder="3500.00">
                    </div>
                    <div class="form-group">
                        <label>Área (m²)</label>
                        <input id="form-area" class="form-control" type="number" step="0.1" name="area_m2" placeholder="65.0">
                    </div>
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                    <div class="form-group">
                        <label>Cidade</label>
                        <input id="form-city" class="form-control" type="text" name="city" required placeholder="São Paulo">
                    </div>
                    <div class="form-group">
                        <label>Estado (UF)</label>
                        <input id="form-state" class="form-control" type="text" name="state" required maxlength="2" placeholder="SP">
                    </div>
                </div>

                <div class="form-group" style="margin-top: 1rem;">
                    <label>Fotos do Imóvel (Envie fotos reais do seu computador)</label>
                    <input id="form-files" class="form-control" type="file" multiple accept="image/*" style="padding: 0.5rem; background: #fff;">
                    <small style="color: var(--text-muted); font-size: 0.8rem;">Selecione uma ou mais fotos para o carrossel do anúncio.</small>
                </div>

                <button type="submit" id="btn-submit-modal" class="btn btn-primary" style="width: 100%; margin-top: 1.5rem; padding: 0.85rem;">Salvar Imóvel</button>
            </form>
        </div>
    </div>

    <!-- MODAL DE CADASTRO DE CONTRATO -->
    <div id="modal-contrato" style="display: none; position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); z-index: 100; overflow-y: auto; padding: 2.5rem 0;">
        <div class="form-container" style="margin: auto; max-width: 580px;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                <h3 style="font-size: 1.35rem; font-weight: 700;">Novo Contrato de Locação</h3>
                <button onclick="fecharModalContrato()" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted);">&times;</button>
            </div>
            <div id="contrato-feedback"></div>
            <form onsubmit="salvarNovoContrato(event)">
                <div class="form-group">
                    <label>Selecione o Imóvel</label>
                    <select id="contrato-property-id" class="form-control" required>
                        <option value="">Selecione um imóvel...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Nome do Inquilino</label>
                    <input id="contrato-tenant-name" class="form-control" type="text" required placeholder="Nome completo do inquilino">
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                    <div class="form-group">
                        <label>WhatsApp / Telefone</label>
                        <input id="contrato-tenant-phone" class="form-control" type="text" placeholder="(11) 99999-9999">
                    </div>
                    <div class="form-group">
                        <label>CPF do Inquilino</label>
                        <input id="contrato-tenant-cpf" class="form-control" type="text" placeholder="000.000.000-00">
                    </div>
                </div>
                <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;">
                    <div class="form-group">
                        <label>Aluguel (R$)</label>
                        <input id="contrato-rent-amount" class="form-control" type="number" step="0.01" required placeholder="Ex: 450.00">
                    </div>
                    <div class="form-group">
                        <label>Dia Vencimento</label>
                        <input id="contrato-due-day" class="form-control" type="number" min="1" max="31" required value="10">
                    </div>
                    <div class="form-group">
                        <label>Meses (Duração)</label>
                        <input id="contrato-duration" class="form-control" type="number" min="1" max="60" required value="12">
                    </div>
                </div>
                <button type="submit" id="btn-submit-contrato" class="btn btn-primary" style="width: 100%; margin-top: 1rem; padding: 0.85rem;">Gerar Contrato & Carnê</button>
            </form>
        </div>
    </div>

    <!-- MODAL DE CHECKOUT PIX -->
    <div id="modal-pix" style="display: none; position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); z-index: 100; overflow-y: auto; padding: 2.5rem 0;">
        <div class="form-container" style="margin: auto; max-width: 500px; text-align: center;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
                <h3 id="pix-title" style="font-size: 1.35rem; font-weight: 700;">Pagamento Pix</h3>
                <button onclick="fecharModalPix()" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted);">&times;</button>
            </div>
            <div id="pix-loading" style="padding: 2rem; color: var(--text-muted);">
                <div style="font-size: 2rem; margin-bottom: 0.5rem;">⏳</div>
                Gerando QR Code Pix junto ao Asaas...
            </div>
            <div id="pix-content" style="display: none;">
                <p style="font-size: 0.95rem; color: var(--text-muted); margin-bottom: 1rem;">Abra o app do seu banco e escaneie o código abaixo:</p>
                <div style="background: #fff; padding: 1rem; border-radius: 0.75rem; display: inline-block; border: 1px solid var(--border); margin-bottom: 1rem;">
                    <img id="pix-qr-img" src="" alt="QR Code Pix" style="max-width: 200px; height: auto;">
                </div>
                <div class="form-group">
                    <label style="text-align: left;">Código Pix Copia e Cola:</label>
                    <input id="pix-code-input" class="form-control" type="text" readonly style="font-size: 0.8rem; background: #f1f5f9;">
                </div>
                <button onclick="copiarPixCode()" class="btn btn-primary" style="width: 100%; margin-top: 0.5rem;">📋 Copiar Código Pix</button>
                <p id="copy-feedback" style="color: #15803d; font-size: 0.85rem; margin-top: 0.5rem; display: none;">Código copiado para a área de transferência!</p>
                <div style="margin-top: 1.5rem; border-top: 1px solid var(--border); padding-top: 1rem;">
                    <span style="font-size: 0.85rem; color: var(--text-muted);">Identificamos o pagamento automaticamente.</span>
                </div>
            </div>
        </div>
    </div>

    <script>
        function abrirModalContrato() {
            document.getElementById('modal-contrato').style.display = 'block';
            fetch('/painel/imoveis/opcoes')
                .then(function(r) { return r.json(); })
                .then(function(data) {
                    var select = document.getElementById('contrato-property-id');
                    select.innerHTML = '<option value="">Selecione um imóvel...</option>';
                    if (data.data) {
                        data.data.forEach(function(p) {
                            var opt = document.createElement('option');
                            opt.value = p.id;
                            opt.textContent = p.title + ' (R$ ' + p.price.toFixed(2) + ')';
                            select.appendChild(opt);
                        });
                    }
                });
        }

        function fecharModalContrato() {
            document.getElementById('modal-contrato').style.display = 'none';
        }

        async function salvarNovoContrato(e) {
            e.preventDefault();
            var btn = document.getElementById('btn-submit-contrato');
            btn.innerText = 'Gerando carnê...';
            btn.disabled = true;

            var postData = new URLSearchParams();
            postData.append('property_id', document.getElementById('contrato-property-id').value);
            postData.append('tenant_name', document.getElementById('contrato-tenant-name').value);
            postData.append('tenant_phone', document.getElementById('contrato-tenant-phone').value);
            postData.append('tenant_cpf', document.getElementById('contrato-tenant-cpf').value);
            postData.append('rent_amount', document.getElementById('contrato-rent-amount').value);
            postData.append('due_day', document.getElementById('contrato-due-day').value);
            postData.append('duration_months', document.getElementById('contrato-duration').value);

            var res = await fetch('/painel/contratos/novo', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: postData.toString()
            });

            if (res.ok) {
                alert('Contrato cadastrado com sucesso! As parcelas do carnê foram geradas.');
                window.location.reload();
            } else {
                document.getElementById('contrato-feedback').innerHTML = '<div class="alert alert-danger">Erro ao criar contrato.</div>';
                btn.innerText = 'Gerar Contrato & Carnê';
                btn.disabled = false;
            }
        }

        async function carregarCarneContrato(contractId, propertyTitle, rentAmount) {
            var container = document.getElementById('carne-container');
            container.style.display = 'block';
            container.innerHTML = '<div style="text-align:center; padding: 2rem;">Carregando carnê de locação...</div>';
            container.scrollIntoView({ behavior: 'smooth' });

            var res = await fetch('/painel/contratos/' + contractId + '/carne');
            var data = await res.json();

            var feeLabel = 'R$ 7,00';
            if (rentAmount <= 500) feeLabel = 'R$ 3,00';
            else if (rentAmount <= 1500) feeLabel = 'R$ 5,00';

            var installmentsRows = '';
            data.installments.forEach(function(inst) {
                var badge = '<span style="background: #fef3c7; color: #b45309; padding: 0.2rem 0.6rem; border-radius: 4px; font-weight: 700; font-size: 0.75rem;">A VENCER / PENDENTE</span>';
                var actions = '<button onclick="darBaixaManual(\'' + inst.id + '\', ' + inst.installment_number + ', ' + inst.amount + ')" class="btn btn-primary" style="padding: 0.3rem 0.65rem; font-size: 0.8rem;">Dar Baixa Manual (' + feeLabel + ')</button> <button onclick="reemitirCobranca(\'' + inst.id + '\')" class="btn btn-outline" style="padding: 0.3rem 0.65rem; font-size: 0.8rem;">Reemitir</button>';

                if (inst.status === 'PAID_MANUAL' || inst.status === 'PAID_SYSTEM') {
                    badge = '<span style="background: #dcfce7; color: #15803d; padding: 0.2rem 0.6rem; border-radius: 4px; font-weight: 700; font-size: 0.75rem;">PAGO / BAIXADO</span>';
                    var code = inst.receipt_code || 'REC-OK';
                    actions = '<button onclick="abrirRecibo(\'' + inst.id + '\', \'' + code + '\', ' + inst.installment_number + ', ' + inst.amount + ', \'' + inst.due_date + '\')" class="btn btn-outline" style="padding: 0.3rem 0.65rem; font-size: 0.8rem;">Ver Recibo</button>';
                }

                installmentsRows += '<tr style="border-bottom: 1px solid var(--border);">' +
                    '<td style="padding: 1rem; font-weight: 700;">Parcela ' + inst.installment_number + '</td>' +
                    '<td style="padding: 1rem;">' + inst.due_date.substring(0, 10) + '</td>' +
                    '<td style="padding: 1rem; font-weight: 800;">R$ ' + inst.amount.toFixed(2) + '</td>' +
                    '<td style="padding: 1rem;">' + badge + '</td>' +
                    '<td style="padding: 1rem; display: flex; gap: 0.5rem; justify-content: flex-end;">' + actions + '</td>' +
                '</tr>';
            });

            container.innerHTML = '<div style="background: #ffffff; border-radius: 1rem; border: 1px solid var(--border); padding: 2rem; box-shadow: var(--shadow-sm);">' +
                '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; border-bottom: 1px solid var(--border); padding-bottom: 1rem;">' +
                    '<div>' +
                        '<span class="badge-purpose" style="position: static;">Carnê de Locação</span>' +
                        '<h2 style="font-size: 1.4rem; font-weight: 800; margin-top: 0.5rem;">' + propertyTitle + '</h2>' +
                        '<p style="color: var(--text-muted); font-size: 0.9rem;">Aluguel mensal: <strong>R$ ' + rentAmount.toFixed(2) + '</strong> &bull; Taxa de baixa manual para este imóvel: <strong>' + feeLabel + '</strong></p>' +
                    '</div>' +
                    '<button onclick="document.getElementById(\'carne-container\').style.display=\'none\'" class="btn btn-outline" style="font-size: 0.85rem;">Fechar Carnê</button>' +
                '</div>' +
                '<table style="width: 100%; border-collapse: collapse; text-align: left; font-size: 0.9rem;">' +
                    '<thead>' +
                        '<tr style="background: #f8fafc; border-bottom: 1px solid var(--border); color: var(--text-muted);">' +
                            '<th style="padding: 0.75rem 1rem;">Mês</th>' +
                            '<th style="padding: 0.75rem 1rem;">Vencimento</th>' +
                            '<th style="padding: 0.75rem 1rem;">Valor</th>' +
                            '<th style="padding: 0.75rem 1rem;">Situação</th>' +
                            '<th style="padding: 0.75rem 1rem; text-align: right;">Ações de Quitação</th>' +
                        '</tr>' +
                    '</thead>' +
                    '<tbody>' + installmentsRows + '</tbody>' +
                '</table>' +
            '</div>';
        }

        async function darBaixaManual(installmentId, num, amount) {
            var fee = 7.00;
            if (amount <= 500) fee = 3.00;
            else if (amount <= 1500) fee = 5.00;

            if (!confirm('Dar baixa manual na Parcela ' + num + '?\n\nO sistema cobrará a taxa de serviço de R$ ' + fee.toFixed(2) + ' e emitirá o Recibo Oficial de Quitação com validade jurídica.')) return;

            var res = await fetch('/painel/contratos/installments/' + installmentId + '/baixar-manual', {
                method: 'POST'
            });
            var data = await res.json();
            if (res.ok) {
                alert('Baixa manual confirmada com sucesso! Código do Recibo: ' + data.receipt_code);
                window.location.reload();
            } else {
                alert('Erro ao dar baixa manual.');
            }
        }

        async function reemitirCobranca(installmentId) {
            var novaData = prompt('Digite a nova data de vencimento (Formato: YYYY-MM-DD):', new Date().toISOString().substring(0, 10));
            if (!novaData) return;

            var res = await fetch('/painel/contratos/installments/' + installmentId + '/reemitir', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: 'new_due_date=' + encodeURIComponent(novaData)
            });

            if (res.ok) {
                alert('Cobrança reemitida com sucesso para o novo vencimento!');
                window.location.reload();
            } else {
                alert('Erro ao reemitir.');
            }
        }

        function abrirRecibo(installmentId, code, num, amount, due) {
            alert('RECIBO DE QUITAÇÃO DE ALUGUEL\n----------------------------------\nCódigo: ' + code + '\nParcela: ' + num + '\nVencimento: ' + due.substring(0, 10) + '\nValor: R$ ' + amount.toFixed(2) + '\nStatus: QUITADO INTEGRALMENTE VIA IMOVXEN\n----------------------------------\nDocumento registrado na plataforma.');
        }

        let pixPollInterval = null;

        function fecharModalPix() {
            document.getElementById('modal-pix').style.display = 'none';
            if (pixPollInterval) clearInterval(pixPollInterval);
        }

        async function abrirCheckoutPix(pkgId, title, price) {
            document.getElementById('pix-title').innerText = 'Pagamento Pix — ' + title;
            document.getElementById('modal-pix').style.display = 'block';
            document.getElementById('pix-loading').style.display = 'block';
            document.getElementById('pix-content').style.display = 'none';

            try {
                const res = await fetch('/painel/financeiro/checkout', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: 'package_id=' + encodeURIComponent(pkgId)
                });
                const data = await res.json();
                if (data.payment && data.payment.pix_qr_base64) {
                    document.getElementById('pix-loading').style.display = 'none';
                    document.getElementById('pix-content').style.display = 'block';
                    document.getElementById('pix-qr-img').src = 'data:image/png;base64,' + data.payment.pix_qr_base64;
                    document.getElementById('pix-code-input').value = data.payment.pix_code;

                    // Polling de verificação de status a cada 3 segundos
                    if (pixPollInterval) clearInterval(pixPollInterval);
                    pixPollInterval = setInterval(async () => {
                        const statusRes = await fetch('/painel/financeiro/status/' + data.payment.id);
                        const statusData = await statusRes.json();
                        if (statusData.status === 'PAID') {
                            clearInterval(pixPollInterval);
                            alert('🎉 Pagamento confirmado! Suas novas cotas foram liberadas com sucesso.');
                            window.location.reload();
                        }
                    }, 3000);
                } else {
                    alert('Erro ao gerar cobrança Pix. Tente novamente.');
                    fecharModalPix();
                }
            } catch (err) {
                alert('Erro de conexão com o servidor.');
                fecharModalPix();
            }
        }

        function copiarPixCode() {
            const input = document.getElementById('pix-code-input');
            input.select();
            navigator.clipboard.writeText(input.value);
            document.getElementById('copy-feedback').style.display = 'block';
            setTimeout(() => { document.getElementById('copy-feedback').style.display = 'none'; }, 3000);
        }

        async function simularAprovacao(paymentId) {
            if (!confirm('Deseja simular a aprovação instantânea desta fatura via webhook/trigger?')) return;
            const res = await fetch('/painel/financeiro/simular/' + paymentId, { method: 'POST' });
            if (res.ok) {
                alert('Pagamento aprovado e cota adicionada!');
                window.location.reload();
            } else {
                alert('Erro ao processar aprovação.');
            }
        }

        function abrirModalCriacao() {
            document.getElementById('modal-title').innerText = 'Publicar Novo Imóvel';
            document.getElementById('btn-submit-modal').innerText = 'Publicar Agora';
            document.getElementById('form-id').value = '';
            document.getElementById('form-title').value = '';
            document.getElementById('form-price').value = '';
            document.getElementById('form-area').value = '';
            document.getElementById('form-city').value = '';
            document.getElementById('form-state').value = '';
            document.getElementById('form-photos').value = '';
            document.getElementById('modal-imovel').style.display = 'block';
        }

        function abrirModalEdicao(id, title, type, purpose, price, area, city, state, photos) {
            document.getElementById('modal-title').innerText = 'Editar Anúncio';
            document.getElementById('btn-submit-modal').innerText = 'Atualizar Anúncio';
            document.getElementById('form-id').value = id;
            document.getElementById('form-title').value = title;
            document.getElementById('form-type').value = type;
            document.getElementById('form-purpose').value = purpose;
            document.getElementById('form-price').value = price;
            document.getElementById('form-area').value = area;
            document.getElementById('form-city').value = city;
            document.getElementById('form-state').value = state;
            document.getElementById('form-photos').value = photos;
            document.getElementById('modal-imovel').style.display = 'block';
        }

        function fecharModal() {
            document.getElementById('modal-imovel').style.display = 'none';
        }

        async function enviarImovelComUpload(e) {
            e.preventDefault();
            const btn = document.getElementById('btn-submit-modal');
            btn.innerText = 'Fazendo upload e salvando...';
            btn.disabled = true;

            const filesInput = document.getElementById('form-files');
            let photoUrls = document.getElementById('form-photos').value;

            if (filesInput.files.length > 0) {
                const formData = new FormData();
                for (let i = 0; i < filesInput.files.length; i++) {
                    formData.append('photos', filesInput.files[i]);
                }
                try {
                    const uploadRes = await fetch('http://localhost:8080/api/upload', {
                        method: 'POST',
                        body: formData
                    });
                    const uploadData = await uploadRes.json();
                    if (uploadData.urls && uploadData.urls.length > 0) {
                        photoUrls = uploadData.urls.join(',');
                    }
                } catch (err) {
                    console.error('Erro de upload:', err);
                }
            }

            const postData = new URLSearchParams();
            postData.append('id', document.getElementById('form-id').value);
            postData.append('title', document.getElementById('form-title').value);
            postData.append('type', document.getElementById('form-type').value);
            postData.append('purpose', document.getElementById('form-purpose').value);
            postData.append('price', document.getElementById('form-price').value);
            postData.append('area_m2', document.getElementById('form-area').value);
            postData.append('city', document.getElementById('form-city').value);
            postData.append('state', document.getElementById('form-state').value);
            postData.append('photos', photoUrls);

            const res = await fetch('/painel/imoveis/salvar', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: postData.toString()
            });

            if (res.ok) {
                window.location.reload();
            } else {
                document.getElementById('modal-feedback').innerHTML = '<div class="alert alert-danger">Erro ao salvar anúncio. Verifique sua cota.</div>';
                btn.innerText = 'Salvar Imóvel';
                btn.disabled = false;
            }
        }
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// Endpoints de Pagamento / Financeiro no Frontend
	r.POST("/painel/financeiro/checkout", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		payload := map[string]interface{}{
			"package_id": c.PostForm("package_id"),
		}
		var apiRes struct {
			Payment PaymentRecord   `json:"payment"`
			Package PlanPackageItem `json:"package"`
		}
		err := apiClient.Post("/api/payments/checkout", payload, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha no checkout"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	r.GET("/painel/financeiro/status/:id", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		paymentID := c.Param("id")
		var apiRes struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		err := apiClient.Get("/api/payments/"+paymentID+"/status", token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar status"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	r.POST("/painel/financeiro/simular/:id", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		paymentID := c.Param("id")
		var apiRes map[string]interface{}
		err := apiClient.Post("/api/payments/"+paymentID+"/simulate-approval", nil, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao simular"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	// Opções de imóveis para selects
	r.GET("/painel/imoveis/opcoes", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		var propsRes struct {
			Data []PropertyItem `json:"data"`
		}
		_ = apiClient.Get("/api/me/properties", token, &propsRes)
		c.JSON(http.StatusOK, propsRes)
	})

	// Endpoints de Contratos e Carnê no Frontend
	r.POST("/painel/contratos/novo", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		rent, _ := strconv.ParseFloat(c.PostForm("rent_amount"), 64)
		day, _ := strconv.Atoi(c.PostForm("due_day"))
		duration, _ := strconv.Atoi(c.PostForm("duration_months"))

		payload := map[string]interface{}{
			"property_id":     c.PostForm("property_id"),
			"tenant_name":     c.PostForm("tenant_name"),
			"tenant_phone":    c.PostForm("tenant_phone"),
			"tenant_cpf":      c.PostForm("tenant_cpf"),
			"rent_amount":     rent,
			"due_day":         day,
			"duration_months": duration,
		}

		var apiRes map[string]interface{}
		err := apiClient.Post("/api/contracts", payload, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar contrato"})
			return
		}
		c.JSON(http.StatusCreated, apiRes)
	})

	r.GET("/painel/contratos/:id/carne", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		contractID := c.Param("id")
		var apiRes struct {
			Contract     ContractItem      `json:"contract"`
			Installments []InstallmentItem `json:"installments"`
			ManualFee    float64           `json:"manual_fee"`
		}
		err := apiClient.Get("/api/contracts/"+contractID, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar carnê"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	r.POST("/painel/contratos/installments/:id/baixar-manual", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		instID := c.Param("id")
		var apiRes map[string]interface{}
		err := apiClient.Post("/api/contracts/installments/"+instID+"/confirm-manual-payment", nil, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao baixar"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	r.POST("/painel/contratos/installments/:id/reemitir", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		instID := c.Param("id")
		payload := map[string]interface{}{
			"new_due_date": c.PostForm("new_due_date"),
		}
		var apiRes map[string]interface{}
		err := apiClient.Post("/api/contracts/installments/"+instID+"/reissue", payload, token, &apiRes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao reemitir"})
			return
		}
		c.JSON(http.StatusOK, apiRes)
	})

	// Handlers de Perfil e Senha
	r.PUT("/painel/perfil/atualizar", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		payload := map[string]interface{}{
			"name":     c.PostForm("name"),
			"phone":    c.PostForm("phone"),
			"cpf_cnpj": c.PostForm("cpf_cnpj"),
		}
		err := apiClient.Put("/api/me/profile", payload, token, nil)
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Erro ao atualizar dados.</div>`))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-success">Dados pessoais atualizados com sucesso!</div>`))
	})

	r.POST("/painel/perfil/senha", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		payload := map[string]interface{}{
			"current_password": c.PostForm("current_password"),
			"new_password":     c.PostForm("new_password"),
		}
		err := apiClient.Post("/api/me/password", payload, token, nil)
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-danger">Senha atual incorreta ou erro no servidor.</div>`))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<div class="alert alert-success">Senha alterada com sucesso!</div>`))
	})

	r.POST("/painel/imoveis/salvar", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		id := c.PostForm("id")
		price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
		area, _ := strconv.ParseFloat(c.PostForm("area_m2"), 64)
		photos := c.PostForm("photos")

		payload := map[string]interface{}{
			"title":           c.PostForm("title"),
			"type":            c.PostForm("type"),
			"purpose":         c.PostForm("purpose"),
			"price":           price,
			"area_m2":         area,
			"city":            c.PostForm("city"),
			"state":           c.PostForm("state"),
			"cover_image_url": photos,
		}

		var apiRes struct {
			ID    string `json:"id"`
			Error string `json:"error"`
		}

		var err error
		if id == "" {
			err = apiClient.Post("/api/properties", payload, token, &apiRes)
		} else {
			err = apiClient.Patch("/api/properties/"+id, payload, token, &apiRes)
		}

		if err != nil || apiRes.ID == "" {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusOK)
	})

	r.DELETE("/painel/imoveis/:id", func(c *gin.Context) {
		token, _ := c.Cookie("access_token")
		id := c.Param("id")

		var apiRes struct {
			Message string `json:"message"`
		}
		_ = apiClient.Delete("/api/properties/"+id, token, &apiRes)

		c.Header("HX-Refresh", "true")
		c.Status(http.StatusOK)
	})

	// Páginas de Cadastro e Login
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
		c.Header("HX-Redirect", "/painel")
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
		c.Header("HX-Redirect", "/painel")
		c.Status(http.StatusOK)
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	})

	slog.Info("Servidor Frontend com Layout Dashboard e Sidebar Lateral", "porta", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Falha ao rodar servidor web", "err", err)
	}
}
