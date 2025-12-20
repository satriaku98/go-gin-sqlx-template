package handler

import (
	"net/http"
	"strconv"

	"go-gin-sqlx-template/internal/delivery/http/middleware"
	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/internal/usecase"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	"go-gin-sqlx-template/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
	redisClient *database.RedisClient
	logger      *logger.Logger
}

func NewUserHandler(userUsecase usecase.UserUsecase, redisClient *database.RedisClient, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with the input payload
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body model.CreateUserRequest true "Create User Request"
// @Success      201  {object}  utils.Response{data=model.UserResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.userUsecase.CreateUser(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User created successfully", user)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get user details by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  utils.Response{data=model.UserResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	user, err := h.userUsecase.GetUserByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Get all users with pagination and optional filters
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page   query     int     false  "Page number" default(1)
// @Param        limit  query     int     false  "Limit per page" default(10)
// @Param        name   query     string  false  "Filter by name (partial match)"
// @Param        email  query     string  false  "Filter by email (partial match)"
// @Success      200  {object}  utils.PaginationResponse{data=[]model.UserResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users [get]
var (
	// getAllUsersAllowedFilters defines which filters are allowed for GetAllUsers
	// only allow name and email
	getAllUsersAllowedFilters = []string{"name", "email"}

	// sort by id, email, name, created_at, updated_at
	// default sort by created_at desc
	// example: ?sort=id,desc&sort=name,asc
	getAllUsersAllowedSorts = map[string]string{
		"id":         "id",
		"email":      "email",
		"name":       "name",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	getAllUsersDefaultSorts = []utils.SortParams{
		{Field: "created_at", Direction: "desc"},
	}
)

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	pagination := utils.ParsePagination(c)

	// Parse sort parameters
	sort, err := utils.ParseSorts(c, getAllUsersAllowedSorts, getAllUsersDefaultSorts)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid sort parameters", err)
		return
	}

	// Parse filter parameters (only allow name and email)
	filters, err := utils.ParseFilters(c, getAllUsersAllowedFilters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Get users with pagination and filters
	users, total, err := h.userUsecase.GetAllUsers(
		c.Request.Context(),
		pagination,
		filters,
		sort,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get users", err)
		return
	}

	// Create pagination metadata
	paginationMeta := utils.CalculatePagination(pagination.Page, pagination.Limit, total)
	utils.PaginatedResponse(c, users, paginationMeta)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update user details by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "User ID"
// @Param        request  body      model.UpdateUserRequest  true  "Update User Request"
// @Success      200  {object}  utils.Response{data=model.UserResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.userUsecase.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	// Invalidate cache
	cacheKey := middleware.GetCacheKey(c)
	if err := h.redisClient.Client.Del(c.Request.Context(), cacheKey).Err(); err != nil {
		h.logger.Errorf(c.Request.Context(), "failed to delete cache: %v", err)
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	err = h.userUsecase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	// Invalidate cache
	cacheKey := middleware.GetCacheKey(c)
	if err := h.redisClient.Client.Del(c.Request.Context(), cacheKey).Err(); err != nil {
		h.logger.Errorf(c.Request.Context(), "failed to delete cache: %v", err)
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
