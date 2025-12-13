package router

import (
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/delivery/http/handler"
	"go-gin-sqlx-template/internal/delivery/http/middleware"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Router struct {
	engine      *gin.Engine
	userHandler *handler.UserHandler
	logger      *logger.Logger
	db          *database.Database
	redisClient *database.RedisClient
	cfg         config.Config
}

func NewRouter(
	userHandler *handler.UserHandler,
	logger *logger.Logger,
	db *database.Database,
	redisClient *database.RedisClient,
	cfg config.Config,
) *Router {
	return &Router{
		engine:      gin.New(),
		userHandler: userHandler,
		logger:      logger,
		db:          db,
		redisClient: redisClient,
		cfg:         cfg,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Add OpenTelemetry middleware FIRST to create span context
	r.engine.Use(otelgin.Middleware(r.cfg.ServiceName))

	// Apply global middleware
	r.engine.Use(middleware.Recovery(r.logger))
	r.engine.Use(middleware.RequestLogger(r.logger))

	// Health check endpoint
	r.engine.GET("/health", r.healthCheck)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", r.userHandler.CreateUser)
			users.GET("", r.userHandler.GetAllUsers)
			users.GET("/:id", middleware.CacheMiddleware(r.redisClient, 1*time.Minute, r.logger), r.userHandler.GetUserByID)
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}
	}

	// 404 Handler
	r.engine.NoRoute(func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotFound, "you are lost", nil)
	})

	return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
	if err := r.db.HealthCheck(); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Database connection failed", err)
		return
	}
	if err := r.redisClient.HealthCheck(c.Request.Context()); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Redis connection failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service is healthy", gin.H{
		"status":   "ok",
		"database": "connected",
		"redis":    "connected",
	})
}
