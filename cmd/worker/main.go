package main

import (
	"fmt"
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/internal/worker"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/telemetry"
	"log"

	"github.com/hibiken/asynq"
)

func main() {
	// Load Config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init Logger
	logger := logger.NewLogger()
	logger.Info("Starting Asynq Worker...")

	// Init Redis Config for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	// Init Opentelemetry
	telemetry.InitTracer(cfg, cfg.WorkerName)

	// Init Asynq Server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: logger,
		},
	)

	// Init Dependencies
	telegramService := telegram.NewTelegramService(cfg.TelegramToken, cfg.TelegramBaseURL)
	telegramHandler := worker.NewTelegramTaskHandler(logger, telegramService)

	// Register Tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeTelegramMessage, telegramHandler.HandleTelegramMessageTask)

	// Run Worker
	logger.Info("Worker server starting...")
	if err := srv.Run(mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
