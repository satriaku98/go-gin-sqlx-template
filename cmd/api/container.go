package main

import (
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/delivery/http/handler"
	"go-gin-sqlx-template/internal/delivery/http/router"
	"go-gin-sqlx-template/internal/repository/postgres"
	"go-gin-sqlx-template/internal/usecase/impl"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
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
		log.Error("Failed to connect to Redis: %v", err)
	}

	// Repository layer
	userRepo := postgres.NewUserRepository(db.DB)

	// Usecase layer
	userUsecase := impl.NewUserUsecase(userRepo)

	// Handler layer
	userHandler := handler.NewUserHandler(userUsecase)

	// Router
	r := router.NewRouter(userHandler, log, db, redisClient)

	return &Container{
		Config:      cfg,
		Logger:      log,
		DB:          db,
		UserHandler: userHandler,
		Router:      r,
	}
}
