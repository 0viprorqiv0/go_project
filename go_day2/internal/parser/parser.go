package parser

import "strings"

// Command is a parsed command name and its arguments.
type Command struct {
	Name string
	Args []string
}

// ParseCommand splits a payload into a command and arguments.
func ParseCommand(payload string) (Command, bool) {
	fields := strings.Fields(payload)
	if len(fields) == 0 {
		return Command{}, false
	}

	return Command{Name: fields[0], Args: fields[1:]}, true
}

// ExtractURL returns the first HTTP or HTTPS URL in the command arguments.
func ExtractURL(command Command) (string, bool) {
	for _, arg := range command.Args {
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			return arg, true
		}
	}

	return "", false
}

// ClonePayload returns an independent copy of payload.
func ClonePayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}
