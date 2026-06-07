package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
	"github.com/optrion/optrion/internal/shared/id"
)

// WebhookSender implements alert delivery via generic HTTP webhooks.
type WebhookSender struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewWebhookSender creates a new webhook message sender.
func NewWebhookSender(maxRetries int, retryDelay time.Duration) *WebhookSender {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if retryDelay == 0 {
		retryDelay = 2 * time.Second
	}
	return &WebhookSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// WebhookPayload is the standard payload sent to webhook endpoints.
type WebhookPayload struct {
	ID        string                 `json:"id"`
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Send delivers a message through a webhook channel with retry logic.
func (ws *WebhookSender) Send(ctx context.Context, channel *alertchannel.AlertChannel, message string) (deliveryID string, err error) {
	webhookURL := channel.Config["url"]
	if webhookURL == "" {
		return "", fmt.Errorf("webhook channel missing url configuration")
	}

	// Validate URL scheme for security
	if !isAllowedScheme(webhookURL) {
		return "", fmt.Errorf("webhook URL must use https:// scheme")
	}

	deliveryID = id.New()

	payload := WebhookPayload{
		ID:        deliveryID,
		Event:     "alert",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return deliveryID, fmt.Errorf("marshaling webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= ws.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return deliveryID, ctx.Err()
			case <-time.After(ws.retryDelay * time.Duration(attempt)):
			}
		}

		lastErr = ws.sendRequest(ctx, webhookURL, body, channel.Config["secret"])
		if lastErr == nil {
			return deliveryID, nil
		}
	}

	return deliveryID, fmt.Errorf("webhook delivery failed after %d attempts: %w", ws.maxRetries+1, lastErr)
}

// sendRequest makes a single HTTP request to the webhook URL.
func (ws *WebhookSender) sendRequest(ctx context.Context, url string, body []byte, secret string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Optrion-Webhook/1.0")

	// Sign payload with HMAC-SHA256 if secret is configured
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Optrion-Signature", "sha256="+signature)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending webhook request: connection failed")
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}

	return nil
}

// isAllowedScheme validates the webhook URL uses HTTPS.
func isAllowedScheme(rawURL string) bool {
	return len(rawURL) >= 8 && rawURL[:8] == "https://"
}
