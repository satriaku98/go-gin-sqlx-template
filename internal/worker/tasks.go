package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task Types
const (
	TypeTelegramMessage = "telegram:send_message"
)

// Payload
type TelegramMessagePayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// NewTelegramMessageTask creates a new task for sending telegram messages
func NewTelegramMessageTask(chatID, text string) (*asynq.Task, error) {
	payload := TelegramMessagePayload{
		ChatID: chatID,
		Text:   text,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTelegramMessage, payloadBytes), nil
}
