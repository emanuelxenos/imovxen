package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/db"
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

// GetMe retorna dados do usuário autenticado e uso de cota
func (h *Handler) GetMe(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	user, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"phone":    user.Phone,
			"cpf_cnpj": user.CpfCnpj,
			"role":     user.Role,
		},
		"quota": gin.H{
			"used":  user.ListingsUsed,
			"limit": user.ListingQuota,
		},
	})
}

// UpdateProfile atualiza dados cadastrais
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)

	var req struct {
		Name    string  `json:"name" binding:"required"`
		Phone   *string `json:"phone"`
		CpfCnpj *string `json:"cpf_cnpj"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.db.Queries.UpdateUserProfile(c.Request.Context(), db.UpdateUserProfileParams{
		ID:      userID,
		Name:    req.Name,
		Phone:   req.Phone,
		CpfCnpj: req.CpfCnpj,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar perfil"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// UpdatePassword altera a senha verificando hash anterior
func (h *Handler) UpdatePassword(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil || !auth.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha atual incorreta"})
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar nova senha"})
		return
	}

	err = h.db.Queries.UpdateUserPassword(c.Request.Context(), db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: newHash,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha no banco"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Senha alterada com sucesso!"})
}
