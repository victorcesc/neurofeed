package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

const telegramAPIMessageLimit = 4096

// TelegramNotifier sends plain text via the Telegram Bot API sendMessage method.
type TelegramNotifier struct {
	// APIBaseURL is optional. If empty, https://api.telegram.org is used. Set in tests to redirect HTTP.
	APIBaseURL string
	Token      string
	ChatID     string
	Client     *http.Client
}

type telegramSendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramAPIResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// Notify implements Notifier: validates token/chat/client, truncates to Telegram limits, POSTs JSON to sendMessage.
func (notifier *TelegramNotifier) Notify(ctx context.Context, message string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	token := strings.TrimSpace(notifier.Token)
	chatID := strings.TrimSpace(notifier.ChatID)
	if token == "" {
		return fmt.Errorf("telegram: empty bot token")
	}
	if chatID == "" {
		return fmt.Errorf("telegram: empty chat id")
	}
	if notifier.Client == nil {
		return fmt.Errorf("telegram: nil HTTP client")
	}

	text := strings.TrimSpace(message)
	if text == "" {
		return fmt.Errorf("telegram: empty message")
	}
	if utf8.RuneCountInString(text) > telegramAPIMessageLimit {
		runes := []rune(text)
		text = string(runes[:telegramAPIMessageLimit-1]) + "…"
	}

	body, err := json.Marshal(telegramSendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		DisableWebPagePreview: true,
	})
	if err != nil {
		return fmt.Errorf("telegram: marshal request: %w", err)
	}

	baseURL := strings.TrimSpace(notifier.APIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", baseURL, token)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := notifier.Client.Do(request)
	if err != nil {
		return fmt.Errorf("telegram: http: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("telegram: read body: %w", err)
	}

	var parsed telegramAPIResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return fmt.Errorf("telegram: decode response (HTTP %d): %s", response.StatusCode, string(responseBody))
	}
	if !parsed.OK {
		return fmt.Errorf("telegram: API error: %s", parsed.Description)
	}
	return nil
}
