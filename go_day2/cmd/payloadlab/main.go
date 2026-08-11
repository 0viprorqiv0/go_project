package main

import (
	"fmt"

	"honeypot-day2/internal/counter"
	"honeypot-day2/internal/parser"
)

type inputEvent struct {
	IP      string
	Payload []byte
}

type parsedEvent struct {
	IP      string
	Payload []byte
	Command string
	Args    []string
	URL     string
	Count   int
}

func main() {
	inputs := []inputEvent{
		{IP: "192.0.2.10", Payload: []byte("wget http://example.com/a.sh")},
		{IP: "192.0.2.10", Payload: []byte("curl https://example.org/dropper")},
		{IP: "198.51.100.5", Payload: []byte("whoami")},
		{IP: "203.0.113.7", Payload: []byte("")},
		{IP: "", Payload: []byte("wget http://invalid.example/file")},
	}

	counts := counter.New()
	for _, input := range inputs {
		event, ok := processEvent(input, counts)
		if !ok {
			fmt.Printf("ignored: ip=%q payload=%q\n", input.IP, input.Payload)
			continue
		}

		fmt.Printf("ip=%s command=%s args=%v url=%q count=%d\n", event.IP, event.Command, event.Args, event.URL, event.Count)
	}
}

func processEvent(input inputEvent, counts *counter.EventCounter) (parsedEvent, bool) {
	if input.IP == "" {
		return parsedEvent{}, false
	}

	payload := parser.ClonePayload(input.Payload)
	command, ok := parser.ParseCommand(string(payload))
	if !ok {
		return parsedEvent{}, false
	}

	event := parsedEvent{
		IP:      input.IP,
		Payload: payload,
		Command: command.Name,
		Args:    command.Args,
		Count:   counts.Add(input.IP),
	}

	if targetURL, found := parser.ExtractURL(command); found {
		event.URL = targetURL
	}
	return event, true
}
