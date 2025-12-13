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
	"go-gin-sqlx-template/pkg/telemetry"

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
	log.Info(context.Background(), "Starting application...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf(context.Background(), "Failed to load config: %v", err)
	}
	log.Info(context.Background(), "Configuration loaded successfully")

	// Initialize database
	db, err := database.NewPostgresDatabase(cfg)
	if err != nil {
		log.Fatalf(context.Background(), "Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Info(context.Background(), "Database connected successfully")

	// Initialize dependency injection container
	container := NewContainer(cfg, log, db)
	log.Info(context.Background(), "Dependencies initialized successfully")

	// Initialize OpenTelemetry Tracer
	shutdown, err := telemetry.InitTracer(cfg, cfg.ServiceName)
	if err != nil {
		log.Fatalf(context.Background(), "Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Errorf(context.Background(), "Failed to shutdown tracer: %v", err)
		}
	}()
	log.Info(context.Background(), "Tracer initialized successfully")

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
		log.Infof(context.Background(), "Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(context.Background(), "Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf(ctx, "Server forced to shutdown: %v", err)
	}

	log.Info(ctx, "Server exited gracefully")
}
