package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imovxen/backend/internal/db"
)

type Handler struct {
	db           *db.Database
	tokenService *TokenService
}

func NewHandler(database *db.Database, tokenService *TokenService) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passHash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	var phonePtr *string
	if req.Phone != "" {
		phonePtr = &req.Phone
	}

	user, err := h.db.Queries.CreateUser(c.Request.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: passHash,
		Name:         req.Name,
		Phone:        phonePtr,
		Role:         "USER",
		PlanType:     "FREE",
		ListingQuota: 2,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "E-mail já cadastrado"})
		return
	}

	accessToken, _ := h.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
	rawRefresh, hashRefresh, expiresAt, _ := GenerateSecureRefreshToken()
	_, _ = h.db.Queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: expiresAt,
	})

	c.JSON(http.StatusCreated, gin.H{
		"id":            user.ID,
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"user": gin.H{
			"id":            user.ID,
			"name":          user.Name,
			"email":         user.Email,
			"phone":         user.Phone,
			"cpf_cnpj":      user.CpfCnpj,
			"role":          user.Role,
			"listing_quota": user.ListingQuota,
			"listings_used": user.ListingsUsed,
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.Queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil || !CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-mail ou senha incorretos"})
		return
	}

	accessToken, _ := h.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
	rawRefresh, hashRefresh, expiresAt, _ := GenerateSecureRefreshToken()
	_, _ = h.db.Queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: expiresAt,
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"user": gin.H{
			"id":            user.ID,
			"name":          user.Name,
			"email":         user.Email,
			"phone":         user.Phone,
			"cpf_cnpj":      user.CpfCnpj,
			"role":          user.Role,
			"listing_quota": user.ListingQuota,
			"listings_used": user.ListingsUsed,
		},
	})
}
