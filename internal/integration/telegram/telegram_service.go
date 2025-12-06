package telegram

import (
	"context"
	"fmt"
	"go-gin-sqlx-template/pkg/utils"
	"net/url"
)

type TelegramService struct {
	BaseURL string
	Token   string
}

func NewTelegramService(token, baseURL string) *TelegramService {
	return &TelegramService{
		BaseURL: baseURL,
		Token:   token,
	}
}

func (s *TelegramService) SendMessage(ctx context.Context, chatID string, text string) error {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", s.BaseURL, s.Token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	formData.Set("text", text)
	encodedBody := formData.Encode()

	// Configure request
	config := utils.HttpRequestConfig{
		Method: utils.MethodPost,
		URL:    endpoint,
		Headers: map[string]string{
			utils.HeaderContentType: utils.ContentTypeForm,
		},
		Body: encodedBody,
	}

	// Send request
	resp, err := utils.SendRequest(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to send telegram request: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram api returned status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}
