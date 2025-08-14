package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/config"
	"github.com/wukong0111/go-banks/internal/handlers"
	"github.com/wukong0111/go-banks/internal/repository"
	"github.com/wukong0111/go-banks/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbPool, err := pgxpool.New(context.Background(), cfg.Database.ConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize dependencies
	bankRepo := repository.NewPostgresBankRepository(dbPool)
	bankService := services.NewBankService(bankRepo)
	bankHandler := handlers.NewBankHandler(bankService)

	// Setup Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/banks", bankHandler.GetBanks)
	}

	log.Printf("Server starting on :%d", cfg.Port)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
