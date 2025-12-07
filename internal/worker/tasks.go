package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Task Types
const (
	TypeTelegramMessage = "telegram:send_message"
)

// Payload
type TelegramMessagePayload struct {
	ChatID       string            `json:"chat_id"`
	Text         string            `json:"text"`
	TraceContext map[string]string `json:"trace_context"`
}

// NewTelegramMessageTask creates a new task for sending telegram messages
func NewTelegramMessageTask(ctx context.Context, chatID, text string) (*asynq.Task, error) {
	// Inject trace context
	traceContext := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(traceContext))

	payload := TelegramMessagePayload{
		ChatID:       chatID,
		Text:         text,
		TraceContext: traceContext,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTelegramMessage, payloadBytes), nil
}
