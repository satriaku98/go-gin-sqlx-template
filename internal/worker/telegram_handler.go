package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/pkg/logger"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// TelegramTaskHandler handles telegram-related tasks
type TelegramTaskHandler struct {
	logger          *logger.Logger
	telegramService *telegram.TelegramService
}

// NewTelegramTaskHandler creates a new TelegramTaskHandler
func NewTelegramTaskHandler(logger *logger.Logger, telegramService *telegram.TelegramService) *TelegramTaskHandler {
	return &TelegramTaskHandler{
		logger:          logger,
		telegramService: telegramService,
	}
}

// HandleTelegramMessageTask processes telegram message sending tasks
func (h *TelegramTaskHandler) HandleTelegramMessageTask(ctx context.Context, t *asynq.Task) error {
	var p TelegramMessagePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Extract trace context and start span
	if p.TraceContext != nil {
		carrier := propagation.MapCarrier(p.TraceContext)
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	tracer := otel.Tracer(t.ResultWriter().TaskID())
	ctx, span := tracer.Start(ctx, "HandleTelegramMessageTask")
	defer span.End()

	h.logger.Info("Sending telegram message to %s", p.ChatID)
	err := h.telegramService.SendMessage(ctx, p.ChatID, p.Text)
	if err != nil {
		h.logger.Error("Failed to send telegram message: %v", err)
		span.RecordError(err)
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	h.logger.Info("Telegram message sent successfully")
	return nil
}
