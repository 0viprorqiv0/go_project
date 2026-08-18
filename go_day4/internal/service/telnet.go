package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type TelnetService struct {
	processor *processor.Processor
}

func NewTelnet(p *processor.Processor) *TelnetService {
	return &TelnetService{processor: p}
}

func (s *TelnetService) Name() string {
	return "telnet"
}

func (s *TelnetService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*TelnetService)(nil)
