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
	"github.com/imovxen/backend/internal/mailer"
	"github.com/imovxen/backend/internal/payments"
)

type Handler struct {
	db           *db.Database
	tokenService *auth.TokenService
	asaas        *payments.AsaasClient
	mailer       *mailer.Mailer
}

func NewHandler(database *db.Database, tokenService *auth.TokenService, asaasKey string, isProd bool, mailService *mailer.Mailer) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
		asaas:        payments.NewAsaasClient(asaasKey, isProd),
		mailer:       mailService,
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

// SearchUsers busca usuários cadastrados no sistema por nome, e-mail ou telefone para vincular como inquilino
func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if len(q) < 2 {
		c.JSON(http.StatusOK, gin.H{"users": []db.UserSearchItem{}})
		return
	}

	users, err := h.db.Queries.SearchUsersByEmailOrName(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

type CreateContractDTO struct {
	PropertyID     string  `json:"property_id" binding:"required"`
	TenantID       string  `json:"tenant_id" binding:"required"` // ID relacional de um usuário do sistema
	RentAmount     float64 `json:"rent_amount" binding:"required,gt=0"`
	DueDay         int32   `json:"due_day" binding:"required,min=1,max=31"`
	DurationMonths int32   `json:"duration_months" binding:"required,min=1,max=60"`
	StartsAt       string  `json:"starts_at"`
}

// CreateContract cria o contrato vinculando o inquilino cadastrado ao imóvel e gera o carnê
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

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de inquilino inválido"})
		return
	}

	// Impedir vincular a si mesmo
	if tenantUUID == landlordID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O proprietário não pode ser inquilino de seu próprio imóvel"})
		return
	}

	// Verificar se o inquilino existe no sistema
	tenant, err := h.db.Queries.GetUserByID(c.Request.Context(), tenantUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inquilino selecionado não encontrado no sistema"})
		return
	}

	// Verificar se o imóvel pertence ao locador
	prop, err := h.db.Queries.GetPropertyByID(c.Request.Context(), propID)
	if err != nil || prop.UserID != landlordID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Imóvel não pertence a você"})
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
		TenantID:       tenantUUID,
		RentAmount:     req.RentAmount,
		DueDay:         req.DueDay,
		DurationMonths: req.DurationMonths,
		StartsAt:       startDate,
		Status:         "ACTIVE",
	})
	if err != nil {
		slog.Error("Erro ao criar contrato relacional", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar contrato"})
		return
	}

	// Gerar as parcelas mensais do carnê
	for i := int32(1); i <= req.DurationMonths; i++ {
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

	// Enviar e-mail automático ao inquilino avisando sobre o vínculo
	h.mailer.Send(tenant.Email, "📜 Novo Contrato de Locação Ativo - "+prop.Title, mailer.TemplateBase("Você foi vinculado a um imóvel", fmt.Sprintf(`
		<p>Olá, <strong>%s</strong>!</p>
		<p>Você foi vinculado como inquilino oficial do imóvel <strong>"%s"</strong>.</p>
		<div style="background: #f8fafc; border: 1px solid #e2e8f0; padding: 16px; border-radius: 8px; margin: 20px 0;">
			<p style="margin: 0 0 6px;"><strong>Valor do Aluguel:</strong> R$ %.2f / mês</p>
			<p style="margin: 0 0 6px;"><strong>Dia de Vencimento:</strong> Todo dia %d</p>
			<p style="margin: 0;"><strong>Duração:</strong> %d meses</p>
		</div>
		<p>Seu carnê digital mensal já foi gerado na plataforma.</p>
	`, tenant.Name, prop.Title, req.RentAmount, req.DueDay, req.DurationMonths)))

	c.JSON(http.StatusCreated, contract)
}

// ListMyContracts lista os contratos onde o usuário é proprietário
func (h *Handler) ListMyContracts(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	list, err := h.db.Queries.ListContractsByLandlord(c.Request.Context(), landlordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar contratos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetPropertyTenantsHistory exibe todo o histórico de inquilinos que já passaram por aquele imóvel
func (h *Handler) GetPropertyTenantsHistory(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	propertyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
		return
	}

	prop, err := h.db.Queries.GetPropertyByID(c.Request.Context(), propertyID)
	if err != nil || prop.UserID != landlordID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Imóvel não pertence ao usuário"})
		return
	}

	history, err := h.db.Queries.ListTenantsHistoryByProperty(c.Request.Context(), propertyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar histórico"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"property": prop,
		"history":  history,
	})
}

// FinalizeContract finaliza/encerra um contrato ativo, liberando o imóvel para um novo inquilino no histórico
func (h *Handler) FinalizeContract(c *gin.Context) {
	landlordID, _ := auth.GetContextUserID(c)
	contractID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de contrato inválido"})
		return
	}

	updated, err := h.db.Queries.FinalizeContract(c.Request.Context(), contractID, landlordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encerrar contrato"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract": updated,
		"message":  "Contrato encerrado e arquivado no histórico de inquilinos do imóvel.",
	})
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
		PaymentMethod string `json:"payment_method"`
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

	// Disparo automático do recibo por e-mail para proprietário e inquilino
	if landlord, err := h.db.Queries.GetUserByID(c.Request.Context(), landlordID); err == nil {
		h.mailer.SendRentalReceipt(landlord.Email, landlord.Name, inst.PropertyTitle, inst.InstallmentNumber, inst.Amount, inst.DueDate.Format("02/01/2006"), receiptCode, true)
	}
	if tenant, err := h.db.Queries.GetUserByID(c.Request.Context(), inst.TenantID); err == nil {
		h.mailer.SendRentalReceipt(tenant.Email, tenant.Name, inst.PropertyTitle, inst.InstallmentNumber, inst.Amount, inst.DueDate.Format("02/01/2006"), receiptCode, false)
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
