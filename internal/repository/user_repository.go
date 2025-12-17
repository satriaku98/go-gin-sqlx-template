package repository

import (
	"context"
	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/pkg/utils"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetAll(ctx context.Context, pagination utils.PaginationParams, filters utils.FilterParams, sort []utils.SortParams) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context, filters utils.FilterParams) (int64, error)
}
