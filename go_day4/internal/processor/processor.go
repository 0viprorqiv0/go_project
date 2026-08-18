package processor

import (
	"context"
	"errors"
	"honeypot-day4/internal/model"
)

type EventStore interface{
	InsertEvent(ctx context.Context, event *model.Event) error
}

type Processor struct{
	store EventStore
}

func New(store EventStore) *Processor{
	return &Processor{store: store}
}

func (p *Processor) Process(ctx context.Context, event *model.Event) error{
	if event == nil{
		return errors.New("event is required")
	}
	return p.store.InsertEvent(ctx, event)
}