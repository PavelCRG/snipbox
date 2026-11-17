package ui

import (
	"strings"
	"testing"
)

func TestCleanOutputRemovesTerminalControlSequences(t *testing.T) {
	got := cleanOutput("\x1b[31mComplete\x1b[0m\r\n")
	if got != "Complete" {
		t.Fatalf("cleanOutput() = %q, want %q", got, "Complete")
	}
}

func TestCommandOutputIsCapped(t *testing.T) {
	var output cappedBuffer
	input := strings.Repeat("x", 70*1024)
	written, err := output.Write([]byte(input))
	if err != nil || written != len(input) {
		t.Fatalf("Write() = (%d, %v), want (%d, nil)", written, err, len(input))
	}
	if output.Len() != 64*1024 || !output.truncated {
		t.Fatalf("buffer has %d bytes and truncated=%t", output.Len(), output.truncated)
	}
}
