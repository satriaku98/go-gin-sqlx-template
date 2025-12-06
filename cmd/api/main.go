package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gin-sqlx-template/config"
	_ "go-gin-sqlx-template/docs" // Import generated docs
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Go Gin SQLX Template API
// @version         1.0
// @description     This is a sample server for Go Gin SQLX Template.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting application...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}
	log.Info("Configuration loaded successfully")

	// Initialize database
	db, err := database.NewPostgresDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Info("Database connected successfully")

	// Initialize dependency injection container
	container := NewContainer(cfg, log, db)
	log.Info("Dependencies initialized successfully")

	// Setup router
	engine := container.Router.Setup()

	// Swagger setup
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: engine,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited gracefully")
}
