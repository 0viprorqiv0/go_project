package processor_test

import (
	"context"
	"honeypot-day4/internal/model"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/store"
	"testing"
)

func TestProcessStoresEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	event := &model.Event{Service: "telnet", Message: "login attempt"}

	if err := eventProcessor.Process(context.Background(), event); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 1 {
		t.Fatalf("len(Events()) = %d, want 1", len(events))
	}
	if events[0] != *event {
		t.Fatalf("stored event = %#v, want %#v", events[0], *event)
	}
}

func TestProcessRejectsNilEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)

	err := eventProcessor.Process(context.Background(), nil)
	if err == nil {
		t.Fatalf("Process() error = %q, want %q", err, "event is requied")
	}
	if got := len(eventStore.Events()); got != 0 {
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}
