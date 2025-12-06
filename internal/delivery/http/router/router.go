package router

import (
	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/delivery/http/handler"
	"go-gin-sqlx-template/internal/delivery/http/middleware"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/utils"
	"time"

	ginzap "github.com/gin-contrib/zap"
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
	// Apply global middleware
	r.engine.Use(middleware.CORS())

	// Add Zap middleware
	zapLogger := r.logger.GetZapLogger()
	r.engine.Use(middleware.ZapMiddleware(zapLogger))
	r.engine.Use(ginzap.RecoveryWithZap(zapLogger, true))

	// Add OpenTelemetry middleware
	r.engine.Use(otelgin.Middleware(r.cfg.ServiceName))

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
			users.GET("/:id", r.userHandler.GetUserByID, middleware.CacheMiddleware(r.redisClient, 1*time.Minute))
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}
	}

	// 404 Handler
	r.engine.NoRoute(func(c *gin.Context) {
		utils.ErrorResponse(c, 404, "you are lost", nil)
	})

	return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
	if err := r.db.HealthCheck(); err != nil {
		utils.ErrorResponse(c, 503, "Database connection failed", err)
		return
	}
	if err := r.redisClient.Client.Ping(c.Request.Context()).Err(); err != nil {
		utils.ErrorResponse(c, 503, "Redis connection failed", err)
		return
	}

	utils.SuccessResponse(c, 200, "Service is healthy", gin.H{
		"status":   "ok",
		"database": "connected",
		"redis":    "connected",
	})
}
