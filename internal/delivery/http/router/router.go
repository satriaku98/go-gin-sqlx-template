package router

import (
	"go-gin-sqlx-template/internal/delivery/http/handler"
	"go-gin-sqlx-template/internal/delivery/http/middleware"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/utils"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine      *gin.Engine
	userHandler *handler.UserHandler
	logger      *logger.Logger
	db          *database.Database
}

func NewRouter(
	userHandler *handler.UserHandler,
	logger *logger.Logger,
	db *database.Database,
) *Router {
	return &Router{
		engine:      gin.New(),
		userHandler: userHandler,
		logger:      logger,
		db:          db,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Apply global middleware
	r.engine.Use(middleware.CORS())

	// Add Zap middleware
	zapLogger := r.logger.GetZapLogger()
	r.engine.Use(middleware.ZapMiddleware(zapLogger))
	r.engine.Use(ginzap.RecoveryWithZap(zapLogger, true))

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
			users.GET("/:id", r.userHandler.GetUserByID)
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}
	}

	return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
	if err := r.db.HealthCheck(); err != nil {
		utils.ErrorResponse(c, 503, "Database connection failed", err)
		return
	}

	utils.SuccessResponse(c, 200, "Service is healthy", gin.H{
		"status":   "ok",
		"database": "connected",
	})
}
