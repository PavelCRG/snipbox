package ui

import (
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/storage"
)

func TestNewStartsWithSplashOnlyWithoutQuery(t *testing.T) {
	if model := New(nil, ""); model.Mode != "splash" {
		t.Fatalf("empty query mode = %q, want splash", model.Mode)
	}
	if model := New(nil, "git"); model.Mode != "tree" {
		t.Fatalf("query mode = %q, want tree", model.Mode)
	}
}

func TestSplashCanBeSkippedAndAdvancesToWelcome(t *testing.T) {
	model := New(nil, "")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := updated.(Model).Mode; got != "welcome" {
		t.Fatalf("mode after splash key = %q, want welcome", got)
	}
	model = New(nil, "")
	updated, cmd := model.Update(splashTickMsg{})
	if got := updated.(Model).SplashFrame; got != 1 || cmd == nil {
		t.Fatalf("splash tick frame/cmd = %d/%v, want 1 and next tick", got, cmd)
	}
	model.SplashFrame = splashFinalFrame - 1
	updated, cmd = model.Update(splashTickMsg{})
	if got := updated.(Model).Mode; got != "welcome" || cmd != nil {
		t.Fatalf("final splash frame mode/cmd = %q/%v, want welcome and no timer", got, cmd)
	}
}

func TestQuestionMarkSearchesUnlessSearchIsEmpty(t *testing.T) {
	model := New(nil, "git")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if got := updated.(Model); got.Modal != "" || got.Query != "git?" {
		t.Fatalf("? with a query = modal %q, query %q; want typed question mark", got.Modal, got.Query)
	}
	model = New(nil, "")
	model.Mode = "tree"
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if got := updated.(Model); got.Modal != "help" || got.Query != "" {
		t.Fatalf("? with an empty query = modal %q, query %q; want help", got.Modal, got.Query)
	}
}

func TestEscapeClearsQueryBeforeQuitting(t *testing.T) {
	model := New(nil, "git")
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if got := updated.(Model); got.Query != "" || cmd != nil {
		t.Fatalf("first Esc query/cmd = %q/%v; want cleared query without quitting", got.Query, cmd)
	}
}

func TestShiftTabMovesToPreviousCategory(t *testing.T) {
	model := New(nil, "")
	model.Mode, model.Tab = "tree", "Git"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := updated.(Model).Tab; got != "All" {
		t.Fatalf("Shift+Tab category = %q, want All", got)
	}
}

func TestControlFTogglesFavorite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "status")
	model.Selected = 2
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	if got := updated.(Model).Commands[0].Favorite; !got {
		t.Fatal("Ctrl+F should favorite the selected command")
	}
}

func TestControlDRequiresConfirmationBeforeDeleting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "status")
	model.Selected = 2
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	model = updated.(Model)
	if model.Modal != "confirm_delete" || len(model.Commands) != 1 {
		t.Fatalf("Ctrl+D modal/count = %q/%d, want confirmation and unchanged command", model.Modal, len(model.Commands))
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updated.(Model)
	if model.Modal != "" || len(model.Commands) != 0 {
		t.Fatalf("confirmed deletion modal/count = %q/%d, want closed modal and no command", model.Modal, len(model.Commands))
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
	if got := len(keys.ShortHelp()); got != 13 {
		t.Fatalf("short help has %d bindings, want 13", got)
	}
	helpView := helpFooter(help.New(), 44)
	for _, desc := range []string{"copy", "run", "tabs", "previous tab", "new", "edit", "favorite", "delete", "history", "about", "help", "welcome", "quit"} {
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
	for _, size := range [][2]int{{120, 40}, {96, 30}, {80, 24}, {70, 18}} {
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

func TestSplashFitsTerminalBounds(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}, {70, 18}} {
		model := New(nil, "")
		model.Width, model.Height, model.SplashFrame = size[0], size[1], 39
		view := model.View()
		if !strings.Contains(view, "SNIPBOX") || !strings.Contains(view, "Ready") {
			t.Fatalf("splash is missing its final frame content: %q", view)
		}
		if got := lipgloss.Height(view); got > size[1] {
			t.Errorf("splash height = %d at %dx%d", got, size[0], size[1])
		}
		for line, content := range strings.Split(view, "\n") {
			if got := lipgloss.Width(content); got > size[0] {
				t.Errorf("splash line %d width = %d at %dx%d", line+1, got, size[0], size[1])
			}
		}
	}
}

func TestDetailsPanelUsesScrollableViewport(t *testing.T) {
	description := strings.Join([]string{"detail line"}, "\n")
	for i := 0; i < 30; i++ {
		description += "\ndetail line"
	}
	model := New([]storage.Command{{Name: "Long details", Category: "Git", Subcategory: "Basics & Staging", Description: description, Command: "git status"}}, "details")
	model.Selected = 2
	model.detailsView.Width, model.detailsView.Height = 54, 6
	model.syncDetails()
	before := model.detailsPanel(58, 8)
	model.detailsView.SetYOffset(10)
	after := model.detailsPanel(58, 8)
	if before == after {
		t.Fatal("scrolling the details viewport should change the visible panel content")
	}
}

func TestTreeSelectionUsesItsOwnSearchHighlightsAndFitsPanelRows(t *testing.T) {
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "status")
	model.Mode = "tree"
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	if got := model.selectedIndices(); len(got) == 0 {
		t.Fatal("selected tree command should retain the match indices for highlighting")
	}
	if got := lipgloss.Height(model.commandList(64, 5)); got > 5 {
		t.Fatalf("command panel height = %d, want at most 5", got)
	}
}

func TestModalsFitTerminalBounds(t *testing.T) {
	for _, modal := range []string{"help", "about", "form", "variables", "history", "confirm", "confirm_delete"} {
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

func TestModalWidthIsCappedAt68Columns(t *testing.T) {
	box := modalBox(68, strings.Repeat("Long modal content ", 12))
	if width := lipgloss.Width(box); width > 68 {
		t.Fatalf("modal box width = %d, want at most 68", width)
	}
}

func TestTinyTerminalShowsResizeHint(t *testing.T) {
	for _, size := range [][2]int{{69, 18}, {70, 17}, {60, 14}, {24, 8}} {
		model := New(nil, "")
		model.Width, model.Height = size[0], size[1]
		view := model.View()
		if !strings.Contains(view, "Terminal") {
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

func TestShellCommandUsesPlatformShell(t *testing.T) {
	cmd := newShellCommand("echo snipbox")
	if cmd == nil || len(cmd.Args) == 0 {
		t.Fatal("shell command should be constructed")
	}
	if runtime.GOOS == "windows" {
		if cmd.Args[len(cmd.Args)-2] != "-Command" || cmd.Args[len(cmd.Args)-1] != "echo snipbox" {
			t.Fatalf("Windows shell args = %#v", cmd.Args)
		}
	} else if len(cmd.Args) < 3 || cmd.Args[len(cmd.Args)-2] != "-c" || cmd.Args[len(cmd.Args)-1] != "echo snipbox" {
		t.Fatalf("Unix shell args = %#v", cmd.Args)
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
