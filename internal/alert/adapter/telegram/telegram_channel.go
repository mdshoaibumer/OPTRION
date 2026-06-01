package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
	"github.com/optrion/optrion/internal/shared/id"
)

// TelegramSender implements alert delivery via Telegram Bot API.
type TelegramSender struct {
	client     *http.Client
	rateLimit  int           // messages per minute
	retryDelay time.Duration // delay between retries
	maxRetries int
}

// NewTelegramSender creates a new Telegram message sender.
func NewTelegramSender(rateLimit int, retryDelay time.Duration) *TelegramSender {
	if rateLimit <= 0 {
		rateLimit = 30
	}
	if retryDelay == 0 {
		retryDelay = 2 * time.Second
	}
	return &TelegramSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimit:  rateLimit,
		retryDelay: retryDelay,
		maxRetries: 3,
	}
}

// Send delivers a message through a Telegram channel with retry logic.
func (ts *TelegramSender) Send(ctx context.Context, channel *alertchannel.AlertChannel, message string) (deliveryID string, err error) {
	botToken := channel.Config["bot_token"]
	chatID := channel.Config["chat_id"]

	if botToken == "" || chatID == "" {
		return "", fmt.Errorf("telegram channel missing bot_token or chat_id configuration")
	}

	deliveryID = id.New()

	var lastErr error
	for attempt := 0; attempt <= ts.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return deliveryID, ctx.Err()
			case <-time.After(ts.retryDelay * time.Duration(attempt)):
			}
		}

		lastErr = ts.sendMessage(ctx, botToken, chatID, message)
		if lastErr == nil {
			return deliveryID, nil
		}
	}

	return deliveryID, fmt.Errorf("telegram delivery failed after %d attempts: %w", ts.maxRetries+1, lastErr)
}

// sendMessage makes a single API call to Telegram.
func (ts *TelegramSender) sendMessage(ctx context.Context, botToken, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("telegram rate limited (429)")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	return nil
}
