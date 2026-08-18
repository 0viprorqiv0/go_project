package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type SSHService struct {
	processor *processor.Processor
}

func NewSSH(p *processor.Processor) *SSHService {
	return &SSHService{processor: p}
}

func (s *SSHService) Name() string {
	return "ssh"
}

func (s *SSHService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*SSHService)(nil)
