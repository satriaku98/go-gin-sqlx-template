package impl

import (
	"context"
	"fmt"

	"go-gin-sqlx-template/config"
	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/internal/repository"
	"go-gin-sqlx-template/internal/usecase"
	"go-gin-sqlx-template/internal/worker"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/logger"
	ps "go-gin-sqlx-template/pkg/pubsub"
	"go-gin-sqlx-template/pkg/utils"

	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
)

type userUsecase struct {
	userRepo     repository.UserRepository
	txManager    database.Transactor
	asynqClient  *asynq.Client
	pubsubClient *ps.Client
	config       config.Config
	logger       *logger.Logger
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	txManager database.Transactor,
	asynqClient *asynq.Client,
	pubsubClient *ps.Client,
	cfg config.Config,
	log *logger.Logger,
) usecase.UserUsecase {
	return &userUsecase{
		userRepo:     userRepo,
		txManager:    txManager,
		asynqClient:  asynqClient,
		pubsubClient: pubsubClient,
		config:       cfg,
		logger:       log,
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

	// NOTE:
	// This is only an example to show both Pub/Sub and Asynq usage.
	// In production, consider using only one of them based on your needs
	// to avoid duplicated async handling.

	// Send PubSub message
	message := fmt.Sprintf("New user created: %s (%s)", user.Name, user.Email)
	if id, err := u.pubsubClient.Publish(ctx, u.config.PubSubTopicUserCreated, []byte(message), nil); err != nil {
		u.logger.Errorf(ctx, "Failed to publish pubsub message: %v", err)
	} else {
		u.logger.Infof(ctx, "Published message: id=%s", id)
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

func (u *userUsecase) GetAllUsers(ctx context.Context, pagination utils.PaginationParams, filters utils.FilterParams, sort []utils.SortParams) ([]model.UserResponse, int64, error) {
	var (
		users []model.User
		total int64
		err   error
	)

	// Run GetAll and Count concurrently
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		users, err = u.userRepo.GetAll(ctx, pagination, filters, sort)
		return err
	})

	g.Go(func() error {
		total, err = u.userRepo.Count(ctx, filters)
		return err
	})

	if err := g.Wait(); err != nil {
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

// CreateUserWithTransaction is an example method demonstrating transaction usage
// This shows how to use the transaction manager when you need multiple repository
// operations to be atomic (all succeed or all fail together)
//
// Example use case: Creating a user and related data in multiple tables
// If any operation fails, all changes are rolled back automatically
func (u *userUsecase) CreateUserWithTransaction(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	var user *model.User

	// Execute all operations within a transaction
	err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if email already exists
		existingUser, _ := u.userRepo.GetByEmail(txCtx, req.Email)
		if existingUser != nil {
			return fmt.Errorf("email already exists")
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user = &model.User{
			Email:    req.Email,
			Name:     req.Name,
			Password: string(hashedPassword),
		}

		// Create user - this will use the transaction from txCtx
		if err := u.userRepo.Create(txCtx, user); err != nil {
			return err // Will trigger rollback
		}

		// If you had other repositories (e.g., ProfileRepository, AuditLogRepository),
		// you would call them here with the same txCtx:
		//
		// profile := &model.Profile{UserID: user.ID, ...}
		// if err := u.profileRepo.Create(txCtx, profile); err != nil {
		//     return err // Will rollback both user and profile creation
		// }
		//
		// auditLog := &model.AuditLog{Action: "user_created", UserID: user.ID}
		// if err := u.auditLogRepo.Create(txCtx, auditLog); err != nil {
		//     return err // Will rollback all previous operations
		// }

		// If we reach here, all operations succeeded and will be committed
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Transaction committed successfully, now send async notifications
	// These are outside the transaction because they're not critical
	// and we don't want to rollback the user creation if notification fails

	// Send PubSub message
	message := fmt.Sprintf("New user created: %s (%s)", user.Name, user.Email)
	if id, err := u.pubsubClient.Publish(ctx, u.config.PubSubTopicUserCreated, []byte(message), nil); err != nil {
		u.logger.Errorf(ctx, "Failed to publish pubsub message: %v", err)
	} else {
		u.logger.Infof(ctx, "Published message: id=%s", id)
	}

	// Send Telegram message with asynq task
	taskPayload := fmt.Sprintf("New user created: %s (%s)", user.Name, user.Email)
	task, _ := worker.NewTelegramMessageTask(ctx, u.config.TelegramChatID, taskPayload)
	if task != nil {
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
