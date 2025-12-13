package logger

import (
	"context"

	"github.com/hibiken/asynq"
)

// AsynqLoggerAdapter adapts our Logger to implement asynq.Logger interface
type AsynqLoggerAdapter struct {
	logger *Logger
	ctx    context.Context
}

// NewAsynqLoggerAdapter creates a new adapter for asynq.Logger
func NewAsynqLoggerAdapter(logger *Logger) asynq.Logger {
	return &AsynqLoggerAdapter{
		logger: logger,
		ctx:    context.Background(),
	}
}

func (a *AsynqLoggerAdapter) Debug(args ...any) {
	a.logger.Debug(a.ctx, args...)
}

func (a *AsynqLoggerAdapter) Info(args ...any) {
	a.logger.Info(a.ctx, args...)
}

func (a *AsynqLoggerAdapter) Warn(args ...any) {
	a.logger.Warn(a.ctx, args...)
}

func (a *AsynqLoggerAdapter) Error(args ...any) {
	a.logger.Error(a.ctx, args...)
}

func (a *AsynqLoggerAdapter) Fatal(args ...any) {
	a.logger.Fatal(a.ctx, args...)
}
