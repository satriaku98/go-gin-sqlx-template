package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gin-sqlx-template/internal/integration/telegram"
	"go-gin-sqlx-template/pkg/logger"

	"github.com/hibiken/asynq"
)

type TelegramTaskHandler struct {
	logger          *logger.Logger
	telegramService *telegram.TelegramService
}

func NewTelegramTaskHandler(logger *logger.Logger, telegramService *telegram.TelegramService) *TelegramTaskHandler {
	return &TelegramTaskHandler{
		logger:          logger,
		telegramService: telegramService,
	}
}

func (h *TelegramTaskHandler) HandleTelegramMessageTask(ctx context.Context, t *asynq.Task) error {
	var p TelegramMessagePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logger.Info("Sending telegram message to %s", p.ChatID)
	err := h.telegramService.SendMessage(ctx, p.ChatID, p.Text)
	if err != nil {
		h.logger.Error("Failed to send telegram message: %v", err)
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	h.logger.Info("Telegram message sent successfully")
	return nil
}
