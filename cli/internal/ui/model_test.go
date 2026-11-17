package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/storage"
)

func TestNewStartsOnWelcomeOnlyWithoutQuery(t *testing.T) {
	if model := New(nil, ""); model.Mode != "welcome" {
		t.Fatalf("empty query mode = %q, want welcome", model.Mode)
	}
	if model := New(nil, "git"); model.Mode != "tree" {
		t.Fatalf("query mode = %q, want tree", model.Mode)
	}
}

func TestWelcomeSelectionAndTreeExpansion(t *testing.T) {
	commands := []storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}, {Name: "Branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git branch"}}
	model := New(commands, "")
	model.enterWelcomeSelection(0)
	if model.Mode != "tree" || model.Tab != "Git" || !model.categoryExpanded["Git"] {
		t.Fatalf("welcome category selection = mode %q tab %q", model.Mode, model.Tab)
	}
	model.enterWelcomeSelection(len(storage.Categories()))
	if model.Mode != "tree" {
		t.Fatalf("all-tree selection mode = %q", model.Mode)
	}
	rows := model.treeRows()
	if len(rows) == 0 || rows[0].Kind != "category" {
		t.Fatalf("tree starts with unexpected rows: %#v", rows)
	}
	model.categoryExpanded[rows[0].Category] = false
	collapsed := model.treeRows()
	model.categoryExpanded[rows[0].Category] = true
	expanded := model.treeRows()
	if len(expanded) <= len(collapsed) || expanded[1].Kind != "subcategory" {
		t.Fatal("expanding category should reveal subcategories")
	}
	model.Query = "branch"
	if got := len(model.treeRows()); got == 0 {
		t.Fatal("search should expand matching branches")
	}
}

func TestTreeEnterStartsCopyVariableFlowAndRunRemainsSeparate(t *testing.T) {
	commands := []storage.Command{{Name: "Checkout branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch {{branch:main}}"}}
	model := New(commands, "")
	model.enterWelcomeSelection(0)
	model.subcategoryExpanded["Git\x00Branches & Merging"] = true
	rows := model.treeRows()
	for i, row := range rows {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	_ = model.beginCopy()
	if model.Modal != "variables" || model.variableAction != "copy" {
		t.Fatalf("Enter copy flow modal/action = %q/%q", model.Modal, model.variableAction)
	}
	model.Modal = ""
	_ = model.beginRun()
	if model.Modal != "variables" || model.variableAction != "run" {
		t.Fatalf("Ctrl+R run flow modal/action = %q/%q", model.Modal, model.variableAction)
	}
}

func TestTreeUsageCountAndResponsiveFooterIncludeAllActions(t *testing.T) {
	commands := []storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status", Usage: 14}}
	model := New(commands, "")
	model.enterWelcomeSelection(0)
	model.subcategoryExpanded["Git\x00Basics & Staging"] = true
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	if rendered := model.commandList(64, 8); !strings.Contains(rendered, "14x") {
		t.Fatalf("tree command row omitted usage count: %q", rendered)
	}
	if got := len(keys.ShortHelp()); got != 9 {
		t.Fatalf("short help has %d bindings, want 9", got)
	}
	helpView := helpFooter(help.New(), 44)
	for _, desc := range []string{"copy", "run", "tabs", "new", "edit", "history", "about", "help", "quit"} {
		if !strings.Contains(helpView, desc) {
			t.Errorf("responsive footer omitted %q: %q", desc, helpView)
		}
	}
	for _, line := range strings.Split(helpView, "\n") {
		if lipgloss.Width(line) > 44 {
			t.Errorf("footer line width %d exceeds 44: %q", lipgloss.Width(line), line)
		}
	}
}

func TestViewFitsTerminalBounds(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {96, 30}, {80, 24}} {
		model := New([]storage.Command{{Name: "Git status", Category: "Git", Description: "Inspect the working tree", Command: "git status --short --branch"}}, "git")
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

func TestModalsFitTerminalBounds(t *testing.T) {
	for _, modal := range []string{"help", "about", "form", "variables", "history", "confirm"} {
		model := New(nil, "")
		model.Width, model.Height, model.Modal = 96, 28, modal
		model.form = []string{"Linux/Bash", "General", "name", "description", "tag", "command"}
		model.variables, model.values = []string{"path"}, map[string]string{"path": "/tmp"}
		model.history = []string{"git status"}
		model.pending = "git reset --hard"
		model.prepareFormInput()
		model.prepareVariableInput()
		model.syncHistory()
		view := model.View()
		if got := lipgloss.Height(view); got > model.Height {
			t.Errorf("%s modal height = %d, terminal height = %d", modal, got, model.Height)
		}
		for lineNo, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > model.Width {
				t.Errorf("%s modal line %d width = %d, terminal width = %d", modal, lineNo+1, got, model.Width)
			}
		}
	}
}

func TestTinyTerminalShowsResizeHint(t *testing.T) {
	for _, size := range [][2]int{{60, 14}, {24, 8}} {
		model := New(nil, "")
		model.Width, model.Height = size[0], size[1]
		view := model.View()
		if !strings.Contains(view, "Resize the terminal") {
			t.Fatal("small terminals should show a resize hint")
		}
		if got := lipgloss.Height(view); got > size[1] {
			t.Errorf("tiny-terminal hint height = %d, terminal height = %d", got, size[1])
		}
		for _, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > size[0] {
				t.Errorf("tiny-terminal line width = %d, terminal width = %d", got, size[0])
			}
		}
	}
}

func TestSearchUsesTextInput(t *testing.T) {
	model := New(nil, "")
	model.Mode = "browse"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g', 'i', 't'}})
	if got := updated.(Model).Query; got != "git" {
		t.Fatalf("search query = %q, want git", got)
	}
}

func TestHelpAndAboutUseEnglishText(t *testing.T) {
	model := New(nil, "")
	model.Modal = "help"
	if view := model.modalView(); !strings.Contains(view, "copy") || !strings.Contains(view, "quit") {
		t.Fatalf("help should list copy and quit bindings: %q", view)
	}
	model.Modal = "about"
	if view := model.modalView(); !strings.Contains(view, "Created by Pavel Slepcov") {
		t.Fatalf("About should identify the author: %q", view)
	}
}

func TestNoEmojiInBrandOrDangerWarning(t *testing.T) {
	model := New([]storage.Command{{Name: "Dangerous", Category: "Linux/Bash", Command: "rm -rf {{path:/tmp/demo}}"}}, "danger")
	model.Width, model.Height = 100, 30
	model.Tab = "Linux/Bash"
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	view := model.View()
	for _, symbol := range []string{"⚡", "⚠", "🔥"} {
		if strings.Contains(view, symbol) {
			t.Fatalf("UI contains forbidden emoji %q", symbol)
		}
	}
	if !strings.Contains(view, "[!]") {
		t.Fatal("dangerous commands should use a text warning chip")
	}
}

func TestFormInputIsInteractive(t *testing.T) {
	model := New(nil, "")
	model.Width, model.Height = 100, 30
	model.form = []string{"Git", "General", "", "", "", ""}
	model.Modal, model.formField = "form", 2
	model.prepareFormInput()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n', 'e', 'w'}})
	if got := updated.(Model).formInput.Value(); got != "new" {
		t.Fatalf("active form input = %q, want new", got)
	}
}

func TestCleanOutputRemovesTerminalControlSequences(t *testing.T) {
	if got := cleanOutput("\x1b[31mComplete\x1b[0m\r\n"); got != "Complete" {
		t.Fatalf("cleanOutput() = %q, want Complete", got)
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
