package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEBroker_SubscribeUnsubscribe(t *testing.T) {
	broker := NewSSEBroker(nil)

	client := broker.Subscribe("tenant-1")
	if broker.ClientCount("tenant-1") != 1 {
		t.Fatalf("expected 1 client, got %d", broker.ClientCount("tenant-1"))
	}

	broker.Unsubscribe(client)
	if broker.ClientCount("tenant-1") != 0 {
		t.Fatalf("expected 0 clients after unsubscribe, got %d", broker.ClientCount("tenant-1"))
	}
}

func TestSSEBroker_BroadcastReachesClients(t *testing.T) {
	broker := NewSSEBroker(nil)

	client1 := broker.Subscribe("tenant-1")
	client2 := broker.Subscribe("tenant-1")
	defer broker.Unsubscribe(client1)
	defer broker.Unsubscribe(client2)

	event := SSEEvent{Type: "incident.opened", Payload: map[string]string{"id": "inc-1"}}
	broker.Broadcast("tenant-1", event)

	select {
	case received := <-client1.Events:
		if received.Type != "incident.opened" {
			t.Fatalf("expected event type incident.opened, got %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event on client1")
	}

	select {
	case received := <-client2.Events:
		if received.Type != "incident.opened" {
			t.Fatalf("expected event type incident.opened, got %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event on client2")
	}
}

func TestSSEBroker_TenantIsolation(t *testing.T) {
	broker := NewSSEBroker(nil)

	client1 := broker.Subscribe("tenant-1")
	client2 := broker.Subscribe("tenant-2")
	defer broker.Unsubscribe(client1)
	defer broker.Unsubscribe(client2)

	broker.Broadcast("tenant-1", SSEEvent{Type: "test", Payload: nil})

	select {
	case <-client1.Events:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("tenant-1 client should have received event")
	}

	select {
	case <-client2.Events:
		t.Fatal("tenant-2 client should NOT have received tenant-1 event")
	case <-time.After(100 * time.Millisecond):
		// Expected: no event for tenant-2
	}
}

func TestSSEBroker_BufferFullDropsEvent(t *testing.T) {
	broker := NewSSEBroker(nil)

	client := broker.Subscribe("tenant-1")
	defer broker.Unsubscribe(client)

	// Fill the buffer (capacity is 64)
	for i := 0; i < 65; i++ {
		broker.Broadcast("tenant-1", SSEEvent{Type: "flood", Payload: i})
	}

	// Should have 64 events in buffer, 65th dropped
	count := 0
	for {
		select {
		case <-client.Events:
			count++
		default:
			goto done
		}
	}
done:
	if count != 64 {
		t.Fatalf("expected 64 buffered events, got %d", count)
	}
}

func TestSSEHandler_RequiresAuth(t *testing.T) {
	broker := NewSSEBroker(nil)
	handler := SSEHandler(broker, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/stream", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSSEHandler_StreamsEvents(t *testing.T) {
	broker := NewSSEBroker(nil)
	handler := SSEHandler(broker, nil)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, ContextKeyTenantID, "tenant-test")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/stream", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	// Give handler time to subscribe
	time.Sleep(50 * time.Millisecond)

	broker.Broadcast("tenant-test", SSEEvent{Type: "test.event", Payload: "hello"})

	// Give time for event delivery
	time.Sleep(50 * time.Millisecond)

	cancel()
	<-done

	body := rec.Body.String()
	if len(body) == 0 {
		t.Fatal("expected SSE output, got empty body")
	}
}

func TestSSEEventBusAdapter_ForwardsEvents(t *testing.T) {
	broker := NewSSEBroker(nil)
	adapter := NewSSEEventBusAdapter(broker)

	client := broker.Subscribe("tenant-1")
	defer broker.Unsubscribe(client)

	adapter.ForwardEvent(context.Background(), "tenant-1", "incident.opened", map[string]string{"id": "inc-1"})

	select {
	case event := <-client.Events:
		if event.Type != "incident.opened" {
			t.Fatalf("expected incident.opened, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for forwarded event")
	}
}
