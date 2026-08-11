package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"honeypot-lab/internal/model"
)

const eventTypeConnectionAttempt = "connection_attempt"

type config struct {
	sourceIP string
	port     int
	service  string
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		flag.Usage()
		os.Exit(1)
	}

	event := newFakeEvent(cfg)
	printEvent(event, cfg.port)
}

func parseConfig() (config, error) {
	var cfg config

	flag.StringVar(&cfg.sourceIP, "source-ip", "", "Source IP observed by the honeypot")
	flag.IntVar(&cfg.port, "port", 0, "destination port from 1 to 65535")
	flag.StringVar(&cfg.service, "service", "", "Service name (e.g., ssh, http)")
	flag.Parse()

	cfg.sourceIP = strings.TrimSpace(cfg.sourceIP)
	cfg.service = strings.ToLower(strings.TrimSpace(cfg.service))

	if cfg.sourceIP == "" {
		return config{}, errors.New("source-ip is required")
	}
	if net.ParseIP(cfg.sourceIP) == nil {
		return config{}, fmt.Errorf("source-ip %q is not a valid IP address", cfg.sourceIP)
	}
	if cfg.port < 1 || cfg.port > 65535 {
		return config{}, fmt.Errorf("port must be between 1 and 65535, got %d", cfg.port)
	}
	if cfg.service == "" {
		return config{}, errors.New("service is required")
	}
	return cfg, nil
}

func newFakeEvent(cfg config) model.Event {
	now := time.Now().UTC()

	return model.Event{
		ID:        fmt.Sprintf("evt-%d", now.UnixNano()),
		SourceIP:  cfg.sourceIP,
		Service:   cfg.service,
		Type:      eventTypeConnectionAttempt,
		Timestamp: now,
	}
}

func printEvent(event model.Event, port int) {
	fmt.Println("=== Fake Honeypot Event ===")
	fmt.Printf("ID: %s\n", event.ID)
	fmt.Printf("Source IP: %s\n", event.SourceIP)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Service: %s\n", event.Service)
	fmt.Printf("Type: %s\n", event.Type)
	fmt.Printf("Timestamp: %s\n", event.Timestamp.Format(time.RFC3339Nano))
}
