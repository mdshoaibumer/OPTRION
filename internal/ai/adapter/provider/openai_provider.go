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

// OpenAIProvider implements the AIProvider interface using OpenAI's Chat Completions API.
type OpenAIProvider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// OpenAIConfig holds configuration for the OpenAI provider.
type OpenAIConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewOpenAIProvider creates a configured OpenAI AI provider.
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	return &OpenAIProvider{
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

// openAIRequest is the Chat Completions API request format.
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse is the Chat Completions API response format.
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (o *OpenAIProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("openai: API key is not configured")
	}

	prompt := fmt.Sprintf(`You are an expert SRE performing root cause analysis. Analyze the following incident context and provide a structured JSON response.

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

	reqBody := openAIRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are an expert Site Reliability Engineer. Respond only with valid JSON."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   o.maxTokens,
		Temperature: 0.2,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openai: marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openai: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("openai: reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("openai: decoding response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("openai: API error: %s (type: %s)", openAIResp.Error.Message, openAIResp.Error.Type)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response from model")
	}

	output := openAIResp.Choices[0].Message.Content
	output = extractJSON(output)

	return []byte(output), nil
}

func (o *OpenAIProvider) Name() string {
	return "OpenAI"
}
