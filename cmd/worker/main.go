package main

import (
	"context"
	"fmt"
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/internal/worker"
	pubsubworker "go-gin-sqlx-template/internal/worker/pubsub"
	"go-gin-sqlx-template/pkg/logger"
	ps "go-gin-sqlx-template/pkg/pubsub"
	"go-gin-sqlx-template/pkg/telemetry"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load Config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init Logger
	loggerInstance := logger.NewLogger()
	loggerInstance.Info(ctx, "Starting Asynq Worker...")

	defer func() {
		if r := recover(); r != nil {
			loggerInstance.Error(context.Background(), "panic recovered in main", r)
		}
	}()

	// Init Redis Config for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	// Init Opentelemetry
	telemetry.InitTracer(cfg, cfg.WorkerName)

	// Init PubSub Worker
	pubsubClient := pubsubWorker(ctx, cfg, loggerInstance)

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
			Logger: logger.NewAsynqLoggerAdapter(loggerInstance),
		},
	)

	// Init Dependencies
	telegramService := telegram.NewTelegramService(cfg.TelegramToken, cfg.TelegramBaseURL)
	telegramHandler := worker.NewTelegramTaskHandler(loggerInstance, telegramService)

	// Register Tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeTelegramMessage, telegramHandler.HandleTelegramMessageTask)

	// Run Worker
	loggerInstance.Info(context.Background(), "Worker server starting...")

	go func() {
		if err := srv.Run(mux); err != nil {
			loggerInstance.Errorf(context.Background(), "asynq stopped: %v", err)
		}
	}()

	<-ctx.Done()
	loggerInstance.Info(context.Background(), "shutdown signal received")

	srv.Shutdown()
	if pubsubClient != nil {
		pubsubClient.Close()
	}
}

func pubsubWorker(ctx context.Context, cfg config.Config, loggerInstance *logger.Logger) *ps.Client {
	pubsubClient, err := ps.NewClient(cfg)
	if err != nil {
		loggerInstance.Fatalf(context.Background(), "Failed to create pubsub client: %v", err)
	}

	if err := pubsubClient.EnsureAll(context.Background(), ps.GetTopicConfig(cfg)); err != nil {
		loggerInstance.Fatalf(context.Background(), "Failed to ensure pubsub topics and subscriptions: %v", err)
	}

	worker := pubsubworker.New(pubsubClient, loggerInstance)

	worker.Start(
		ctx,
		worker.SubscribeUserCreated(ctx, cfg.PubSubSubscriptionUserCreated),
		// add more subscription here
	)

	return pubsubClient
}
