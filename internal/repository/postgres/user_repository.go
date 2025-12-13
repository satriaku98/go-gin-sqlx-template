package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"go-gin-sqlx-template/internal/model"
	"go-gin-sqlx-template/internal/repository"
	"go-gin-sqlx-template/pkg/database"
	"go-gin-sqlx-template/pkg/utils"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, name, password, created_at, updated_at)
		VALUES (:email, :name, :password, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	args := map[string]any{
		"email":    user.Email,
		"name":     user.Name,
		"password": user.Password,
	}

	row, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer row.Close()

	if row.Next() {
		err = row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan created user: %w", err)
		}
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, name, password, created_at, updated_at FROM users WHERE id = :id`

	args := map[string]any{
		"id": id,
	}

	row, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer row.Close()

	if row.Next() {
		err = row.StructScan(&user)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, name, password, created_at, updated_at FROM users WHERE email = :email`

	args := map[string]any{
		"email": email,
	}

	row, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer row.Close()

	if !row.Next() {
		return nil, fmt.Errorf("user not found")
	}

	err = row.StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int, filters utils.FilterParams) ([]model.User, error) {
	var users []model.User

	// Build WHERE clause based on filters
	whereClauses := []string{}
	args := map[string]any{
		"limit":  limit,
		"offset": offset,
	}

	if name, ok := filters.Get("name"); ok {
		whereClauses = append(whereClauses, "name ILIKE :name")
		args["name"] = "%" + name + "%"
	}

	if email, ok := filters.Get("email"); ok {
		whereClauses = append(whereClauses, "email ILIKE :email")
		args["email"] = "%" + email + "%"
	}

	// Build query
	query := `SELECT id, email, name, password, created_at, updated_at FROM users`
	query += utils.BuildWhereClause(whereClauses)
	query += ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`

	rows, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	err = sqlx.StructScan(rows, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", err)
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users 
		SET email = :email, name = :name, updated_at = NOW()
		WHERE id = :id
		RETURNING updated_at
	`
	args := map[string]any{
		"email": user.Email,
		"name":  user.Name,
		"id":    user.ID,
	}

	row, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	if !row.Next() {
		return fmt.Errorf("user not found")
	}

	err = row.StructScan(&user)
	if err != nil {
		return fmt.Errorf("failed to scan updated user: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = :id`

	args := map[string]any{
		"id": id,
	}

	result, err := r.db.NamedExecContext(ctx, query, database.SetMapSqlNamed(args))
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Count(ctx context.Context, filters utils.FilterParams) (int64, error) {
	var count int64

	// Build WHERE clause based on filters
	whereClauses := []string{}
	args := map[string]any{}

	if name, ok := filters.Get("name"); ok {
		whereClauses = append(whereClauses, "name ILIKE :name")
		args["name"] = "%" + name + "%"
	}

	if email, ok := filters.Get("email"); ok {
		whereClauses = append(whereClauses, "email ILIKE :email")
		args["email"] = "%" + email + "%"
	}

	// Build query
	query := `SELECT COUNT(*) FROM users`
	query += utils.BuildWhereClause(whereClauses)

	if len(args) > 0 {
		// Use NamedQuery for parameterized query
		rows, err := r.db.NamedQueryContext(ctx, query, database.SetMapSqlNamed(args))
		if err != nil {
			return 0, fmt.Errorf("failed to count users: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return 0, fmt.Errorf("failed to scan count: %w", err)
			}
		}
	} else {
		// No filters, use simple query
		err := r.db.GetContext(ctx, &count, query)
		if err != nil {
			return 0, fmt.Errorf("failed to count users: %w", err)
		}
	}

	return count, nil
}
