package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements the AIProvider interface using a local Ollama instance.
type OllamaProvider struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

// OllamaConfig holds configuration for the Ollama provider.
type OllamaConfig struct {
	Endpoint string
	Model    string
}

// NewOllamaProvider creates a configured Ollama AI provider.
func NewOllamaProvider(cfg OllamaConfig) *OllamaProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "llama3"
	}
	return &OllamaProvider{
		endpoint: cfg.Endpoint,
		model:    cfg.Model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Ollama can be slow on first inference
		},
	}
}

// ollamaRequest is the Ollama generate API request format.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

// ollamaResponse is the Ollama generate API response format.
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func (o *OllamaProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
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

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", o.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ollama: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed (is Ollama running at %s?): %w", o.endpoint, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("ollama: reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("ollama: decoding response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama: API error: %s", ollamaResp.Error)
	}

	if ollamaResp.Response == "" {
		return nil, fmt.Errorf("ollama: empty response from model")
	}

	output := extractJSON(ollamaResp.Response)

	return []byte(output), nil
}

func (o *OllamaProvider) Name() string {
	return "Ollama"
}
