package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/snipbox/cli/internal/storage"
)

func TestNewStartsInGitCommandBrowser(t *testing.T) {
	commands := []storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git branch"},
	}
	model := New(commands, "")
	if model.Mode != "tree" || model.Tab != "Git" {
		t.Fatalf("empty query mode/tab = %q/%q, want tree/Git", model.Mode, model.Tab)
	}
	rows := model.treeRows()
	commandRows := 0
	for _, row := range rows {
		if row.Kind == "category" {
			t.Fatal("Git tab should not repeat the category in its tree")
		}
		if row.Kind == "command" {
			commandRows++
		}
	}
	if commandRows != len(commands) {
		t.Fatalf("Git tree exposes %d commands at startup; want all %d", commandRows, len(commands))
	}
	if row, ok := model.selectedTreeRow(); !ok || row.Kind != "command" {
		t.Fatalf("startup selection = %#v/%v; want the first command", row, ok)
	}
	if model := New(nil, "git"); model.Mode != "tree" {
		t.Fatalf("query mode = %q, want tree", model.Mode)
	}
}

func TestStartupScreenExplainsWhatSnipboxDoes(t *testing.T) {
	model := New([]storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Description: "Review repository changes.", Command: "git status"}}, "")
	model.Width, model.Height = 70, 18
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "A personal library for reusable terminal commands") {
		t.Fatal("startup view should explain Snipbox's purpose without opening a separate welcome screen")
	}
	if lipgloss.Height(model.headerView()) > 5 {
		t.Fatalf("purpose line made the header too tall: %d rows", lipgloss.Height(model.headerView()))
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

func TestAllTreeCanExpandAndCollapseCategories(t *testing.T) {
	commands := []storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}, {Name: "Branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git branch"}}
	model := New(commands, "")
	model.Mode, model.Tab = "tree", "All"
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

func TestCategoryTabDoesNotRepeatItsNameInTheTreeAndDetails(t *testing.T) {
	commands := []storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git branch"},
	}
	model := New(commands, "")
	model.Mode, model.Tab = "tree", "Git"
	rows := model.treeRows()
	if len(rows) == 0 || rows[0].Kind != "subcategory" {
		t.Fatalf("Git tab should start at subcategories instead of repeating the tab name: %#v", rows)
	}
	if details := ansi.Strip(model.detailsPanel(56, 8)); strings.Contains(details, "Git") {
		t.Fatalf("Git details should not repeat the active tab name: %q", details)
	}
}

func TestTreeEnterStartsCopyVariableFlowAndRunRemainsSeparate(t *testing.T) {
	commands := []storage.Command{{Name: "Checkout branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch {{branch:main}}"}}
	model := New(commands, "")
	model.Mode, model.Tab = "tree", "Git"
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

func TestTreeUsageCountAndCompactFooter(t *testing.T) {
	commands := []storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status", Usage: 14}}
	model := New(commands, "")
	model.Mode, model.Tab = "tree", "Git"
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
	if got := len(keys.ShortHelp()); got != 5 {
		t.Fatalf("short help has %d bindings, want 5", got)
	}
	helpView := helpFooter(help.New(), 44)
	for _, desc := range []string{"copy", "actions", "categories", "help", "quit"} {
		if !strings.Contains(helpView, desc) {
			t.Errorf("responsive footer omitted %q: %q", desc, helpView)
		}
	}
	for _, desc := range []string{"previous tab", "favorite", "delete", "history", "about"} {
		if strings.Contains(helpView, desc) {
			t.Errorf("compact footer repeats secondary action %q: %q", desc, helpView)
		}
	}
	for _, line := range strings.Split(helpView, "\n") {
		if lipgloss.Width(line) > 44 {
			t.Errorf("footer line width %d exceeds 44: %q", lipgloss.Width(line), line)
		}
	}
}

func TestActionsMenuOpensOnlyForSelectedCommandAndRunsFavoriteAction(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Width, model.Height = 80, 24
	model.subcategoryExpanded["Git\x00Basics & Staging"] = true
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(Model)
	if model.Modal != "actions" {
		t.Fatalf("M opened modal %q, want actions", model.Modal)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.Modal != "" || !model.Commands[0].Favorite {
		t.Fatalf("favorite menu action modal/favorite = %q/%t, want closed/true", model.Modal, model.Commands[0].Favorite)
	}
}

func TestActionsMenuDoesNotOpenDuringSearchAndFitsTerminal(t *testing.T) {
	model := New([]storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "git")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(Model)
	if model.Modal != "" || model.Query != "gitm" {
		t.Fatalf("M during search modal/query = %q/%q, want no modal and typed query", model.Modal, model.Query)
	}
	model = New([]storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Width, model.Height = 70, 18
	model.subcategoryExpanded["Git\x00Basics & Staging"] = true
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	model.Modal = "actions"
	view := model.View()
	if !strings.Contains(view, "Command Actions") || !strings.Contains(view, "[C] Copy to clipboard") || !strings.Contains(view, "[R] Run in terminal") {
		t.Fatalf("actions modal is missing its title or choices: %q", view)
	}
	if lipgloss.Height(view) > model.Height {
		t.Fatalf("actions modal height = %d, terminal height = %d", lipgloss.Height(view), model.Height)
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > model.Width {
			t.Fatalf("actions modal line width = %d, terminal width = %d", lipgloss.Width(line), model.Width)
		}
	}
}

func TestRunShowsAReviewBeforeSafeCommandExecution(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	model = updated.(Model)
	if model.Modal != "preview_run" || model.pending != "git status" || model.Commands[0].Usage != 0 {
		t.Fatalf("run preview state = modal:%q command:%q usage:%d", model.Modal, model.pending, model.Commands[0].Usage)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.Modal != "" || model.Commands[0].Usage != 0 {
		t.Fatalf("canceling run preview state = modal:%q usage:%d", model.Modal, model.Commands[0].Usage)
	}
}

func TestRunPreviewEnterSchedulesExecutionAndRecordsUsage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil || model.Modal != "" || model.Commands[0].Usage != 1 {
		t.Fatalf("confirmed run state = modal:%q usage:%d cmd:%v", model.Modal, model.Commands[0].Usage, cmd)
	}
}

func TestRunPreviewShowsTheCommandShellAndWorkingDirectory(t *testing.T) {
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Width, model.Height, model.pending, model.Modal = 100, 30, "git status", "preview_run"
	view := ansi.Strip(model.modalView())
	for _, expected := range []string{"Run Command", "Shell:", "Working directory:", "git status", "Enter runs", "Esc cancels"} {
		if !strings.Contains(view, expected) {
			t.Errorf("run preview omits %q: %q", expected, view)
		}
	}
}

func TestRunPreviewFitsMinimumTerminalAndKeepsActionsVisible(t *testing.T) {
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Width, model.Height, model.pending, model.Modal = 70, 18, "git status --short --branch", "preview_run"
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "Enter runs the command") || !strings.Contains(view, "Esc cancels") {
		t.Fatalf("minimum-size run preview clipped its actions: %q", view)
	}
	if lipgloss.Height(view) > model.Height {
		t.Fatalf("minimum-size run preview height = %d, terminal height = %d", lipgloss.Height(view), model.Height)
	}
}

func TestDangerousCommandStillRequiresConfirmation(t *testing.T) {
	model := New([]storage.Command{{Name: "Delete files", Category: "Linux/Bash", Subcategory: "Files & Search", Command: "rm -rf /tmp/snipbox-demo"}}, "")
	model.Tab = "Linux/Bash"
	selectCommandRow(&model, 0)
	model.beginRun()
	if model.Modal != "confirm" || model.pending != "rm -rf /tmp/snipbox-demo" {
		t.Fatalf("dangerous run state = modal:%q command:%q", model.Modal, model.pending)
	}
}

func TestActionsMenuRunShortcutStartsVariableFlow(t *testing.T) {
	model := New([]storage.Command{{Name: "Checkout branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch {{branch:main}}"}}, "")
	for i, row := range model.treeRows() {
		if row.Kind == "command" {
			model.Selected = i
			break
		}
	}
	model.Modal = "actions"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)
	if model.Modal != "variables" || model.variableAction != "run" {
		t.Fatalf("Run action shortcut modal/action = %q/%q, want variable prompt/run", model.Modal, model.variableAction)
	}
}

func TestActionsMenuShowsSelectedCommandContext(t *testing.T) {
	model := New([]storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Modal = "actions"
	view := model.modalView()
	if !strings.Contains(view, "Working tree status") || !strings.Contains(view, "Git") {
		t.Fatalf("actions menu lacks command context: %q", view)
	}
	if strings.Contains(view, "Show the current branch") {
		t.Fatalf("actions menu repeats the command description: %q", view)
	}
}

func TestTreeEmptyStatesAndScopedHeaderCount(t *testing.T) {
	commands := []storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Containers", Category: "Docker", Subcategory: "Containers", Command: "docker ps"},
	}
	model := New(commands, "")
	model.Width, model.Height = 100, 24
	model.Tab = "Network"
	if view := ansi.Strip(model.commandList(60, 8)); !strings.Contains(view, "No commands in this category yet.") {
		t.Fatalf("empty category message = %q", view)
	}
	model.Tab, model.Query = "Git", "no-such-command"
	model.searchInput.SetValue(model.Query)
	if view := ansi.Strip(model.commandList(60, 8)); !strings.Contains(view, "No commands match your search.") {
		t.Fatalf("empty search message = %q", view)
	}
	model.Query = ""
	model.searchInput.SetValue("")
	if header := ansi.Strip(model.headerView()); !strings.Contains(header, "1 / 1 · Git") {
		t.Fatalf("category-scoped header count = %q", header)
	}
}

func TestTabAndSearchClearKeepSelectedCommandWhenAvailable(t *testing.T) {
	commands := []storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Commit", Category: "Git", Subcategory: "Basics & Staging", Command: "git commit"},
		{Name: "Processes", Category: "PowerShell", Subcategory: "Processes & Ports", Command: "Get-Process"},
		{Name: "Containers", Category: "Docker", Subcategory: "Containers", Command: "docker ps"},
	}
	model := New(commands, "")
	selectCommandRow(&model, 1)
	model.cycleTab(true)
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Commit" {
		t.Fatalf("All tab lost selected command: %#v/%t", command, ok)
	}
	model.Tab = "Git"
	model.cycleTab(false)
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Processes" {
		t.Fatalf("new category did not select its first command: %#v/%t", command, ok)
	}
	model.Tab = "Git"
	model.Query, model.Tab = "status", "Git"
	model.searchInput.SetValue("status")
	selectCommandRow(&model, 0)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Status" {
		t.Fatalf("clearing search lost selected command: %#v/%t", command, ok)
	}
}

func TestFormValidationKeepsDialogOpenAndMarksInvalidFields(t *testing.T) {
	model := New(nil, "")
	model.Width, model.Height, model.Modal = 80, 24, "form"
	model.form = []string{"Unknown", "", "", "", "", ""}
	model.formField = 5
	model.prepareFormInput()
	model.saveForm()
	if model.Modal != "form" {
		t.Fatalf("invalid form modal = %q, want form", model.Modal)
	}
	view := ansi.Strip(model.modalView())
	if !strings.Contains(view, "Use a listed section") || !strings.Contains(view, "Command is required") {
		t.Fatalf("form does not show field-level validation: %q", view)
	}
	model.Width, model.Height = 70, 18
	view = model.View()
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > model.Width {
			t.Fatalf("invalid form line width %d exceeds terminal width: %q", lipgloss.Width(line), line)
		}
	}
	if !strings.Contains(view, "Esc cancels") {
		t.Fatal("minimum-size form hides its cancel action after validation errors")
	}
}

func TestDuplicateCommandOffersOpeningExistingEntry(t *testing.T) {
	commands := []storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}
	model := New(commands, "")
	model.Width, model.Height = 80, 24
	model.form = []string{"Git", "Basics & Staging", "Status again", "Show status", "git", " git status "}
	model.formField, model.editIndex, model.Modal = 5, -1, "form"
	model.prepareFormInput()
	model.formInput.SetValue(model.form[5])
	model.saveForm()
	if model.Modal != "duplicate" || len(model.Commands) != 1 {
		t.Fatalf("duplicate save modal/count = %q/%d", model.Modal, len(model.Commands))
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.Modal != "" || model.Tab != "Git" || model.Query != "" {
		t.Fatalf("opening duplicate modal/tab/query = %q/%q/%q", model.Modal, model.Tab, model.Query)
	}
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Working tree status" {
		t.Fatalf("selected duplicate entry = %#v/%t", command, ok)
	}
}

func selectCommandRow(model *Model, commandIndex int) {
	for _, row := range model.treeRows() {
		if row.Kind == "subcategory" {
			model.subcategoryExpanded[row.Category+"\x00"+row.Subcategory] = true
		}
	}
	for i, row := range model.treeRows() {
		if row.Kind == "command" && row.Index == commandIndex {
			model.Selected = i
			return
		}
	}
}

func TestMouseClickSelectsCommandAndRefreshesDetails(t *testing.T) {
	commands := []storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Create branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch -c {{branch:topic}}"},
	}
	model := New(commands, "")
	model.Width, model.Height = 100, 30
	model.detailsView.Height = 4
	model.detailsView.SetContent(strings.Repeat("detail\n", 12))
	model.detailsView.SetYOffset(8)
	model.subcategoryExpanded["Git\x00Branches & Merging"] = true
	rows := model.treeRows()
	rowIndex := -1
	for i, row := range rows {
		if row.Kind == "command" && row.Index == 1 {
			rowIndex = i
			break
		}
	}
	if rowIndex < 0 {
		t.Fatal("test command is not visible in the tree")
	}
	headerHeight := lipgloss.Height(model.headerView())
	updated, _ := model.Update(tea.MouseMsg{X: 3, Y: headerHeight + 2 + rowIndex, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	model = updated.(Model)
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Create branch" {
		t.Fatalf("mouse-selected command = %#v/%t", command, ok)
	}
	if !strings.Contains(model.detailsContent(), "git switch -c") || model.Modal != "" {
		t.Fatalf("click did not update details without activating command: %q / modal %q", model.detailsContent(), model.Modal)
	}
	if model.detailsView.YOffset != 0 {
		t.Fatalf("selecting a command kept stale details scroll at offset %d", model.detailsView.YOffset)
	}
}

func TestRightClickOpensActionsForClickedCommand(t *testing.T) {
	model := New([]storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Create branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch -c topic"},
	}, "")
	model.Width, model.Height = 100, 30
	model.subcategoryExpanded["Git\x00Branches & Merging"] = true
	rows := model.treeRows()
	rowIndex := -1
	for i, row := range rows {
		if row.Kind == "command" && row.Index == 1 {
			rowIndex = i
			break
		}
	}
	if rowIndex < 0 {
		t.Fatal("test command is not visible in the tree")
	}
	headerHeight := lipgloss.Height(model.headerView())
	updated, _ := model.Update(tea.MouseMsg{X: 3, Y: headerHeight + 2 + rowIndex, Button: tea.MouseButtonRight, Action: tea.MouseActionPress})
	model = updated.(Model)
	if model.Modal != "actions" {
		t.Fatalf("right click opened modal %q, want actions", model.Modal)
	}
	if _, command, ok := model.selectedCommand(); !ok || command.Name != "Create branch" {
		t.Fatalf("right-clicked command = %#v/%t", command, ok)
	}
}

func TestTreeUsesCategoryColorAndVisibleSelectionPointer(t *testing.T) {
	previous := lipgloss.DefaultRenderer().ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)
	model := New([]storage.Command{
		{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Create branch", Category: "Git", Subcategory: "Branches & Merging", Command: "git switch -c topic"},
	}, "")
	model.Mode, model.Tab = "tree", "All"
	rendered := model.commandList(64, 8)
	if !strings.Contains(rendered, "48;2;255;158;100") {
		t.Fatalf("Git tree category did not render its Tokyo Night accent color: %q", rendered)
	}
	selected := ansi.Strip(model.commandList(64, 8))
	if !strings.Contains(selected, "▶") {
		t.Fatalf("selected row is missing its visible pointer: %q", selected)
	}
	if !strings.Contains(selected, "├─") || !strings.Contains(selected, "└─") {
		t.Fatalf("tree rows are missing visible branch connectors: %q", selected)
	}
}

func TestTreeSearchKeepsJAndKAvailableForTyping(t *testing.T) {
	model := New([]storage.Command{{Name: "Status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"}}, "")
	model.Mode = "tree"
	for _, char := range []rune{'j', 'k'} {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		model = updated.(Model)
	}
	if model.Query != "jk" {
		t.Fatalf("tree search query = %q, want jk", model.Query)
	}
}

func TestFilteredTreeCountsOnlyMatchingCommands(t *testing.T) {
	commands := []storage.Command{
		{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Command: "git status"},
		{Name: "Create commit", Category: "Git", Subcategory: "Basics & Staging", Command: "git commit"},
	}
	model := New(commands, "status")
	model.Width, model.Height = 100, 30
	rows := model.treeRows()
	if len(rows) == 0 || rows[0].Kind != "category" {
		t.Fatalf("search tree rows = %#v, want a matching category", rows)
	}
	list := model.commandList(30, 8)
	if !strings.Contains(list, "Git (1)") {
		t.Fatalf("filtered category count missing from tree: %q", list)
	}
	if details := model.detailsPanel(60, 8); !strings.Contains(details, "Select a subcategory") || strings.Contains(details, "1 command") {
		t.Fatalf("category details should guide navigation without repeating the tree count: %q", details)
	}
	model.Selected = 1
	if details := model.detailsPanel(60, 8); !strings.Contains(details, "1 command") {
		t.Fatalf("filtered subcategory count missing from details: %q", details)
	}
}

func TestSuccessfulHistoryImportClosesModalAndPersistsOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	commands := []storage.Command{{Name: "Existing", Category: "Git", Command: "git log -1"}}
	if err := storage.Save(commands); err != nil {
		t.Fatal(err)
	}
	model := New(commands, "")
	model.Modal, model.history = "history", []string{"git status"}
	model.historyView.Width, model.historyView.Height = 40, 5
	model.Selected = 0
	model.syncHistory()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.Modal != "history_preview" || len(model.Commands) != 1 {
		t.Fatalf("history review modal/count = %q/%d, want preview without import", model.Modal, len(model.Commands))
	}
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.Modal != "" || cmd == nil {
		t.Fatalf("successful import modal/cmd = %q/%v, want closed modal and status timer", model.Modal, cmd)
	}
	if len(model.Commands) != 2 || model.Commands[1].Command != "git status" {
		t.Fatalf("in-memory history imports = %#v, want the existing item plus one import", model.Commands)
	}
	commands, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	imported, existing := 0, false
	for _, command := range commands {
		if command.Name == "History - git status" {
			imported++
		}
		if command.Name == "Existing" {
			existing = true
		}
	}
	if imported != 1 || !existing {
		t.Fatalf("persisted import count/existing item = %d/%t", imported, existing)
	}
}

func TestFailedHistoryImportDoesNotKeepUnsavedCommandInMemory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".snipbox"), []byte("file blocks directory creation"), 0o600); err != nil {
		t.Fatal(err)
	}
	model := New(nil, "")
	model.Modal, model.history = "history", []string{"git status"}
	model.Selected = 0
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if len(model.Commands) != 0 || !strings.Contains(model.Status, "Could not save command") {
		t.Fatalf("failed import commands/status = %#v/%q, want no in-memory command and a save error", model.Commands, model.Status)
	}
}

func TestHistoryImportSkipsCommandAlreadyInLibrary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	commands := []storage.Command{{Name: "Working tree status", Category: "Git", Command: "git status"}}
	model := New(commands, "")
	model.Modal, model.history = "history", []string{"git status"}
	model.Selected = 0
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if len(model.Commands) != 1 || model.Modal != "history" || !strings.Contains(model.Status, "already in your library") || cmd == nil {
		t.Fatalf("duplicate import state = commands:%d modal:%q status:%q cmd:%v", len(model.Commands), model.Modal, model.Status, cmd)
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

func TestBrowsePanelsFillTerminalWidthWithoutClipping(t *testing.T) {
	for _, width := range []int{70, 80, 96, 120} {
		model := New([]storage.Command{{Name: "Git status", Category: "Git", Description: "Inspect the working tree", Command: "git status --short --branch"}}, "git")
		model.Width = width
		leftWidth, rightWidth, leftContent, rightContent := panelWidths(width)
		header := model.headerView()
		if got := lipgloss.Width(header); got != width {
			t.Errorf("header width at %d columns = %d, want %d", width, got, width)
		}
		if !strings.Contains(header, "Network") {
			t.Errorf("header at %d columns clips the final tab: %q", width, header)
		}
		if !strings.Contains(header, "1 / 1") {
			t.Errorf("header at %d columns clips the result counter: %q", width, header)
		}
		left := model.commandList(leftContent, 8)
		right := model.detailsPanel(rightContent, 8)
		if got := lipgloss.Width(left); got != leftWidth {
			t.Errorf("left panel width at %d columns = %d, want %d", width, got, leftWidth)
		}
		if got := lipgloss.Width(right); got != rightWidth {
			t.Errorf("right panel width at %d columns = %d, want %d", width, got, rightWidth)
		}
		if got := lipgloss.Width(left) + 2 + lipgloss.Width(right); got != width {
			t.Errorf("panels and gutter at %d columns use %d columns", width, got)
		}
		if got := lipgloss.Height(left); got != 10 {
			t.Errorf("left panel height at %d columns = %d, want 10 including borders", width, got)
		}
		if got := lipgloss.Height(right); got != 10 {
			t.Errorf("right panel height at %d columns = %d, want 10 including borders", width, got)
		}
	}
}

func TestGitCommandBrowserFitsMinimumTerminalSizeWithoutDashboard(t *testing.T) {
	commands := []storage.Command{{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Description: "Show the current branch and working tree.", Command: "git status"}}
	for _, size := range [][2]int{{70, 18}, {70, 24}, {70, 30}, {96, 30}, {120, 40}} {
		model := New(commands, "")
		model.Width, model.Height = size[0], size[1]
		view := model.View()
		for _, text := range []string{"SNIPBOX", "Commands", "Command Details", "Basics & Staging", "Working tree status"} {
			if !strings.Contains(view, text) {
				t.Errorf("Git command browser at %dx%d is missing %q", size[0], size[1], text)
			}
		}
		if strings.Contains(view, "Browse All Tree") || strings.Contains(view, "favorites") {
			t.Errorf("Git command browser at %dx%d includes dashboard-only content", size[0], size[1])
		}
		if got := lipgloss.Height(view); got > size[1] {
			t.Errorf("Git command browser height at %dx%d = %d", size[0], size[1], got)
		}
		for lineNo, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > size[0] {
				t.Errorf("Git command browser line %d at %dx%d is %d columns", lineNo+1, size[0], size[1], got)
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
	if !strings.Contains(ansi.Strip(before), "More details below") {
		t.Fatal("details panel should indicate that more content is available below")
	}
	model.detailsView.SetYOffset(10)
	after := model.detailsPanel(58, 8)
	if before == after {
		t.Fatal("scrolling the details viewport should change the visible panel content")
	}
	if !strings.Contains(ansi.Strip(after), "More details above and below") {
		t.Fatal("scrolled details panel should indicate content in both directions")
	}
}

func TestDeleteConfirmationNamesCommandAndEffect(t *testing.T) {
	model := New([]storage.Command{{Name: "Remove stale branch", Category: "Git", Command: "git branch -D {{branch:old}}"}}, "")
	model.pendingDelete, model.Modal = 0, "confirm_delete"
	view := model.modalView()
	if !strings.Contains(view, "Remove stale branch") || !strings.Contains(view, "permanently removed") {
		t.Fatalf("delete confirmation lacks command or consequence: %q", view)
	}
}

func TestRunStatusClearsAfterStandardNoticeDuration(t *testing.T) {
	model := New(nil, "")
	updated, cmd := model.Update(runMsg{output: "done"})
	if updated.(Model).Status == "" || cmd == nil {
		t.Fatal("completed command should show a timed status notification")
	}
	updated, cmd = model.Update(runMsg{err: os.ErrInvalid})
	if updated.(Model).Status == "" || cmd == nil {
		t.Fatal("failed command should show a timed status notification")
	}
}

func TestUsageRecorderPersistsSharedCopyAndRunCount(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Status", Category: "Git", Command: "git status"}}, "")
	if err := model.recordUsage(0); err != nil {
		t.Fatal(err)
	}
	saved, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	if model.Commands[0].Usage != 1 || saved[0].Usage != 1 {
		t.Fatalf("in-memory/persisted usage = %d/%d, want 1/1", model.Commands[0].Usage, saved[0].Usage)
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
	if got := lipgloss.Height(model.commandList(64, 5)); got != 7 {
		t.Fatalf("command panel height = %d, want 7 including borders", got)
	}
}

func TestModalsFitTerminalBounds(t *testing.T) {
	for _, modal := range []string{"help", "about", "form", "variables", "history", "history_preview", "duplicate", "preview_run", "confirm", "confirm_delete"} {
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
	if width := lipgloss.Width(box); width != 68 {
		t.Fatalf("modal box width = %d, want 68 with intact borders", width)
	}
}

func TestFormModalKeepsItsActionsVisibleAtMinimumTerminalSize(t *testing.T) {
	model := New(nil, "")
	model.Width, model.Height, model.Modal = 70, 18, "form"
	model.form = []string{"Git", "General", "New branch", "Create a branch", "git", "git switch -c topic"}
	model.prepareFormInput()
	view := model.View()
	if !strings.Contains(view, "Esc cancels") || !strings.Contains(view, "Ctrl+S saves") {
		t.Fatalf("minimum-size form modal clipped its actions: %q", view)
	}
	if got := lipgloss.Height(view); got > model.Height {
		t.Fatalf("form modal height = %d at %d rows", got, model.Height)
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
	model.Width, model.Height = 70, 18
	view := model.modalView()
	for _, text := range []string{"About Snipbox", "reliable terminal commands", "Rust fuzzy search", "Your data stays local", "History import and Git sync are opt-in", "right-click for command actions", "Created by Pavel Slepcov"} {
		if !strings.Contains(view, text) {
			t.Errorf("About is missing %q: %q", text, view)
		}
	}
	if view := ansi.Strip(model.View()); !strings.Contains(view, "Press Esc to close.") {
		t.Errorf("About should keep its close hint visible at the minimum terminal size: %q", view)
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

func TestCommandDetailsShowExecutionEnvironment(t *testing.T) {
	model := New([]storage.Command{{
		Name: "Inspect DNS", Category: "Network", Subcategory: "DNS", Command: "dig example.com",
		Shells: []string{"Bash", "Zsh"}, Platforms: []string{"Linux", "macOS"}, Requirements: []string{"dig"},
	}}, "inspect")
	model.Mode, model.Selected = "browse", 0
	details := ansi.Strip(model.detailsContent())
	for _, expected := range []string{"Environment", "Shells: Bash, Zsh", "Platforms: Linux, macOS", "Tools: dig"} {
		if !strings.Contains(details, expected) {
			t.Errorf("command details omit %q: %q", expected, details)
		}
	}
}

func TestEditFormSavesRuntimeMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	model := New([]storage.Command{{Name: "Echo", Category: "Git", Subcategory: "Basics", Description: "Print a message", Command: "echo hello", Tags: []string{"demo"}}}, "")
	model.beginEdit()
	if len(model.form) != 9 {
		t.Fatalf("edit form has %d fields, want 9", len(model.form))
	}
	model.form[6], model.form[7], model.form[8] = "PowerShell, Bash", "Windows, Linux", "Git"
	model.formField = 8
	model.saveForm()
	saved, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) == 0 || strings.Join(saved[0].Shells, ",") != "PowerShell,Bash" || strings.Join(saved[0].Platforms, ",") != "Windows,Linux" || strings.Join(saved[0].Requirements, ",") != "Git" {
		t.Fatalf("saved execution metadata = %+v", saved)
	}
}

func TestSmallFormKeepsFocusedRuntimeFieldVisible(t *testing.T) {
	model := New(nil, "")
	model.Width, model.Height, model.Modal = 70, 18, "form"
	model.form = []string{"Git", "Basics", "Name", "Description", "git", "echo ok", "Bash", "Linux", "Git"}
	model.formField = 8
	model.prepareFormInput()
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "Required tools (comma separated)") || !strings.Contains(view, "Tab moves forward") {
		t.Fatalf("small form hid its focused field or actions: %q", view)
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
