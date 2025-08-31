package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/wukong0111/go-banks/internal/auth"
	"github.com/wukong0111/go-banks/internal/config"
	"github.com/wukong0111/go-banks/internal/handlers"
	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/middleware"
	"github.com/wukong0111/go-banks/internal/repository"
	"github.com/wukong0111/go-banks/internal/services"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger
	loggerCfg := &logger.Config{
		Level:       cfg.Logger.Level,
		Outputs:     cfg.Logger.Outputs,
		JSONFile:    cfg.Logger.JSONFile,
		TextFile:    cfg.Logger.TextFile,
		AddSource:   cfg.Logger.AddSource,
		MaxFileSize: cfg.Logger.MaxFileSize,
	}

	appLogger, err := logger.SetupLogger(loggerCfg)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	appLogger.Info("application starting",
		"version", "v1.0.0",
		"port", cfg.Port,
		"log_level", cfg.Logger.Level,
		"log_outputs", cfg.Logger.Outputs,
	)

	// Connect to database
	dbPool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("database connected successfully")

	// Initialize dependencies
	bankRepo := repository.NewPostgresBankRepository(dbPool)
	bankService := services.NewBankService(bankRepo)
	bankHandler := handlers.NewBankHandler(bankService)

	// Initialize bank creation dependencies
	bankWriter := repository.NewPostgresBankWriter(dbPool)
	bankCreatorService := services.NewBankCreatorService(bankWriter)
	bankCreatorHandler := handlers.NewBankCreatorHandler(bankCreatorService)

	// Initialize bank update dependencies
	bankUpdaterService := services.NewBankUpdaterService(bankWriter, bankRepo)
	bankUpdaterHandler := handlers.NewBankUpdaterHandler(bankUpdaterService)

	// Initialize bank filters dependencies
	bankFiltersService := services.NewBankFiltersService(bankRepo)
	bankFiltersHandler := handlers.NewBankFiltersHandler(bankFiltersService)

	// Initialize bank group dependencies
	bankGroupRepo := repository.NewPostgresBankGroupRepository(dbPool)
	bankGroupService := services.NewBankGroupService(bankGroupRepo)
	bankGroupWriter := repository.NewPostgresBankGroupWriter(dbPool)
	bankGroupCreatorService := services.NewBankGroupCreatorService(bankGroupWriter)
	bankGroupUpdaterService := services.NewBankGroupUpdaterService(bankGroupWriter, bankGroupRepo)
	bankGroupHandler := handlers.NewBankGroupHandler(bankGroupService, bankGroupCreatorService, bankGroupUpdaterService)

	// Initialize JWT service and auth middleware
	jwtExpiry, err := time.ParseDuration(cfg.JWT.Expiry)
	if err != nil {
		return fmt.Errorf("invalid JWT expiry duration: %w", err)
	}
	jwtService := auth.NewJWTService(cfg.JWT.Secret, jwtExpiry, appLogger)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup Gin router
	r := gin.Default()

	// Add request logging middleware
	r.Use(logger.RequestLogger(appLogger))

	// Health and readiness endpoints
	r.GET("/health", healthHandler())
	r.GET("/ready", readinessHandler(dbPool))

	// API routes with authentication
	api := r.Group("/api")
	// Bank endpoints require banks:read permission
	api.GET("/banks",
		authMiddleware.RequireAuth("banks:read"),
		bankHandler.GetBanks)
	api.GET("/banks/:bankId/details",
		authMiddleware.RequireAuth("banks:read"),
		bankHandler.GetBankDetails)
	// Bank creation endpoint requires banks:write permission
	api.POST("/banks",
		authMiddleware.RequireAuth("banks:write"),
		bankCreatorHandler.CreateBank)
	// Bank update endpoint requires banks:write permission
	api.PUT("/banks/:bankId",
		authMiddleware.RequireAuth("banks:write"),
		bankUpdaterHandler.UpdateBank)
	// Bank filters endpoint requires banks:read permission
	api.GET("/filters",
		authMiddleware.RequireAuth("banks:read"),
		bankFiltersHandler.GetFilters)
	// Bank groups endpoints
	api.GET("/bank-groups",
		authMiddleware.RequireAuth("banks:read"),
		bankGroupHandler.GetBankGroups)
	api.POST("/bank-groups",
		authMiddleware.RequireAuth("banks:write"),
		bankGroupHandler.CreateBankGroup)
	api.PUT("/bank-groups/:groupId",
		authMiddleware.RequireAuth("banks:write"),
		bankGroupHandler.UpdateBankGroup)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("HTTP server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("server startup failed", "error", err)
		}
	}()

	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-quit
	appLogger.Info("shutdown signal received", "signal", sig.String())

	// Create context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server gracefully
	appLogger.Info("initiating graceful shutdown")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("forced server shutdown", "error", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	appLogger.Info("graceful shutdown completed successfully")
	return nil
}

// healthHandler returns a basic health check endpoint
func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// readinessHandler returns a readiness probe that checks database connectivity
func readinessHandler(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Check database connection
		if err := dbPool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"reason":    "database connection failed",
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"database":  "connected",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
