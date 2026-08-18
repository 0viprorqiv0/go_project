package store

import(
	"context"
	"sync"
	"honeypot-day4/internal/model"
)

type MemoryEventStore struct{
	mu sync.RWMutex
	events []model.Event
}

func NewMemoryEventStore() *MemoryEventStore{
	return &MemoryEventStore{}
}

func (s *MemoryEventStore) InsertEvent(ctx context.Context, event *model.Event) error{
	if err := ctx.Err(); err != nil{
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, *event)
	return nil
}

func(s *MemoryEventStore) Events() []model.Event{
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]model.Event(nil), s.events...)
}
