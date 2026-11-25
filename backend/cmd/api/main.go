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
	"github.com/imovxen/backend/internal/properties"
	"github.com/imovxen/backend/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
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

	// Grupo da API
	api := r.Group("/api")
	{
		// Auth endpoints públicos
		authGroup := api.Group("/auth")
		{
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

				newUserID := uuid.New()
				accessToken, _ := tokenService.GenerateAccessToken(newUserID, req.Email, "USER")
				rawRefresh, _, _, _ := auth.GenerateSecureRefreshToken()

				c.JSON(http.StatusCreated, gin.H{
					"id":            newUserID,
					"access_token":  accessToken,
					"refresh_token": rawRefresh,
					"user": gin.H{
						"id":    newUserID,
						"name":  req.Name,
						"email": req.Email,
						"role":  "USER",
					},
				})
			})

			authGroup.POST("/login", func(c *gin.Context) {
				var req struct {
					Email    string `json:"email" binding:"required,email"`
					Password string `json:"password" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				mockUserID := uuid.New()
				accessToken, _ := tokenService.GenerateAccessToken(mockUserID, req.Email, "USER")
				rawRefresh, _, _, _ := auth.GenerateSecureRefreshToken()

				c.JSON(http.StatusOK, gin.H{
					"access_token":  accessToken,
					"refresh_token": rawRefresh,
					"user": gin.H{
						"id":    mockUserID,
						"email": req.Email,
						"name":  "Usuário Teste",
						"role":  "USER",
					},
				})
			})
		}

		// Rota protegida: Me
		api.GET("/me", auth.AuthMiddleware(tokenService), func(c *gin.Context) {
			userID, _ := auth.GetContextUserID(c)
			email, _ := c.Get(auth.ContextEmailKey)
			role, _ := c.Get(auth.ContextRoleKey)

			c.JSON(http.StatusOK, gin.H{
				"user": gin.H{
					"id":    userID,
					"email": email,
					"role":  role,
				},
				"quota": gin.H{
					"used":  0,
					"limit": 2,
				},
			})
		})

		// Imóveis públicos
		api.GET("/properties", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"data":      []gin.H{},
				"page":      1,
				"page_size": 20,
				"total":     0,
			})
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
