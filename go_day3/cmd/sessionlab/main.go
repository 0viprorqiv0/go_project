package main

import (
	"fmt"
	"time"
	"honeypot-day3/internal/session"
)

func main() {
	startedAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	sess, err := session.NewSession("sess-001", "203.0.113.10", startedAt)
	if err != nil {
		fmt.Println("create session:", err)
		return
	}
	fmt.Printf("session created: %T\n", sess)
}

