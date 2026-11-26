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
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/config"
	"github.com/imovxen/backend/internal/contracts"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/leads"
	"github.com/imovxen/backend/internal/mailer"
	"github.com/imovxen/backend/internal/payments"
	"github.com/imovxen/backend/internal/properties"
	"github.com/imovxen/backend/internal/storage"
	"github.com/imovxen/backend/internal/users"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Erro ao inicializar conexão com banco de dados", "err", err)
	}
	defer database.Close()

	// Pasta para uploads de arquivos reais
	_ = os.MkdirAll("./uploads", os.ModePerm)

	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTTTLMinutes, cfg.RefreshTokenTTLDays)
	storageService := storage.NewStorageService(cfg)
	_ = storageService
	mailService := mailer.NewMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.MailFrom)

	// Handlers de domínio
	authHandler := auth.NewHandler(database, tokenService)
	usersHandler := users.NewHandler(database, tokenService)
	propertiesHandler := properties.NewHandler(database, tokenService)
	leadsHandler := leads.NewHandler(database, tokenService, mailService)
	paymentsHandler := payments.NewHandler(database, tokenService, cfg.AsaasAPIKey, cfg.AsaasEnv == "production", mailService)
	contractsHandler := contracts.NewHandler(database, tokenService, cfg.AsaasAPIKey, cfg.AsaasEnv == "production", mailService)

	r := gin.Default()

	// Servir uploads de fotos estáticas
	r.Static("/uploads", "./uploads")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "imovxen-backend",
		})
	})

	api := r.Group("/api")
	{
		// Upload real com compressão Lanczos3
		api.POST("/upload", propertiesHandler.UploadPhotos)

		// Auth
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// Usuário & Perfil
		authMiddleware := auth.AuthMiddleware(tokenService)
		api.GET("/me", authMiddleware, usersHandler.GetMe)
		api.PUT("/me/profile", authMiddleware, usersHandler.UpdateProfile)
		api.POST("/me/password", authMiddleware, usersHandler.UpdatePassword)

		// Leads & Propostas
		api.GET("/me/leads", authMiddleware, leadsHandler.ListMyLeads)
		api.POST("/properties/:id/leads", leadsHandler.CreateLead)

		// Imóveis
		api.GET("/properties", propertiesHandler.ListProperties)
		api.GET("/properties/:id", propertiesHandler.GetProperty)
		api.POST("/properties", authMiddleware, propertiesHandler.CreateProperty)
		api.GET("/me/properties", authMiddleware, propertiesHandler.ListMyProperties)
		api.PATCH("/properties/:id", authMiddleware, propertiesHandler.UpdateProperty)
		api.DELETE("/properties/:id", authMiddleware, propertiesHandler.DeleteProperty)

		// Financeiro & Pagamentos Asaas
		api.GET("/plans/packages", paymentsHandler.GetPackages)
		api.POST("/payments/checkout", authMiddleware, paymentsHandler.CheckoutPack)
		api.GET("/me/payments", authMiddleware, paymentsHandler.ListMyPayments)
		api.GET("/payments/:id/status", authMiddleware, paymentsHandler.GetPaymentStatus)
		api.POST("/payments/webhook/asaas", paymentsHandler.WebhookAsaas)
		api.POST("/payments/:id/simulate-approval", authMiddleware, paymentsHandler.SimulatePaymentApproval)

		// Contratos & Carnê de Locação (Baixa Manual Escalonada R$ 3, 5, 7)
		api.POST("/contracts", authMiddleware, contractsHandler.CreateContract)
		api.GET("/me/contracts", authMiddleware, contractsHandler.ListMyContracts)
		api.GET("/contracts/:id", authMiddleware, contractsHandler.GetContractDetail)
		api.POST("/contracts/installments/:id/manual-fee-checkout", authMiddleware, contractsHandler.CheckoutManualFee)
		api.POST("/contracts/installments/:id/confirm-manual-payment", authMiddleware, contractsHandler.ConfirmManualPayment)
		api.POST("/contracts/installments/:id/reissue", authMiddleware, contractsHandler.ReissueInstallment)
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
