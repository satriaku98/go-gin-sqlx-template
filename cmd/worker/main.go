package main

import (
	"fmt"
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/internal/worker"
	"log"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Logger
	// Simple zap production logger
	prodConfig := zap.NewProductionConfig()
	prodConfig.DisableCaller = true
	logger, _ := prodConfig.Build()
	sugar := logger.Sugar()

	sugar.Info("Starting Asynq Worker...")

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
			Logger: sugar,
		},
	)

	// 5. Init Dependencies
	telegramService := telegram.NewTelegramService(cfg.TelegramToken, cfg.TelegramBaseURL)
	telegramHandler := worker.NewTelegramTaskHandler(sugar, telegramService)

	// 6. Register Tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeTelegramMessage, telegramHandler.HandleTelegramMessageTask)

	// 7. Run Worker
	sugar.Info("Worker server starting...")
	if err := srv.Run(mux); err != nil {
		sugar.Fatalf("could not run server: %v", err)
	}
}
