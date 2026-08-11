package model

import "time"

type Event struct {
	ID        string
	SourceIP  string
	Service   string
	Type      string
	Timestamp time.Time
}
