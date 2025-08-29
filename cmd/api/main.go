package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// Initialize JWT service and auth middleware
	jwtExpiry, err := time.ParseDuration(cfg.JWT.Expiry)
	if err != nil {
		return fmt.Errorf("invalid JWT expiry duration: %w", err)
	}
	jwtService := auth.NewJWTService(cfg.JWT.Secret, jwtExpiry)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes with authentication
	api := r.Group("/api")
	// Bank endpoints require banks:read permission
	api.GET("/banks",
		authMiddleware.RequireAuth("banks:read"),
		bankHandler.GetBanks)
	api.GET("/banks/:bankId/details",
		authMiddleware.RequireAuth("banks:read"),
		bankHandler.GetBankDetails)

	log.Printf("Server starting on :%d", cfg.Port)
	if err := r.Run(":8080"); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
