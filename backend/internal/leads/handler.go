package leads

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/properties"
)

type Handler struct {
	db           *db.Database
	tokenService *auth.TokenService
}

func NewHandler(database *db.Database, tokenService *auth.TokenService) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
	}
}

// ListMyLeads lista propostas recebidas pelos imóveis do usuário autenticado
func (h *Handler) ListMyLeads(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	leadList, err := h.db.Queries.ListLeadsForOwner(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar propostas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": leadList})
}

// CreateLead cria uma nova proposta para um imóvel
func (h *Handler) CreateLead(c *gin.Context) {
	propertyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
		return
	}

	var req properties.CreateLeadDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdLead, err := h.db.Queries.CreateLead(c.Request.Context(), db.CreateLeadParams{
		PropertyID: propertyID,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Message:    req.Message,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar proposta"})
		return
	}

	c.JSON(http.StatusCreated, createdLead)
}
