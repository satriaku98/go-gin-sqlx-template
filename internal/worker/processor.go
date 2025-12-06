package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gin-sqlx-template/internal/integration/telegram"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TelegramTaskHandler struct {
	logger          *zap.SugaredLogger
	telegramService *telegram.TelegramService
}

func NewTelegramTaskHandler(logger *zap.SugaredLogger, telegramService *telegram.TelegramService) *TelegramTaskHandler {
	return &TelegramTaskHandler{
		logger:          logger,
		telegramService: telegramService,
	}
}

func (h *TelegramTaskHandler) HandleTelegramMessageTask(ctx context.Context, t *asynq.Task) error {
	var p TelegramMessagePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logger.Infof("Sending telegram message to %s", p.ChatID)
	err := h.telegramService.SendMessage(ctx, p.ChatID, p.Text)
	if err != nil {
		h.logger.Errorf("Failed to send telegram message: %v", err)
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	h.logger.Infof("Telegram message sent successfully")
	return nil
}
