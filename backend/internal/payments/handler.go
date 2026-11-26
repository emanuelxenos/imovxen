package payments

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/mailer"
)

type Handler struct {
	db           *db.Database
	tokenService *auth.TokenService
	asaas        *AsaasClient
	mailer       *mailer.Mailer
}

func NewHandler(database *db.Database, tokenService *auth.TokenService, asaasKey string, isProd bool, mailService *mailer.Mailer) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
		asaas:        NewAsaasClient(asaasKey, isProd),
		mailer:       mailService,
	}
}

// PlanPackage define um pacote de anúncios disponível para compra
type PlanPackage struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Listings    int32   `json:"listings"`
	Price       float64 `json:"price"`
	PriceFormatted string `json:"price_formatted"`
}

var AvailablePackages = []PlanPackage{
	{
		ID:             "pack_basic",
		Title:          "Pacote Start",
		Description:    "Adiciona +5 anúncios ativos permanentes à sua conta",
		Listings:       5,
		Price:          49.90,
		PriceFormatted: "R$ 49,90",
	},
	{
		ID:             "pack_pro",
		Title:          "Pacote Corretor Pro",
		Description:    "Adiciona +15 anúncios ativos e destaque na vitrine",
		Listings:       15,
		Price:          99.90,
		PriceFormatted: "R$ 99,90",
	},
	{
		ID:             "pack_agency",
		Title:          "Pacote Imobiliária",
		Description:    "Adiciona +50 anúncios ativos com suporte prioritário",
		Listings:       50,
		Price:          249.90,
		PriceFormatted: "R$ 249,90",
	},
}

// GetPackages lista os pacotes de anúncios à venda
func (h *Handler) GetPackages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"packages": AvailablePackages,
	})
}

// CheckoutPack cria uma cobrança Pix no Asaas para o pacote escolhido
func (h *Handler) CheckoutPack(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)

	var req struct {
		PackageID string `json:"package_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pacote inválido"})
		return
	}

	var selected *PlanPackage
	for _, p := range AvailablePackages {
		if p.ID == req.PackageID {
			selected = &p
			break
		}
	}
	if selected == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pacote selecionado não encontrado"})
		return
	}

	user, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Criar ou localizar cliente no Asaas
	customer, err := h.asaas.CreateCustomer(AsaasCustomerRequest{
		Name:    user.Name,
		Email:   user.Email,
		CpfCnpj: func() string { if user.CpfCnpj != nil { return *user.CpfCnpj }; return "" }(),
		Phone:   func() string { if user.Phone != nil { return *user.Phone }; return "" }(),
	})
	if err != nil {
		slog.Error("Erro ao criar cliente no Asaas", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro de comunicação com o gateway de pagamento"})
		return
	}

	// Criar cobrança Pix
	payRes, pixRes, err := h.asaas.CreatePixPayment(customer.ID, selected.Price, "Imovxen - "+selected.Title)
	if err != nil {
		slog.Error("Erro ao gerar Pix no Asaas", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar cobrança Pix"})
		return
	}

	var gatewayID *string
	if payRes != nil && payRes.ID != "" {
		gatewayID = &payRes.ID
	}

	var pixCode, pixQR *string
	if pixRes != nil {
		pixCode = &pixRes.Payload
		pixQR = &pixRes.EncodedImage
	}

	// Registrar o pagamento no banco
	payment, err := h.db.Queries.CreatePayment(c.Request.Context(), db.CreatePaymentParams{
		UserID:      userID,
		Kind:        "LISTING_PACK",
		Amount:      selected.Price,
		ItemsCount:  selected.Listings,
		DueDate:     time.Now().AddDate(0, 0, 1),
		Status:      "PENDING",
		Method:      "PIX",
		GatewayID:   gatewayID,
		PixCode:     pixCode,
		PixQrBase64: pixQR,
	})
	if err != nil {
		slog.Error("Erro ao gravar pagamento no banco", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar fatura"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment": payment,
		"package": selected,
	})
}

// ListMyPayments lista todas as transações do usuário logado
func (h *Handler) ListMyPayments(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	list, err := h.db.Queries.ListPaymentsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar transações"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetPaymentStatus verifica se um pagamento específico já foi liquidado
func (h *Handler) GetPaymentStatus(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pagamento inválido"})
		return
	}

	payment, err := h.db.Queries.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          payment.ID,
		"status":      payment.Status,
		"items_count": payment.ItemsCount,
		"paid_at":     payment.PaidAt,
	})
}

// WebhookAsaas processa eventos de notificação do gateway Asaas
func (h *Handler) WebhookAsaas(c *gin.Context) {
	var payload struct {
		Event   string `json:"event"`
		Payment struct {
			ID          string  `json:"id"`
			Status      string  `json:"status"`
			Value       float64 `json:"value"`
			BillingType string  `json:"billingType"`
		} `json:"payment"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	slog.Info("Webhook Asaas recebido", "event", payload.Event, "payment_id", payload.Payment.ID)

	if payload.Event == "PAYMENT_RECEIVED" || payload.Event == "PAYMENT_CONFIRMED" {
		// Atualiza o pagamento para PAID — a trigger trg_process_payment_confirmation no Postgres
		// incrementa atomicamente o listing_quota do usuário!
		updated, err := h.db.Queries.UpdatePaymentStatusByGatewayID(c.Request.Context(), payload.Payment.ID, "PAID")
		if err != nil {
			slog.Warn("Pagamento do webhook não localizado no banco local", "gateway_id", payload.Payment.ID)
		} else {
			slog.Info("Cota creditada com sucesso via webhook", "payment_id", updated.ID, "items", updated.ItemsCount)
			// Disparar e-mail de confirmação ao anunciante
			if u, err := h.db.Queries.GetUserByID(c.Request.Context(), updated.UserID); err == nil {
				h.mailer.SendPaymentConfirmed(u.Email, u.Name, "Pacote de Anúncios", updated.ItemsCount, updated.Amount)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// SimulatePaymentApproval endpoint de apoio para testes locais simulando confirmação de Pix
func (h *Handler) SimulatePaymentApproval(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pagamento inválido"})
		return
	}

	updated, err := h.db.Queries.UpdatePaymentStatus(c.Request.Context(), paymentID, "PAID")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao aprovar pagamento"})
		return
	}

	// Disparar e-mail de confirmação
	if u, err := h.db.Queries.GetUserByID(c.Request.Context(), updated.UserID); err == nil {
		h.mailer.SendPaymentConfirmed(u.Email, u.Name, "Pacote de Anúncios", updated.ItemsCount, updated.Amount)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Pagamento aprovado e cota creditada automaticamente via trigger!",
		"payment_id":  updated.ID,
		"items_count": updated.ItemsCount,
		"status":      updated.Status,
	})
}
