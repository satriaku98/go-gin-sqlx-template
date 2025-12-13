package impl

import (
	"context"
	"fmt"

	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/internal/repository"
	"go-gin-sqlx-template/internal/usecase"
	"go-gin-sqlx-template/internal/worker"
	"go-gin-sqlx-template/pkg/logger"

	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo    repository.UserRepository
	asynqClient *asynq.Client
	config      config.Config
	logger      *logger.Logger
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	asynqClient *asynq.Client,
	cfg config.Config,
	log *logger.Logger,
) usecase.UserUsecase {
	return &userUsecase{
		userRepo:    userRepo,
		asynqClient: asynqClient,
		config:      cfg,
		logger:      log,
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

	// Send Telegram message with asynq task
	taskPayload := fmt.Sprintf("New user created: %s (%s)", user.Name, user.Email)
	task, _ := worker.NewTelegramMessageTask(ctx, u.config.TelegramChatID, taskPayload)
	if task != nil {
		// Enqueue task to be processed asynchronously
		info, err := u.asynqClient.Enqueue(task)
		if err != nil {
			u.logger.Errorf(ctx, "Failed to enqueue telegram task: %v", err)
		} else {

			u.logger.Infof(ctx, "Enqueued task: id=%s queue=%s", info.ID, info.Queue)
		}
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

	// Send Telegram message with asynq task
	taskPayload := fmt.Sprintf("User updated: %s (%s)", user.Name, user.Email)
	task, _ := worker.NewTelegramMessageTask(ctx, u.config.TelegramChatID, taskPayload)
	if task != nil {
		// Enqueue task to be processed asynchronously
		info, err := u.asynqClient.Enqueue(task)
		if err != nil {
			u.logger.Errorf(ctx, "Failed to enqueue telegram task: %v", err)
		} else {
			u.logger.Infof(ctx, "Enqueued task: id=%s queue=%s", info.ID, info.Queue)
		}
	}

	response := user.ToResponse()
	return &response, nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, id int64) error {
	return u.userRepo.Delete(ctx, id)
}
