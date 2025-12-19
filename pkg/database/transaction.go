package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// txKey is the context key for storing transaction
	txKey contextKey = "tx"
)

// Transactor defines the interface for transaction management
type Transactor interface {
	// WithTransaction executes the given function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// If the function completes successfully, the transaction is committed.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// GetExecutor returns the appropriate executor (DB or TX) from context.
	// If a transaction exists in context, it returns the transaction.
	// Otherwise, it returns the database connection.
	GetExecutor(ctx context.Context) sqlx.ExtContext
}

// TransactionManager implements the Transactor interface
type TransactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sqlx.DB) Transactor {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if we're already in a transaction
	if tx := tm.getTxFromContext(ctx); tx != nil {
		// Already in a transaction, just execute the function
		// This allows nested WithTransaction calls to reuse the same transaction
		return fn(ctx)
	}

	// Start a new transaction
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store transaction in context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Defer rollback in case of panic
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	// Execute the function
	err = fn(txCtx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetExecutor returns the appropriate executor from context
func (tm *TransactionManager) GetExecutor(ctx context.Context) sqlx.ExtContext {
	if tx := tm.getTxFromContext(ctx); tx != nil {
		return tx
	}
	return tm.db
}

// getTxFromContext retrieves transaction from context if it exists
func (tm *TransactionManager) getTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}
