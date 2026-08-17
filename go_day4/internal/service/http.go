package service

import (
	"context"
	"honeypot-day4/internal/orchestration"
)

type HTTPService struct{}

func (s *HTTPService) Name() string {
	return "http"
}

func (s *HTTPService) Start(ctx context.Context) error {
	return nil
}

var _ orchestration.Service = (*HTTPService)(nil)
