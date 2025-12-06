package main

import (
	"fmt"
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/internal/worker"
	"go-gin-sqlx-template/pkg/logger"
	"log"

	"github.com/hibiken/asynq"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Logger
	logger := logger.NewLogger()
	logger.Info("Starting Asynq Worker...")

	// 3. Init Redis Config for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	// 4. Init Asynq Server
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

	// 5. Init Dependencies
	telegramService := telegram.NewTelegramService(cfg.TelegramToken, cfg.TelegramBaseURL)
	telegramHandler := worker.NewTelegramTaskHandler(logger, telegramService)

	// 6. Register Tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeTelegramMessage, telegramHandler.HandleTelegramMessageTask)

	// 7. Run Worker
	logger.Info("Worker server starting...")
	if err := srv.Run(mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
