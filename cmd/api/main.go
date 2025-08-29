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

	// Initialize dependencies
	bankRepo := repository.NewPostgresBankRepository(dbPool)
	bankService := services.NewBankService(bankRepo)
	bankHandler := handlers.NewBankHandler(bankService)

	// Initialize bank creation dependencies
	bankWriter := repository.NewPostgresBankWriter(dbPool)
	bankCreatorService := services.NewBankCreatorService(bankWriter)
	bankCreatorHandler := handlers.NewBankCreatorHandler(bankCreatorService)

	// Initialize JWT service and auth middleware
	jwtExpiry, err := time.ParseDuration(cfg.JWT.Expiry)
	if err != nil {
		return fmt.Errorf("invalid JWT expiry duration: %w", err)
	}
	jwtService := auth.NewJWTService(cfg.JWT.Secret, jwtExpiry)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup Gin router
	r := gin.Default()

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
		log.Printf("Server starting on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-quit
	log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

	// Create context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server gracefully
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server shutdown completed gracefully")
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
