package session_test

import (
	"testing"
	"time"
	"honeypot-day3/internal/session"
)

func TestNewSessionRejectsMissingRequiredFields(t *testing.T){
	tests := []struct {
		name string
		id string
		sourceIP string
	}{
		{name: "missing ID", sourceIP: "2-3.0.113.10"},
		{name: "missing source IP", id: "sess-001"},
		{name: "missing both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			got, err := session.NewSession(tt.id, tt.sourceIP, time.Now())
			if err == nil{
				t.Fatal("NewSession() error = nil, want an error")
			}
			if err.Error() != "id and source IP are required" {
				t.Fatalf("NewSession() error = %q, want %q", err, "id and source IP are required")
			}
			if got != nil {
				t.Fatalf("NewSession() session = %#v, want nil", got)
			}
		})
	}
}

func TestNewSessionInitializesValidState(t *testing.T){
	now := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)

	sess, err := session.NewSession("sess-001", "203.0.113.10", now)
	if err != nil {
		t.Fatalf("NewSession)_ error = %v", err)
	}
	got := sess.Snapshot()
	if got.Identity.ID != "sess-001"{
		t.Fatalf("ID = %q, want %q", got.Identity.ID, "sess-001")
	}
	if got.Identity.SourceIP != "203.0.113.10"{
		t.Fatalf("StartedAt = %v, want %v", got.StartedAt, now)
	}
	if !got.StartedAt.Equal(now){
		t.Fatalf("StartedAt = %v, want %v", got.StartedAt, now)
	}
	if got.CommandCount != 0{
		t.Fatalf("CommandCount = %d, want 0", got.CommandCount)
	}
	if got.Closed{
		t.Fatal("new session must be open")
	}
}

func TestAddCommandIncrementsCount(t *testing.T){
	sess := newTestSession(t)

	for want := 1; want <= 2; want++{
		if err := sess.AddCommand(); err != nil{
			t.Fatalf("AddCommand() error = %v", err)
		}
		if got := sess.Snapshot().CommandCount; got != want{
			t.Fatalf("CommandCount = %d, want %d", got, want)
		}
	}
}

func newTestSession(t *testing.T) *session.Session{
	t.Helper()
	
	sess, err := session.NewSession(
		"sess-test",
		"192.0.2.10",
		time.Date(2006, time.August, 13, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	return sess
}