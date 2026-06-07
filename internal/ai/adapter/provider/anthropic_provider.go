package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// AnthropicProvider implements the AIProvider interface using Anthropic's Messages API.
type AnthropicProvider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// AnthropicConfig holds configuration for the Anthropic provider.
type AnthropicConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewAnthropicProvider creates a configured Anthropic AI provider.
func NewAnthropicProvider(cfg AnthropicConfig) *AnthropicProvider {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	return &AnthropicProvider{
		apiKey:    cfg.APIKey,
		model:     cfg.Model,
		maxTokens: cfg.MaxTokens,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

// anthropicRequest is the Messages API request format.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the Messages API response format.
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *AnthropicProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("anthropic: API key is not configured")
	}

	prompt := fmt.Sprintf(`Analyze the following incident context and provide a structured JSON response.

INCIDENT CONTEXT:
%s

Respond ONLY with valid JSON in this exact format:
{
  "root_cause": "A clear, concise description of the root cause",
  "affected_components": ["component1", "component2"],
  "confidence": 0.85,
  "investigation_hints": ["Check X", "Verify Y", "Monitor Z"]
}

Rules:
- confidence must be between 0.0 and 1.0
- Be specific and actionable in investigation_hints
- Only include components that are actually affected
- Do NOT include speculative causes without evidence`, string(analysisContext))

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: a.maxTokens,
		System:    "You are an expert Site Reliability Engineer performing root cause analysis. Respond only with valid JSON.",
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("anthropic: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("anthropic: reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("anthropic: decoding response: %w", err)
	}

	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("anthropic: API error: %s (type: %s)", anthropicResp.Error.Message, anthropicResp.Error.Type)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("anthropic: empty response from model")
	}

	output := anthropicResp.Content[0].Text
	output = extractJSON(output)

	return []byte(output), nil
}

func (a *AnthropicProvider) Name() string {
	return "Anthropic"
}
