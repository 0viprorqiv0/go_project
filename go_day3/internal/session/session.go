package session

import (
	"errors"
	"time"
)

type Session struct {
	id           string
	sourceIP     string
	startedAt    time.Time
	commandCount int
	closed       bool
}

func NewSession(id, sourceIP string, now time.Time) (*Session, error) {
	if id == "" || sourceIP == "" {
		return nil, errors.New("id and source IP are required")
	}
	return &Session{
		id: id,
		sourceIP: sourceIP,
		startedAt: now,
	}, nil
}
