package service

import (
	"context"
	"honeypot-day4/internal/orchestration"
)

type TelnetService struct{}

func (s *TelnetService) Name() string {
	return "telnet"
}

func (s *TelnetService) Start(ctx context.Context) error {
	return nil
}

var _ orchestration.Service = (*TelnetService)(nil)
