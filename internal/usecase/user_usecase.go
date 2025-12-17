package usecase

import (
	"context"
	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/pkg/utils"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id int64) (*model.UserResponse, error)
	GetAllUsers(ctx context.Context, pagination utils.PaginationParams, filters utils.FilterParams, sort []utils.SortParams) ([]model.UserResponse, int64, error)
	UpdateUser(ctx context.Context, id int64, req model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id int64) error
}
