package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/storage"
)

func TestNewStartsInGitCollection(t *testing.T) {
	model := New(nil, "")
	if model.Tab != "Git" {
		t.Fatalf("initial tab = %q, want Git", model.Tab)
	}
}

func TestCommandPanelsRenderSideBySide(t *testing.T) {
	model := New([]storage.Command{{Name: "Git status", Category: "Git", Command: "git status"}}, "")
	model.Width, model.Height = 100, 30
	if len(model.matches()) != 1 {
		t.Fatalf("Git collection has %d matches; want 1", len(model.matches()))
	}
	view := model.View()
	if !strings.Contains(view, "Git status") || !strings.Contains(view, "git status") {
		t.Fatal("selected Git command and its details were not rendered")
	}
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Commands") && strings.Contains(line, "Command Details") {
			return
		}
	}
	t.Fatal("Commands and Command Details were not rendered on the same row")
}

func TestViewFitsTerminalBounds(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {96, 30}, {80, 24}} {
		model := New([]storage.Command{{Name: "Git status", Category: "Git", Description: "Inspect the working tree", Command: "git status --short --branch"}}, "")
		model.Width, model.Height = size[0], size[1]
		view := model.View()
		if got := lipgloss.Height(view); got > size[1] {
			t.Errorf("View() height = %d at %dx%d", got, size[0], size[1])
		}
		for lineNo, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > size[0] {
				t.Errorf("View() line %d width = %d at %dx%d", lineNo+1, got, size[0], size[1])
			}
		}
	}
}

func TestViewRequestsResizeForVerySmallTerminal(t *testing.T) {
	model := New(nil, "")
	model.Width, model.Height = 60, 20
	if view := model.View(); !strings.Contains(view, "Resize the terminal") {
		t.Fatalf("small terminal view should explain how to recover: %q", view)
	}
}

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
