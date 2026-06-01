package eventbus

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
)

type testEvent struct {
	eventType string
	tenantID  string
}

func (e testEvent) EventType() string { return e.eventType }
func (e testEvent) TenantID() string  { return e.tenantID }

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestBus_PublishToSubscriber(t *testing.T) {
	bus := New(testLogger())

	var called int32
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	bus.Publish(context.Background(), testEvent{eventType: "test.event", tenantID: "t1"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected handler to be called once, got %d", atomic.LoadInt32(&called))
	}
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := New(testLogger())

	var count int32
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	bus.Publish(context.Background(), testEvent{eventType: "test.event", tenantID: "t1"})

	if atomic.LoadInt32(&count) != 2 {
		t.Fatalf("expected 2 handlers called, got %d", atomic.LoadInt32(&count))
	}
}

func TestBus_NoSubscribers(t *testing.T) {
	bus := New(testLogger())

	// Should not panic
	bus.Publish(context.Background(), testEvent{eventType: "unsubscribed", tenantID: "t1"})
}

func TestBus_DifferentEventTypes(t *testing.T) {
	bus := New(testLogger())

	var eventACalled, eventBCalled int32
	bus.Subscribe("event.a", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&eventACalled, 1)
		return nil
	})
	bus.Subscribe("event.b", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&eventBCalled, 1)
		return nil
	})

	bus.Publish(context.Background(), testEvent{eventType: "event.a", tenantID: "t1"})

	if atomic.LoadInt32(&eventACalled) != 1 {
		t.Fatal("expected event.a handler to be called")
	}
	if atomic.LoadInt32(&eventBCalled) != 0 {
		t.Fatal("expected event.b handler NOT to be called")
	}
}

func TestBus_HandlerErrorDoesNotStopOthers(t *testing.T) {
	bus := New(testLogger())

	var secondCalled int32
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		return errors.New("handler error")
	})
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&secondCalled, 1)
		return nil
	})

	bus.Publish(context.Background(), testEvent{eventType: "test.event", tenantID: "t1"})

	if atomic.LoadInt32(&secondCalled) != 1 {
		t.Fatal("second handler should still be called despite first handler error")
	}
}

func TestBus_PublishAll(t *testing.T) {
	bus := New(testLogger())

	var count int32
	bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	events := []Event{
		testEvent{eventType: "test.event", tenantID: "t1"},
		testEvent{eventType: "test.event", tenantID: "t2"},
		testEvent{eventType: "test.event", tenantID: "t3"},
	}
	bus.PublishAll(context.Background(), events)

	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 calls, got %d", atomic.LoadInt32(&count))
	}
}

func TestBus_EventMetadata(t *testing.T) {
	bus := New(testLogger())

	var receivedTenant string
	var receivedType string
	bus.Subscribe("incident.opened", func(ctx context.Context, event Event) error {
		receivedTenant = event.TenantID()
		receivedType = event.EventType()
		return nil
	})

	bus.Publish(context.Background(), testEvent{eventType: "incident.opened", tenantID: "tenant-123"})

	if receivedTenant != "tenant-123" {
		t.Fatalf("expected tenant 'tenant-123', got %q", receivedTenant)
	}
	if receivedType != "incident.opened" {
		t.Fatalf("expected event type 'incident.opened', got %q", receivedType)
	}
}
