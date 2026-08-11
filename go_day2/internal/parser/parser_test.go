package parser

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Command
		wantValid bool
	}{
		{
			name:      "wget with URL",
			input:     "wget http://example.com/a.sh",
			want:      Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantValid: true,
		},
		{
			name:      "extra whitespace",
			input:     "  \t wget   http://example.com/a.sh \r\n",
			want:      Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantValid: true,
		},
		{
			name:      "command without arguments",
			input:     "whoami",
			want:      Command{Name: "whoami", Args: []string{}},
			wantValid: true,
		},
		{
			name:      "empty input",
			input:     "",
			want:      Command{},
			wantValid: false,
		},
		{
			name:      "whitespace only",
			input:     " \t\r\n",
			want:      Command{},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := ParseCommand(tt.input)

			if valid != tt.wantValid {
				t.Fatalf("valid=%v, want %v", valid, tt.wantValid)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name      string
		command   Command
		wantURL   string
		wantFound bool
	}{
		{
			name:      "HTTP URL",
			command:   Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantURL:   "http://example.com/a.sh",
			wantFound: true,
		},
		{
			name:      "HTTPS URL after option",
			command:   Command{Name: "curl", Args: []string{"-L", "https://example.com/a"}},
			wantURL:   "https://example.com/a",
			wantFound: true,
		},
		{
			name:      "no URL",
			command:   Command{Name: "whoami"},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, found := ExtractURL(tt.command)

			if found != tt.wantFound {
				t.Fatalf("found=%v, want %v", found, tt.wantFound)
			}

			if gotURL != tt.wantURL {
				t.Fatalf("URL=%q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

func TestClonePayload(t *testing.T) {
	original := []byte("attack")
	cloned := ClonePayload(original)

	original[0] = 'X'

	if string(cloned) != "attack" {
		t.Fatalf("clone changed after source mutation: %q", cloned)
	}
}

func TestClonePayloadNil(t *testing.T) {
	if got := ClonePayload(nil); got != nil {
		t.Fatalf("ClonePayload(nil)=%v, want nil", got)
	}
}
