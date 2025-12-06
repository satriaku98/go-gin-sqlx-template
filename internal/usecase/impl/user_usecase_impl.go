package impl

import (
	"context"
	"fmt"

	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/internal/repository"
	"go-gin-sqlx-template/internal/usecase"

	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) usecase.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	// Check if email already exists
	existingUser, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (u *userUsecase) GetUserByID(ctx context.Context, id int64) (*model.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (u *userUsecase) GetAllUsers(ctx context.Context, page, limit int) ([]model.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, err := u.userRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, id int64, req model.UpdateUserRequest) (*model.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new email already exists (if email is being updated)
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := u.userRepo.GetByEmail(ctx, req.Email)
		if existingUser != nil {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, id int64) error {
	return u.userRepo.Delete(ctx, id)
}
