package main

import (
	"context"
	"fmt"
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/delivery/http/handler"
	"go-gin-sqlx-template/internal/delivery/http/router"
	"go-gin-sqlx-template/internal/repository/postgres"
	"go-gin-sqlx-template/internal/usecase/impl"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/pubsub"

	"github.com/hibiken/asynq"
)

// Container holds all application dependencies
type Container struct {
	Config      config.Config
	Logger      *logger.Logger
	DB          *database.Database
	UserHandler *handler.UserHandler
	Router      *router.Router
}

// NewContainer initializes all dependencies and wires them together
func NewContainer(cfg config.Config, log *logger.Logger, db *database.Database) *Container {
	// Initialize Redis
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Errorf(context.Background(), "Failed to connect to Redis: %v", err)
	}

	// Initialize PubSub Client
	pubsubClient, err := pubsub.NewClient(cfg)
	if err != nil {
		log.Errorf(context.Background(), "Failed to create pubsub client: %v", err)
	}

	// Ensure all topics and subscriptions exist
	if err := pubsubClient.EnsureAll(context.Background(), pubsub.GetTopicConfig(cfg)); err != nil {
		log.Errorf(context.Background(), "Failed to ensure pubsub topics and subscriptions: %v", err)
	}

	// Initialize Asynq Client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Repository layer
	userRepo := postgres.NewUserRepository(db.DB)

	// Usecase layer
	userUsecase := impl.NewUserUsecase(userRepo, asynqClient, pubsubClient, cfg, log)

	// Handler layer
	userHandler := handler.NewUserHandler(userUsecase, redisClient)

	// Router
	r := router.NewRouter(userHandler, log, db, redisClient, cfg)

	return &Container{
		Config:      cfg,
		Logger:      log,
		DB:          db,
		UserHandler: userHandler,
		Router:      r,
	}
}
