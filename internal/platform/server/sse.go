package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SSEClient represents a connected SSE client.
type SSEClient struct {
	TenantID string
	Events   chan SSEEvent
	Done     chan struct{}
}

// SSEBroker manages Server-Sent Event subscriptions and broadcasting.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[string]map[*SSEClient]struct{} // tenantID -> clients
	logger  *slog.Logger
}

// NewSSEBroker creates a new SSE broker.
func NewSSEBroker(logger *slog.Logger) *SSEBroker {
	if logger == nil {
		logger = slog.Default()
	}
	return &SSEBroker{
		clients: make(map[string]map[*SSEClient]struct{}),
		logger:  logger,
	}
}

// Subscribe adds a new client to the broker.
func (b *SSEBroker) Subscribe(tenantID string) *SSEClient {
	client := &SSEClient{
		TenantID: tenantID,
		Events:   make(chan SSEEvent, 64),
		Done:     make(chan struct{}),
	}

	b.mu.Lock()
	if b.clients[tenantID] == nil {
		b.clients[tenantID] = make(map[*SSEClient]struct{})
	}
	b.clients[tenantID][client] = struct{}{}
	b.mu.Unlock()

	b.logger.Debug("SSE client subscribed", "tenant_id", tenantID)
	return client
}

// Unsubscribe removes a client from the broker.
func (b *SSEBroker) Unsubscribe(client *SSEClient) {
	b.mu.Lock()
	if tenantClients, ok := b.clients[client.TenantID]; ok {
		delete(tenantClients, client)
		if len(tenantClients) == 0 {
			delete(b.clients, client.TenantID)
		}
	}
	b.mu.Unlock()

	close(client.Done)
	b.logger.Debug("SSE client unsubscribed", "tenant_id", client.TenantID)
}

// Broadcast sends an event to all clients subscribed to a tenant.
func (b *SSEBroker) Broadcast(tenantID string, event SSEEvent) {
	b.mu.RLock()
	tenantClients := b.clients[tenantID]
	b.mu.RUnlock()

	for client := range tenantClients {
		select {
		case client.Events <- event:
		default:
			// Client buffer full, skip to avoid blocking
			b.logger.Warn("SSE client buffer full, dropping event",
				"tenant_id", tenantID,
				"event_type", event.Type,
			)
		}
	}
}

// ClientCount returns the number of connected clients for a tenant.
func (b *SSEBroker) ClientCount(tenantID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients[tenantID])
}

// SSEHandler returns an HTTP handler for the SSE event stream endpoint.
func SSEHandler(broker *SSEBroker, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := TenantIDFromContext(r.Context())
		if tenantID == "" {
			WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			WriteError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		client := broker.Subscribe(tenantID)
		defer broker.Unsubscribe(client)

		// Send initial connection event
		writeSSEEvent(w, flusher, SSEEvent{Type: "connected", Payload: map[string]string{
			"message": "SSE stream established",
		}})

		// Heartbeat ticker to keep connection alive
		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-client.Events:
				writeSSEEvent(w, flusher, event)
			case <-heartbeat.C:
				writeSSEEvent(w, flusher, SSEEvent{Type: "heartbeat", Payload: map[string]string{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}})
			}
		}
	}
}

// writeSSEEvent writes a single SSE event to the response writer.
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
	flusher.Flush()
}

// SSEEventBusAdapter bridges the internal event bus to SSE clients.
type SSEEventBusAdapter struct {
	broker *SSEBroker
}

// NewSSEEventBusAdapter creates an adapter that forwards event bus events to SSE.
func NewSSEEventBusAdapter(broker *SSEBroker) *SSEEventBusAdapter {
	return &SSEEventBusAdapter{broker: broker}
}

// ForwardEvent converts an event bus event to an SSE event and broadcasts it.
func (a *SSEEventBusAdapter) ForwardEvent(_ context.Context, tenantID, eventType string, payload interface{}) {
	a.broker.Broadcast(tenantID, SSEEvent{
		Type:    eventType,
		Payload: payload,
	})
}
