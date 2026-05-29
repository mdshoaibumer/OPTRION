package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/optrion/optrion/internal/platform/config"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	log.Info("test message", "key", "value")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", entry["key"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "text"}
	log := NewWithWriter(cfg, &buf)

	log.Info("text test")

	if buf.Len() == 0 {
		t.Fatal("expected non-empty log output")
	}
}

func TestContextHandler_CorrelationID(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	ctx := WithCorrelationID(context.Background(), "corr-123")
	log.InfoContext(ctx, "with correlation")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry["correlation_id"] != "corr-123" {
		t.Errorf("expected correlation_id 'corr-123', got %v", entry["correlation_id"])
	}
}

func TestContextHandler_RequestID(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	ctx := WithRequestID(context.Background(), "req-456")
	log.InfoContext(ctx, "with request id")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry["request_id"] != "req-456" {
		t.Errorf("expected request_id 'req-456', got %v", entry["request_id"])
	}
}

func TestContextHandler_BothIDs(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	ctx := WithCorrelationID(context.Background(), "corr-abc")
	ctx = WithRequestID(ctx, "req-def")
	log.InfoContext(ctx, "both ids")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry["correlation_id"] != "corr-abc" {
		t.Errorf("expected correlation_id 'corr-abc', got %v", entry["correlation_id"])
	}
	if entry["request_id"] != "req-def" {
		t.Errorf("expected request_id 'req-def', got %v", entry["request_id"])
	}
}

func TestLogLevel_Filtering(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "warn", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	log.Info("this should not appear")
	if buf.Len() != 0 {
		t.Error("expected info message to be filtered at warn level")
	}

	log.Warn("this should appear")
	if buf.Len() == 0 {
		t.Error("expected warn message to pass at warn level")
	}
}

func TestWithAttrs_PreservesContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := NewWithWriter(cfg, &buf)

	child := log.With("service", "optrion")
	ctx := WithCorrelationID(context.Background(), "corr-child")
	child.InfoContext(ctx, "child logger")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry["service"] != "optrion" {
		t.Errorf("expected service 'optrion', got %v", entry["service"])
	}
	if entry["correlation_id"] != "corr-child" {
		t.Errorf("expected correlation_id 'corr-child', got %v", entry["correlation_id"])
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"DEBUG", slog.LevelDebug},
		{"unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseLevel(tt.input); got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
