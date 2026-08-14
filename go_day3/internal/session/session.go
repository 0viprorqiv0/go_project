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

type Identity struct {
	ID string
	SourceIP string
}

type Snapshot struct {
	Identity Identity
	StartedAt time.Time
	CommandCount int
	Closed bool
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

func (s Session) Snapshot() Snapshot{
	return Snapshot{
		Identity: Identity{
			ID: s.id,
			SourceIP: s.sourceIP,
		},
		StartedAt: s.startedAt,
		CommandCount: s.commandCount,
		Closed: s.closed,
	}
}

func (s *Session) AddCommand() error {
	if s.closed{
		return errors.New("session is closed")
	}
	s.commandCount++
	return nil
}

func (s *Session) Close() {
	s.closed = true
}
