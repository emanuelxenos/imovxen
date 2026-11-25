package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/config"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/properties"
	"github.com/imovxen/backend/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	// Conexão com o pool de banco de dados
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Erro fatal ao configurar banco de dados", "err", err)
	}
	defer database.Close()

	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTTTLMinutes, cfg.RefreshTokenTTLDays)
	storageService := storage.NewStorageService(cfg)

	r := gin.Default()

	// CORS configurado para o frontend SSR
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rota de Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "imovxen-backend",
		})
	})

	api := r.Group("/api")
	{
		// Auth Endpoints
		authGroup := api.Group("/auth")
		{
			// Cadastro de Usuário
			authGroup.POST("/register", func(c *gin.Context) {
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

				passHash, err := auth.HashPassword(req.Password)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
					return
				}

				var phonePtr *string
				if req.Phone != "" {
					phonePtr = &req.Phone
				}

				user, err := database.Queries.CreateUser(c.Request.Context(), db.CreateUserParams{
					Email:        req.Email,
					PasswordHash: passHash,
					Name:         req.Name,
					Phone:        phonePtr,
					Role:         "USER",
					PlanType:     "FREE",
					ListingQuota: 2,
				})
				if err != nil {
					c.JSON(http.StatusConflict, gin.H{"error": "E-mail já cadastrado ou erro ao criar conta"})
					return
				}

				accessToken, err := tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao emitir token"})
					return
				}

				rawRefresh, hashRefresh, expiresAt, _ := auth.GenerateSecureRefreshToken()
				database.Queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
					UserID:    user.ID,
					TokenHash: hashRefresh,
					ExpiresAt: expiresAt,
				})

				c.JSON(http.StatusCreated, gin.H{
					"id":            user.ID,
					"access_token":  accessToken,
					"refresh_token": rawRefresh,
					"user": gin.H{
						"id":           user.ID,
						"name":         user.Name,
						"email":        user.Email,
						"role":         user.Role,
						"plan_type":    user.PlanType,
						"listing_quota": user.ListingQuota,
						"listings_used": user.ListingsUsed,
					},
				})
			})

			// Login de Usuário
			authGroup.POST("/login", func(c *gin.Context) {
				var req struct {
					Email    string `json:"email" binding:"required,email"`
					Password string `json:"password" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				user, err := database.Queries.GetUserByEmail(c.Request.Context(), req.Email)
				if err != nil || !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "E-mail ou senha incorretos"})
					return
				}

				accessToken, err := tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao emitir token"})
					return
				}

				rawRefresh, hashRefresh, expiresAt, _ := auth.GenerateSecureRefreshToken()
				database.Queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
					UserID:    user.ID,
					TokenHash: hashRefresh,
					ExpiresAt: expiresAt,
				})

				c.JSON(http.StatusOK, gin.H{
					"access_token":  accessToken,
					"refresh_token": rawRefresh,
					"user": gin.H{
						"id":           user.ID,
						"name":         user.Name,
						"email":        user.Email,
						"role":         user.Role,
						"listing_quota": user.ListingQuota,
						"listings_used": user.ListingsUsed,
					},
				})
			})

			// Rotação de Refresh Token
			authGroup.POST("/refresh", func(c *gin.Context) {
				var req struct {
					RefreshToken string `json:"refresh_token" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				tokenHash := auth.HashRawRefreshToken(req.RefreshToken)
				savedToken, err := database.Queries.GetRefreshToken(c.Request.Context(), tokenHash)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expirado ou inválido"})
					return
				}

				user, err := database.Queries.GetUserByID(c.Request.Context(), savedToken.UserID)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
					return
				}

				// Rotacionar token antigo
				database.Queries.RevokeRefreshToken(c.Request.Context(), tokenHash)

				newAccessToken, _ := tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
				newRawRefresh, newHashRefresh, expiresAt, _ := auth.GenerateSecureRefreshToken()
				database.Queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
					UserID:    user.ID,
					TokenHash: newHashRefresh,
					ExpiresAt: expiresAt,
				})

				c.JSON(http.StatusOK, gin.H{
					"access_token":  newAccessToken,
					"refresh_token": newRawRefresh,
				})
			})
		}

		// Rota protegida: Dados do Usuário e Cota
		api.GET("/me", auth.AuthMiddleware(tokenService), func(c *gin.Context) {
			userID, _ := auth.GetContextUserID(c)
			user, err := database.Queries.GetUserByID(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"user": gin.H{
					"id":    user.ID,
					"name":  user.Name,
					"email": user.Email,
					"role":  user.Role,
				},
				"quota": gin.H{
					"used":  user.ListingsUsed,
					"limit": user.ListingQuota,
				},
			})
		})

		// Criação de Imóvel com Validação Atômica de Cota
		api.POST("/properties", auth.AuthMiddleware(tokenService), func(c *gin.Context) {
			userID, _ := auth.GetContextUserID(c)
			user, err := database.Queries.GetUserByID(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
				return
			}

			// Regra de Cota
			if user.ListingsUsed >= user.ListingQuota {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "quota_exceeded",
					"message": "Você atingiu o limite de anúncios ativos do seu plano. Faça upgrade para continuar.",
				})
				return
			}

			var req properties.CreatePropertyDTO
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			prop, err := database.Queries.CreateProperty(c.Request.Context(), db.CreatePropertyParams{
				UserID:        userID,
				Title:         req.Title,
				Description:   req.Description,
				Type:          req.Type,
				Purpose:       req.Purpose,
				Status:        "ACTIVE",
				Price:         req.Price,
				CondoFee:      req.CondoFee,
				Iptu:          req.IPTU,
				AreaM2:        req.AreaM2,
				Bedrooms:      req.Bedrooms,
				Bathrooms:     req.Bathrooms,
				ParkingSpots:  req.ParkingSpots,
				ZipCode:       req.ZipCode,
				Street:        req.Street,
				Number:        req.Number,
				Complement:    req.Complement,
				Neighborhood:  req.Neighborhood,
				City:          req.City,
				State:         req.State,
				Latitude:      req.Latitude,
				Longitude:     req.Longitude,
				CoverImageUrl: req.CoverImageURL,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar imóvel no banco de dados"})
				return
			}

			c.JSON(http.StatusCreated, prop)
		})

		// Listar meus imóveis
		api.GET("/me/properties", auth.AuthMiddleware(tokenService), func(c *gin.Context) {
			userID, _ := auth.GetContextUserID(c)
			props, err := database.Queries.ListPropertiesByUser(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar anúncios"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": props})
		})

		// Presign de Imagens (Upload direto ao storage)
		api.POST("/properties/:id/images/presign", auth.AuthMiddleware(tokenService), func(c *gin.Context) {
			propertyID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
				return
			}

			var req properties.PresignImageDTO
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			presignRes, err := storageService.GeneratePresignedUploadURL(c.Request.Context(), propertyID, req.Filename, req.MimeType)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, presignRes)
		})
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("Servidor Backend Iniciado", "porta", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Erro ao rodar servidor", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
