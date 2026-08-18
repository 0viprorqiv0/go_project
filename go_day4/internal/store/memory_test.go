package store_test

import (
	"context"
	"testing"
	"honeypot-day4/internal/model"
	"honeypot-day4/internal/store"
)

func TestMomoryEventsStoreKeepsIndependentCopies(t * testing.T){
	eventStore := store.NewMemoryEventStore()
	event := &model.Event{Service: "http", Message: "request"}

	if err := eventStore.InsertEvent(context.Background(), event); err != nil{
		t.Fatalf("InsertEvent() error = %v", err)
	}
	event.Message = "changed after insert"

	firstRead := eventStore.Events()
	firstRead[0].Message = "changed returned copy"

	secondRead := eventStore.Events()
	if got, want := secondRead[0].Message, "request"; got != want{
		t.Fatalf("stored Message = %q, want %q", got, want)
	}
}

func TestMemoryEventStoreHonorsCancelledContext(t *testing.T){
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	eventStore := store.NewMemoryEventStore()
	err := eventStore.InsertEvent(ctx, &model.Event{})
	if err !=  context.Canceled{
		t.Fatalf("InsertEvent() error = %v, want %v", err, context.Canceled)
	}
	if got := len(eventStore.Events()); got != 0{
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}