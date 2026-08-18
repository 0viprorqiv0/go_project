package service_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/service"
	"honeypot-day4/internal/store"
)

func TestServicesStartThroughInterface(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	services := []orchestration.Service{
		service.NewTelnet(eventProcessor),
		service.NewHTTP(eventProcessor),
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 2 {
		t.Fatalf("len(Events()) = %d, want 2", len(events))
	}
	if events[0].Service != "telnet" || events[1].Service != "http" {
		t.Fatalf(
			"event services = [%q %q], want [telnet http]",
			events[0].Service,
			events[1].Service,
		)
	}
}
