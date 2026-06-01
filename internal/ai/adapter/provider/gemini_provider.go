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

// GeminiProvider implements the AIProvider interface using Google's Gemini API.
type GeminiProvider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// GeminiConfig holds configuration for the Gemini provider.
type GeminiConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewGeminiProvider creates a configured Gemini AI provider.
func NewGeminiProvider(cfg GeminiConfig) *GeminiProvider {
	if cfg.Model == "" {
		cfg.Model = "gemini-2.0-flash"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	return &GeminiProvider{
		apiKey:    cfg.apiKey(),
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

func (c GeminiConfig) apiKey() string {
	return c.APIKey
}

// geminiRequest is the Gemini API request format.
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
}

// geminiResponse is the Gemini API response format.
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (g *GeminiProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("gemini: API key is not configured")
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

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: g.maxTokens,
			Temperature:     0.2,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshaling request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", g.model)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
	if err != nil {
		return nil, fmt.Errorf("gemini: reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: decoding response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini: API error: %s (code %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty response from model")
	}

	output := geminiResp.Candidates[0].Content.Parts[0].Text

	// Extract JSON from response (model may wrap in markdown code block)
	output = extractJSON(output)

	return []byte(output), nil
}

func (g *GeminiProvider) Name() string {
	return "Gemini"
}

// extractJSON strips markdown code fences if present.
func extractJSON(s string) string {
	// Check for ```json ... ``` pattern
	if len(s) > 7 && s[:7] == "```json" {
		end := len(s) - 1
		for end > 7 && s[end] != '`' {
			end--
		}
		if end > 10 && s[end-2:end+1] == "```" {
			s = s[7 : end-2]
		}
	} else if len(s) > 3 && s[:3] == "```" {
		// Plain ``` ... ```
		start := 3
		for start < len(s) && s[start] != '\n' {
			start++
		}
		end := len(s) - 1
		for end > start && s[end] != '`' {
			end--
		}
		if end > start+3 && s[end-2:end+1] == "```" {
			s = s[start : end-2]
		}
	}
	return s
}
