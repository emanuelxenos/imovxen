package contracts

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/payments"
)

type Handler struct {
	db           *db.Database
	tokenService *auth.TokenService
	asaas        *payments.AsaasClient
}

func NewHandler(database *db.Database, tokenService *auth.TokenService, asaasKey string, isProd bool) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
		asaas:        payments.NewAsaasClient(asaasKey, isProd),
	}
}

// CalculateManualFee calcula a taxa da baixa manual escalonada:
// <= 500: R$ 3,00 | <= 1500: R$ 5,00 | > 1500: R$ 7,00
func CalculateManualFee(rentAmount float64) float64 {
	if rentAmount <= 500.0 {
		return 3.00
	} else if rentAmount <= 1500.0 {
		return 5.00
	}
	return 7.00
}

type CreateContractDTO struct {
	PropertyID     string  `json:"property_id" binding:"required"`
	TenantName     string  `json:"tenant_name" binding:"required"`
	TenantEmail    *string `json:"tenant_email"`
	TenantPhone    *string `json:"tenant_phone"`
	TenantCpf      *string `json:"tenant_cpf"`
	RentAmount     float64 `json:"rent_amount" binding:"required,gt=0"`
	DueDay         int32   `json:"due_day" binding:"required,min=1,max=31"`
	DurationMonths int32   `json:"duration_months" binding:"required,min=1,max=60"`
	StartsAt       string  `json:"starts_at"`
}

// CreateContract cria o contrato e gera atomicamente as parcelas do carnê
func (h *Handler) CreateContract(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)

	var req CreateContractDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	propID, err := uuid.Parse(req.PropertyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
		return
	}

	startDate := time.Now()
	if req.StartsAt != "" {
		if t, err := time.Parse("2006-01-02", req.StartsAt); err == nil {
			startDate = t
		}
	}

	contract, err := h.db.Queries.CreateContract(c.Request.Context(), db.CreateContractParams{
		PropertyID:     propID,
		LandlordID:     landlordID,
		TenantName:     req.TenantName,
		TenantEmail:    req.TenantEmail,
		TenantPhone:    req.TenantPhone,
		TenantCpf:      req.TenantCpf,
		RentAmount:     req.RentAmount,
		DueDay:         req.DueDay,
		DurationMonths: req.DurationMonths,
		StartsAt:       startDate,
		Status:         "ACTIVE",
	})
	if err != nil {
		slog.Error("Erro ao criar contrato", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar contrato"})
		return
	}

	// Gerar as parcelas mensais do carnê
	for i := int32(1); i <= req.DurationMonths; i++ {
		// Ajusta o mês de vencimento
		monthDate := startDate.AddDate(0, int(i-1), 0)
		year, month, _ := monthDate.Date()
		dueDate := time.Date(year, month, int(req.DueDay), 0, 0, 0, 0, time.UTC)

		_, _ = h.db.Queries.CreateInstallment(c.Request.Context(), db.CreateInstallmentParams{
			ContractID:        contract.ID,
			InstallmentNumber: i,
			DueDate:           dueDate,
			Amount:            req.RentAmount,
			Status:            "PENDING",
		})
	}

	c.JSON(http.StatusCreated, contract)
}

// ListMyContracts lista os contratos do proprietário
func (h *Handler) ListMyContracts(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	list, err := h.db.Queries.ListContractsByLandlord(c.Request.Context(), landlordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar contratos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetContractDetail retorna o contrato com todas as parcelas do carnê
func (h *Handler) GetContractDetail(c *gin.Context) {
	contractID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de contrato inválido"})
		return
	}

	contract, err := h.db.Queries.GetContractByID(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contrato não encontrado"})
		return
	}

	installments, err := h.db.Queries.ListInstallmentsByContract(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar parcelas"})
		return
	}

	manualFee := CalculateManualFee(contract.RentAmount)

	c.JSON(http.StatusOK, gin.H{
		"contract":     contract,
		"installments": installments,
		"manual_fee":   manualFee,
	})
}

// CheckoutManualFee gera o Pix da taxa de serviço (R$ 3, 5 ou 7) para o proprietário dar baixa manual
func (h *Handler) CheckoutManualFee(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	installmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	inst, err := h.db.Queries.GetInstallmentByID(c.Request.Context(), installmentID)
	if err != nil || inst.LandlordID != landlordID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parcela não encontrada"})
		return
	}

	fee := CalculateManualFee(inst.RentAmount)
	user, _ := h.db.Queries.GetUserByID(c.Request.Context(), landlordID)

	customer, err := h.asaas.CreateCustomer(payments.AsaasCustomerRequest{
		Name:    user.Name,
		Email:   user.Email,
		CpfCnpj: func() string { if user.CpfCnpj != nil { return *user.CpfCnpj }; return "" }(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no gateway"})
		return
	}

	desc := fmt.Sprintf("Taxa de Baixa e Recibo - Parcela %d/%s", inst.InstallmentNumber, inst.PropertyTitle)
	payRes, pixRes, err := h.asaas.CreatePixPayment(customer.ID, fee, desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar Pix da taxa"})
		return
	}

	var pixCode, pixQR, gwID string
	if pixRes != nil {
		pixCode = pixRes.Payload
		pixQR = pixRes.EncodedImage
	}
	if payRes != nil {
		gwID = payRes.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"installment_id": inst.ID,
		"fee_amount":     fee,
		"gateway_id":     gwID,
		"pix_code":       pixCode,
		"pix_qr_base64":  pixQR,
	})
}

// ConfirmManualPayment efetua a baixa manual da parcela e gera o recibo oficial
func (h *Handler) ConfirmManualPayment(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	installmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	inst, err := h.db.Queries.GetInstallmentByID(c.Request.Context(), installmentID)
	if err != nil || inst.LandlordID != landlordID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parcela não encontrada"})
		return
	}

	var req struct {
		PaymentMethod string `json:"payment_method"` // "DINHEIRO", "PIX_DIRETO", "TED", "OUTRO"
		Notes         string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

	fee := CalculateManualFee(inst.RentAmount)
	receiptCode := fmt.Sprintf("REC-%d-%s", inst.InstallmentNumber, uuid.New().String()[:8])

	method := "PIX_DIRETO"
	if req.PaymentMethod != "" {
		method = req.PaymentMethod
	}

	notes := req.Notes

	updated, err := h.db.Queries.MarkInstallmentPaidManual(c.Request.Context(), db.MarkInstallmentPaidManualParams{
		ID:              inst.ID,
		ManualFeeAmount: fee,
		PaymentMethod:   &method,
		Notes:           &notes,
		ReceiptCode:     &receiptCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao registrar quitação manual"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"installment":  updated,
		"receipt_code": receiptCode,
		"message":      "Baixa manual realizada com sucesso! Recibo gerado.",
	})
}

// ReissueInstallment reemite a cobrança com nova data de vencimento
func (h *Handler) ReissueInstallment(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	installmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	inst, err := h.db.Queries.GetInstallmentByID(c.Request.Context(), installmentID)
	if err != nil || inst.LandlordID != landlordID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parcela não encontrada"})
		return
	}

	var req struct {
		NewDueDate string `json:"new_due_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nova data obrigatória"})
		return
	}

	newDue, err := time.Parse("2006-01-02", req.NewDueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido (YYYY-MM-DD)"})
		return
	}

	updated, err := h.db.Queries.UpdateInstallmentDueDate(c.Request.Context(), db.UpdateInstallmentDueDateParams{
		ID:        inst.ID,
		DueDate:   newDue,
		GatewayID: inst.GatewayID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar vencimento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"installment": updated,
		"message":     "Cobrança reemitida com sucesso para o novo vencimento!",
	})
}
