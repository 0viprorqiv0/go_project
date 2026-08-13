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