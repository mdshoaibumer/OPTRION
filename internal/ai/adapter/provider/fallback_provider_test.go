package provider

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

// mockProvider is a test helper.
type mockProvider struct {
	name string
	err  error
	data []byte
}

func (m *mockProvider) Analyze(_ context.Context, _ []byte) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func (m *mockProvider) Name() string { return m.name }

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestFallbackProvider_UsePrimary(t *testing.T) {
	primary := &mockProvider{name: "primary", data: []byte("primary-result")}
	secondary := &mockProvider{name: "secondary", data: []byte("secondary-result")}

	fp, err := NewFallbackProvider(testLogger(), primary, secondary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := fp.Analyze(context.Background(), []byte("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "primary-result" {
		t.Fatalf("expected primary-result, got %s", string(result))
	}
}

func TestFallbackProvider_FallbackOnError(t *testing.T) {
	primary := &mockProvider{name: "primary", err: fmt.Errorf("primary failed")}
	secondary := &mockProvider{name: "secondary", data: []byte("secondary-result")}

	fp, err := NewFallbackProvider(testLogger(), primary, secondary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := fp.Analyze(context.Background(), []byte("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "secondary-result" {
		t.Fatalf("expected secondary-result, got %s", string(result))
	}
}

func TestFallbackProvider_AllFail(t *testing.T) {
	p1 := &mockProvider{name: "p1", err: fmt.Errorf("p1 failed")}
	p2 := &mockProvider{name: "p2", err: fmt.Errorf("p2 failed")}

	fp, err := NewFallbackProvider(testLogger(), p1, p2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = fp.Analyze(context.Background(), []byte("test"))
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestFallbackProvider_NoProviders(t *testing.T) {
	_, err := NewFallbackProvider(testLogger())
	if err == nil {
		t.Fatal("expected error with no providers")
	}
}

func TestFallbackProvider_Name(t *testing.T) {
	p1 := &mockProvider{name: "gemini"}
	p2 := &mockProvider{name: "openai"}

	fp, _ := NewFallbackProvider(testLogger(), p1, p2)
	name := fp.Name()
	if name != "fallback(gemini+1)" {
		t.Fatalf("expected fallback(gemini+1), got %s", name)
	}
}
