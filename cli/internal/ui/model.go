package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/bridge"
	"github.com/snipbox/cli/internal/history"
	"github.com/snipbox/cli/internal/storage"
)

type searchHit struct {
	Index   int   `json:"index"`
	Indices []int `json:"indices"`
}

type templateInfo struct {
	Variables []struct {
		Name    string `json:"name"`
		Default string `json:"default"`
	} `json:"variables"`
	Dangerous bool `json:"dangerous"`
}

type runMsg struct {
	output   string
	err      error
	usageErr string
}

type clearStatusMsg string

type treeRow struct {
	Kind        string
	Category    string
	Subcategory string
	Index       int
	Indices     []int
	MatchCount  int
}

type keyMap struct {
	Up, Down, PageUp, PageDown, Copy, Run, NextTab, PreviousTab, New, Edit, Favorite, Delete, History, Actions, Help, About, Quit key.Binding
}

var actionKeys = []string{"c", "r", "f", "e", "d"}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Copy, k.Actions, k.NextTab, k.Help, k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.PageUp, k.PageDown}, {k.Copy, k.Run, k.Actions, k.NextTab, k.PreviousTab}, {k.New, k.Edit, k.Favorite, k.Delete, k.History, k.Help, k.About, k.Quit}}
}

var keys = keyMap{
	Up:          key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "move")),
	Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "move")),
	PageUp:      key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "scroll details")),
	PageDown:    key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "scroll details")),
	Copy:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "copy")),
	Run:         key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "run")),
	NextTab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "categories")),
	PreviousTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous tab")),
	New:         key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "new")),
	Edit:        key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "edit")),
	Favorite:    key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "favorite")),
	Delete:      key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
	History:     key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "history")),
	Actions:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "actions")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	About:       key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "about")),
	Quit:        key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
}

type cappedBuffer struct {
	bytes.Buffer
	truncated bool
}

var terminalEscape = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07]*(?:\x07|\x1b\\))`)

type Model struct {
	Commands []storage.Command
	Query    string
	Selected int
	Tab      string
	Mode     string
	Status   string
	Modal    string
	Width    int
	Height   int

	searchInput   textinput.Model
	formInput     textinput.Model
	variableInput textinput.Model
	detailsView   viewport.Model
	historyView   viewport.Model
	helpView      help.Model

	form                []string
	formErrors          map[int]string
	formField           int
	editIndex           int
	duplicateIndex      int
	history             []string
	pendingImport       string
	variables           []string
	values              map[string]string
	varField            int
	actionSelected      int
	pending             string
	variableAction      string
	pendingDelete       int
	categoryExpanded    map[string]bool
	subcategoryExpanded map[string]bool
}

func New(commands []storage.Command, query string) Model {
	search := textinput.New()
	search.Prompt = "> Search: "
	search.Placeholder = "type to filter commands"
	search.SetValue(query)
	search.Focus()
	search.PromptStyle = Selected
	search.TextStyle = Base
	search.Cursor.Style = Selected
	mode := "tree"
	tab := "Git"
	if query != "" {
		mode = "tree"
		tab = "All"
	}
	expanded := make(map[string]bool)
	for _, category := range storage.Categories() {
		expanded[category] = false
	}
	subcategoryExpanded := make(map[string]bool)
	if query == "" {
		expanded["Git"] = true
		for _, subcategory := range storage.Subcategories(commands, "Git") {
			subcategoryExpanded["Git\x00"+subcategory] = true
		}
	}
	model := Model{
		Commands: commands, Query: query, Tab: tab, Mode: mode, editIndex: -1,
		pendingDelete: -1, duplicateIndex: -1,
		categoryExpanded: expanded, subcategoryExpanded: subcategoryExpanded,
		searchInput: search, formInput: textinput.New(), variableInput: textinput.New(),
		detailsView: viewport.New(1, 1), historyView: viewport.New(1, 1), helpView: help.New(),
	}
	if query == "" {
		for index, row := range model.treeRows() {
			if row.Kind == "command" {
				model.Selected = index
				break
			}
		}
	}
	return model
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) matches() []searchHit {
	result, err := bridge.Search(m.Commands, m.Query)
	if err != nil {
		return nil
	}
	var hits []searchHit
	if json.Unmarshal([]byte(result), &hits) != nil {
		return nil
	}
	filtered := hits[:0]
	for _, hit := range hits {
		if hit.Index >= 0 && hit.Index < len(m.Commands) && (m.Tab == "All" || m.Commands[hit.Index].Category == m.Tab) {
			filtered = append(filtered, hit)
		}
	}
	return filtered
}

func (m Model) selectedCommand() (int, storage.Command, bool) {
	if m.Mode == "tree" {
		rows := m.treeRows()
		if m.Selected >= 0 && m.Selected < len(rows) && rows[m.Selected].Kind == "command" {
			index := rows[m.Selected].Index
			return index, m.Commands[index], true
		}
		return 0, storage.Command{}, false
	}
	hits := m.matches()
	if len(hits) == 0 {
		return 0, storage.Command{}, false
	}
	index := min(max(m.Selected, 0), len(hits)-1)
	commandIndex := hits[index].Index
	return commandIndex, m.Commands[commandIndex], true
}

func (m Model) treeRows() []treeRow {
	hits := m.matches()
	match := make(map[int][]int, len(hits))
	for _, hit := range hits {
		match[hit.Index] = hit.Indices
	}
	var rows []treeRow
	for _, category := range storage.Categories() {
		if m.Tab != "All" && m.Tab != category {
			continue
		}
		count := 0
		for index, command := range m.Commands {
			if command.Category == category {
				if _, ok := match[index]; ok {
					count++
				}
			}
		}
		if count == 0 {
			continue
		}
		if m.Tab == "All" {
			rows = append(rows, treeRow{Kind: "category", Category: category, Index: -1, MatchCount: count})
			if !m.categoryExpanded[category] && m.Query == "" {
				continue
			}
		}
		for _, subcategory := range storage.Subcategories(m.Commands, category) {
			subCount := 0
			for index, command := range m.Commands {
				if command.Category == category && command.Subcategory == subcategory {
					if _, ok := match[index]; ok {
						subCount++
					}
				}
			}
			if subCount == 0 {
				continue
			}
			rows = append(rows, treeRow{Kind: "subcategory", Category: category, Subcategory: subcategory, Index: -1, MatchCount: subCount})
			key := category + "\x00" + subcategory
			if !m.subcategoryExpanded[key] && m.Query == "" {
				continue
			}
			for index, command := range m.Commands {
				indices, ok := match[index]
				if ok && command.Category == category && command.Subcategory == subcategory {
					rows = append(rows, treeRow{Kind: "command", Category: category, Subcategory: subcategory, Index: index, Indices: indices})
				}
			}
		}
	}
	return rows
}

func (m Model) selectedTreeRow() (treeRow, bool) {
	rows := m.treeRows()
	if m.Selected < 0 || m.Selected >= len(rows) {
		return treeRow{}, false
	}
	return rows[m.Selected], true
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = v.Width, v.Height
		m.layout()
		m.syncDetails()
		m.syncHistory()
	case runMsg:
		if v.err != nil {
			message := "Command failed: " + v.err.Error() + " " + truncate(strings.TrimSpace(v.output), 100)
			if v.usageErr != "" {
				message += " · Usage count was not saved"
			}
			return m, m.failureCmd(strings.TrimSpace(message))
		}
		status := string(rune(0x2713)) + " Command completed"
		if output := strings.TrimSpace(v.output); output != "" {
			status += ": " + truncate(output, 100)
		}
		if v.usageErr != "" {
			status += " · Usage count was not saved"
		}
		return m, m.noticeCmd(status)
	case clearStatusMsg:
		if string(v) == m.Status {
			m.Status = ""
		}
	case tea.MouseMsg:
		return m.updateMouse(v)
	case tea.KeyMsg:
		if m.Modal != "" {
			return m.updateModal(v)
		}
		if m.Mode == "tree" {
			return m.updateTree(v)
		}
		switch v.String() {
		case "esc":
			if m.Query != "" {
				selected := m.selectedCommandIndex()
				m.Query = ""
				m.searchInput.SetValue("")
				m.restoreTreeSelection(selected)
				m.syncDetails()
				return m, nil
			}
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
				m.syncDetails()
			}
			return m, nil
		case "down", "j":
			if m.Selected < len(m.matches())-1 {
				m.Selected++
				m.syncDetails()
			}
			return m, nil
		case "pgup":
			m.detailsView, _ = m.detailsView.Update(v)
			return m, nil
		case "pgdown":
			m.detailsView, _ = m.detailsView.Update(v)
			return m, nil
		case "tab", "shift+tab":
			m.cycleTab(v.String() == "shift+tab")
			return m, nil
		case "?":
			if m.Query == "" {
				m.Modal = "help"
				return m, nil
			}
		case "ctrl+o":
			m.Modal = "about"
			return m, nil
		case "m":
			if m.Query == "" {
				return m, m.openActions()
			}
		case "enter":
			index, command, ok := m.selectedCommand()
			if ok {
				if err := clipboard.WriteAll(command.Command); err != nil {
					return m, m.failureCmd("Clipboard error: " + err.Error())
				} else {
					return m, m.recordCopy(index)
				}
			}
			return m, nil
		case "ctrl+r":
			cmd := m.beginRun()
			return m, cmd
		case "ctrl+a":
			category := m.Tab
			if category == "All" {
				category = "Git"
			}
			m.form = []string{category, "General", "", "", "", "", "", "", ""}
			m.formField, m.editIndex, m.Modal = 0, -1, "form"
			m.prepareFormInput()
			return m, nil
		case "ctrl+e":
			m.beginEdit()
			return m, nil
		case "ctrl+f":
			return m, m.toggleFavorite()
		case "ctrl+d":
			m.beginDelete()
			return m, nil
		case "ctrl+h":
			m.history = history.Recent()
			if len(m.history) == 0 {
				return m, m.failureCmd("No shell history found")
			} else {
				m.Selected, m.Modal = 0, "history"
				m.syncHistory()
			}
			return m, nil
		}
		selected := m.selectedCommandIndex()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(v)
		m.Query = m.searchInput.Value()
		if m.Query != "" {
			m.Mode = "tree"
		}
		m.restoreTreeSelection(selected)
		m.syncDetails()
		return m, cmd
	default:
		m.searchInput, _ = m.searchInput.Update(msg)
		m.detailsView, _ = m.detailsView.Update(msg)
		m.historyView, _ = m.historyView.Update(msg)
		m.helpView, _ = m.helpView.Update(msg)
		if m.Modal == "form" {
			m.formInput, _ = m.formInput.Update(msg)
			if m.formField < len(m.form) {
				m.form[m.formField] = m.formInput.Value()
			}
		} else if m.Modal == "variables" {
			m.variableInput, _ = m.variableInput.Update(msg)
			if m.varField < len(m.variables) {
				m.values[m.variables[m.varField]] = m.variableInput.Value()
			}
		}
	}
	return m, nil
}

func (m Model) updateMouse(mouse tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.Modal != "" || m.Width < 70 || m.Height < 18 || mouse.Action != tea.MouseActionPress || (mouse.Button != tea.MouseButtonLeft && mouse.Button != tea.MouseButtonRight) {
		return m, nil
	}
	leftWidth, _, _, _ := panelWidths(m.Width)
	if mouse.X < 0 || mouse.X >= leftWidth {
		return m, nil
	}
	headerHeight := lipgloss.Height(m.headerView())
	rowOffset := mouse.Y - headerHeight - 2
	if rowOffset < 0 {
		return m, nil
	}
	helpWidth := max(1, m.Width-4)
	helpHeight := strings.Count(helpFooter(m.helpView, helpWidth), "\n") + 1
	if m.Status != "" {
		helpHeight++
	}
	bodyHeight := max(3, m.Height-headerHeight-helpHeight)
	visibleRows := max(1, max(1, bodyHeight-2)-1)
	if rowOffset >= visibleRows {
		return m, nil
	}
	rows := m.treeRows()
	if m.Mode != "tree" {
		rows = nil
		for _, hit := range m.matches() {
			rows = append(rows, treeRow{Kind: "command", Index: hit.Index})
		}
	}
	index := listStart(m.Selected, visibleRows, len(rows)) + rowOffset
	if index >= len(rows) {
		return m, nil
	}
	m.Selected = index
	m.syncDetails()
	if mouse.Button == tea.MouseButtonRight {
		if _, _, ok := m.selectedCommand(); ok {
			return m, m.openActions()
		}
	}
	return m, nil
}

func (m Model) updateTree(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.treeRows()
	if len(rows) == 0 {
		m.Selected = 0
	} else if m.Selected >= len(rows) {
		m.Selected = len(rows) - 1
	}
	var row treeRow
	if len(rows) > 0 {
		row = rows[m.Selected]
	}
	switch key.String() {
	case "esc":
		if m.Query != "" {
			selected := m.selectedCommandIndex()
			m.Query = ""
			m.searchInput.SetValue("")
			m.restoreTreeSelection(selected)
			m.syncDetails()
			return m, nil
		}
		return m, tea.Quit
	case "ctrl+c":
		return m, tea.Quit
	case "up":
		if m.Selected > 0 {
			m.Selected--
			m.syncDetails()
		}
		return m, nil
	case "down":
		if m.Selected < len(rows)-1 {
			m.Selected++
			m.syncDetails()
		}
		return m, nil
	case "left":
		if row.Kind == "category" {
			m.categoryExpanded[row.Category] = false
		} else if row.Kind == "subcategory" {
			m.subcategoryExpanded[row.Category+"\x00"+row.Subcategory] = false
		} else if row.Kind == "command" {
			for i := m.Selected - 1; i >= 0; i-- {
				if rows[i].Kind == "subcategory" && rows[i].Category == row.Category && rows[i].Subcategory == row.Subcategory {
					m.Selected = i
					break
				}
			}
		}
		m.syncDetails()
		return m, nil
	case "right":
		if row.Kind == "category" {
			m.categoryExpanded[row.Category] = true
		} else if row.Kind == "subcategory" {
			m.subcategoryExpanded[row.Category+"\x00"+row.Subcategory] = true
		}
		m.syncDetails()
		return m, nil
	case " ", "space", "enter":
		if row.Kind == "category" {
			m.categoryExpanded[row.Category] = !m.categoryExpanded[row.Category]
		} else if row.Kind == "subcategory" {
			path := row.Category + "\x00" + row.Subcategory
			m.subcategoryExpanded[path] = !m.subcategoryExpanded[path]
		} else if row.Kind == "command" && key.String() == "enter" {
			return m, m.beginCopy()
		}
		m.syncDetails()
		return m, nil
	case "pgup", "pgdown":
		m.detailsView, _ = m.detailsView.Update(key)
		return m, nil
	case "tab", "shift+tab":
		m.cycleTab(key.String() == "shift+tab")
		return m, nil
	case "ctrl+r":
		cmd := m.beginRun()
		return m, cmd
	case "ctrl+a":
		category := m.Tab
		if category == "All" {
			category = "Git"
		}
		m.form = []string{category, "General", "", "", "", "", "", "", ""}
		m.formField, m.editIndex, m.Modal = 0, -1, "form"
		m.prepareFormInput()
		return m, nil
	case "ctrl+e":
		m.beginEdit()
		return m, nil
	case "ctrl+f":
		return m, m.toggleFavorite()
	case "ctrl+d":
		m.beginDelete()
		return m, nil
	case "ctrl+h":
		m.history = history.Recent()
		if len(m.history) == 0 {
			return m, m.failureCmd("No shell history found")
		} else {
			m.Selected, m.Modal = 0, "history"
			m.syncHistory()
		}
		return m, nil
	case "?":
		if m.Query == "" {
			m.Modal = "help"
			return m, nil
		}
	case "ctrl+o":
		m.Modal = "about"
		return m, nil
	case "m":
		if m.Query == "" {
			return m, m.openActions()
		}
	}
	selected := m.selectedCommandIndex()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(key)
	m.Query = m.searchInput.Value()
	if m.Query != "" {
		m.Mode = "tree"
	}
	m.restoreTreeSelection(selected)
	m.syncDetails()
	return m, cmd
}

func (m *Model) cycleTab(previous bool) {
	selected := m.selectedCommandIndex()
	tabs := append([]string{"All"}, storage.Categories()...)
	for i, tab := range tabs {
		if tab == m.Tab {
			step := 1
			if previous {
				step = len(tabs) - 1
			}
			m.Tab = tabs[(i+step)%len(tabs)]
			break
		}
	}
	m.restoreTreeSelection(selected)
	m.syncDetails()
}

func (m Model) selectedCommandIndex() int {
	if index, _, ok := m.selectedCommand(); ok {
		return index
	}
	return -1
}

func (m *Model) restoreTreeSelection(preferred int) {
	if preferred >= 0 && preferred < len(m.Commands) && (m.Tab == "All" || m.Commands[preferred].Category == m.Tab) {
		m.selectTreeCommand(preferred)
		if row, ok := m.selectedTreeRow(); ok && row.Kind == "command" && row.Index == preferred {
			return
		}
	}
	for _, hit := range m.matches() {
		m.selectTreeCommand(hit.Index)
		if row, ok := m.selectedTreeRow(); ok && row.Kind == "command" && row.Index == hit.Index {
			return
		}
	}
	m.Selected = 0
}

func (m *Model) selectTreeCommand(index int) {
	if index < 0 || index >= len(m.Commands) {
		return
	}
	command := m.Commands[index]
	if m.Tab == "All" {
		m.categoryExpanded[command.Category] = true
	}
	m.subcategoryExpanded[command.Category+"\x00"+command.Subcategory] = true
	for i, row := range m.treeRows() {
		if row.Kind == "command" && row.Index == index {
			m.Selected = i
			return
		}
	}
}

func (m *Model) toggleFavorite() tea.Cmd {
	index, _, ok := m.selectedCommand()
	if !ok {
		return nil
	}
	m.Commands[index].Favorite = !m.Commands[index].Favorite
	if err := storage.Save(m.Commands); err != nil {
		m.Commands[index].Favorite = !m.Commands[index].Favorite
		return m.failureCmd("Could not save favorite: " + err.Error())
	}
	if m.Commands[index].Favorite {
		return m.successCmd("[OK] Added to favorites")
	}
	return m.successCmd("[OK] Removed from favorites")
}

func (m *Model) beginDelete() {
	index, _, ok := m.selectedCommand()
	if !ok {
		return
	}
	m.pendingDelete = index
	m.Modal = "confirm_delete"
}

func (m *Model) beginEdit() {
	if index, command, ok := m.selectedCommand(); ok {
		m.form = []string{command.Category, command.Subcategory, command.Name, command.Description, strings.Join(command.Tags, ", "), command.Command, strings.Join(command.Shells, ", "), strings.Join(command.Platforms, ", "), strings.Join(command.Requirements, ", ")}
		m.formField, m.editIndex, m.Modal = 0, index, "form"
		m.prepareFormInput()
	}
}

func (m *Model) openActions() tea.Cmd {
	if _, _, ok := m.selectedCommand(); !ok {
		return m.failureCmd("Select a command to view its actions")
	}
	m.actionSelected, m.Modal = 0, "actions"
	return nil
}

func (m Model) actionLabels() []string {
	_, command, ok := m.selectedCommand()
	if !ok {
		return nil
	}
	favorite := "Add to favorites"
	if command.Favorite {
		favorite = "Remove from favorites"
	}
	return []string{"Copy to clipboard", "Run in terminal", favorite, "Edit command", "Delete command"}
}

func (m Model) runSelectedAction() (tea.Model, tea.Cmd) {
	m.Modal = ""
	switch m.actionSelected {
	case 0:
		return m, m.beginCopy()
	case 1:
		return m, m.beginRun()
	case 2:
		return m, m.toggleFavorite()
	case 3:
		m.beginEdit()
	case 4:
		m.beginDelete()
	}
	return m, nil
}

func (m *Model) deletePending() tea.Cmd {
	if m.pendingDelete < 0 || m.pendingDelete >= len(m.Commands) {
		m.Modal = ""
		m.pendingDelete = -1
		return nil
	}
	commands := append([]storage.Command(nil), m.Commands[:m.pendingDelete]...)
	commands = append(commands, m.Commands[m.pendingDelete+1:]...)
	if err := storage.Save(commands); err != nil {
		m.Modal = ""
		m.pendingDelete = -1
		return m.failureCmd("Could not delete command: " + err.Error())
	}
	m.Commands = commands
	m.Modal = ""
	m.pendingDelete = -1
	if m.Selected >= len(m.treeRows()) {
		m.Selected = max(0, len(m.treeRows())-1)
	}
	m.syncDetails()
	return m.successCmd("[OK] Command deleted")
}

func (m *Model) successCmd(value string) tea.Cmd {
	return m.noticeCmd(value)
}

func (m *Model) failureCmd(value string) tea.Cmd {
	return m.noticeCmd(value)
}

func (m *Model) noticeCmd(value string) tea.Cmd {
	m.Status = value
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearStatusMsg(value) })
}

func (m *Model) recordUsage(index int) error {
	if index < 0 || index >= len(m.Commands) {
		return nil
	}
	m.Commands[index].Usage++
	if err := storage.Save(m.Commands); err != nil {
		m.Commands[index].Usage--
		return err
	}
	return nil
}

func (m *Model) recordCopy(index int) tea.Cmd {
	if err := m.recordUsage(index); err != nil {
		return m.failureCmd("Copied to clipboard, but usage count could not be saved")
	}
	return m.successCmd("[OK] Copied to clipboard!")
}

func (m *Model) layout() {
	innerWidth := max(1, m.Width-4)
	m.searchInput.Width = max(1, innerWidth-54)
	formWidth := max(1, min(64, m.Width-24))
	m.formInput.Width, m.variableInput.Width = formWidth, formWidth
	_, _, _, rightWidth := panelWidths(m.Width)
	m.helpView.Width = innerWidth
	helpHeight := strings.Count(helpFooter(m.helpView, innerWidth), "\n") + 1
	if m.Status != "" {
		helpHeight++
	}
	bodyHeight := max(3, m.Height-4-helpHeight)
	m.detailsView.Width, m.detailsView.Height = rightWidth, max(1, bodyHeight-3)
	m.historyView.Width, m.historyView.Height = formWidth, max(1, m.Height-12)
}
func (m *Model) syncDetails() {
	_, _, _, width := panelWidths(m.Width)
	m.detailsView.Width = width
	m.detailsView.GotoTop()
	m.detailsView.SetContent(wrapLines(m.detailsContent(), max(1, width)))
}

func (m *Model) syncHistory() {
	var b strings.Builder
	for i, line := range m.history {
		cursor := "  "
		if i == m.Selected {
			cursor = "▶ "
		}
		b.WriteString(cursor + truncate(line, max(1, m.historyView.Width-4)) + "\n")
	}
	m.historyView.SetContent(b.String())
}

func (m *Model) prepareFormInput() {
	m.formInput = textinput.New()
	m.formInput.Prompt = "> "
	m.formInput.SetValue(m.form[m.formField])
	m.formInput.Focus()
	m.formInput.Width = max(1, min(64, m.Width-24))
}

func (m Model) updateModal(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.Modal {
	case "actions":
		labels := m.actionLabels()
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "up":
			if m.actionSelected > 0 {
				m.actionSelected--
			}
		case "down":
			if m.actionSelected < len(labels)-1 {
				m.actionSelected++
			}
		case "enter":
			return m.runSelectedAction()
		case "c", "r", "f", "e", "d":
			for i, shortcut := range actionKeys {
				if key.String() == shortcut {
					m.actionSelected = i
					return m.runSelectedAction()
				}
			}
		}
	case "duplicate":
		switch key.String() {
		case "esc":
			m.Modal = "form"
		case "enter":
			return m, m.openDuplicate()
		}
	case "help", "about":
		if key.String() == "esc" || key.String() == "?" || key.String() == "enter" {
			m.Modal = ""
		}
	case "history":
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
				m.syncHistory()
			}
		case "down", "j":
			if m.Selected < len(m.history)-1 {
				m.Selected++
				m.syncHistory()
			}
		case "pgup", "pgdown":
			m.historyView, _ = m.historyView.Update(key)
		case "enter":
			if len(m.history) > 0 {
				m.pendingImport = m.history[m.Selected]
				m.Modal = "history_preview"
			}
		}
	case "history_preview":
		switch key.String() {
		case "esc":
			m.Modal = "history"
		case "enter":
			return m, m.importPendingHistory()
		}
	case "form":
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "tab", "enter":
			m.form[m.formField] = m.formInput.Value()
			delete(m.formErrors, m.formField)
			if m.formField < len(m.form)-1 {
				m.formField++
				m.prepareFormInput()
			} else {
				cmd := m.saveForm()
				return m, cmd
			}
		case "shift+tab":
			m.form[m.formField] = m.formInput.Value()
			delete(m.formErrors, m.formField)
			if m.formField > 0 {
				m.formField--
				m.prepareFormInput()
			}
		case "ctrl+s":
			m.form[m.formField] = m.formInput.Value()
			delete(m.formErrors, m.formField)
			cmd := m.saveForm()
			return m, cmd
		default:
			var cmd tea.Cmd
			m.formInput, cmd = m.formInput.Update(key)
			m.form[m.formField] = m.formInput.Value()
			delete(m.formErrors, m.formField)
			return m, cmd
		}
	case "variables":
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "tab", "enter", "ctrl+r":
			m.values[m.variables[m.varField]] = m.variableInput.Value()
			if m.varField < len(m.variables)-1 && key.String() != "ctrl+r" {
				m.varField++
				m.prepareVariableInput()
			} else {
				cmd := m.finishVariables()
				return m, cmd
			}
		default:
			var cmd tea.Cmd
			m.variableInput, cmd = m.variableInput.Update(key)
			m.values[m.variables[m.varField]] = m.variableInput.Value()
			return m, cmd
		}
	case "preview_run":
		switch key.String() {
		case "esc":
			m.Modal = ""
			return m, m.noticeCmd("Command run canceled")
		case "enter":
			return m, m.execute(m.pending)
		}
	case "confirm":
		switch key.String() {
		case "esc", "n":
			m.Modal = ""
			return m, m.noticeCmd("Command canceled")
		case "y", "enter":
			cmd := m.execute(m.pending)
			return m, cmd
		}
	case "confirm_delete":
		switch key.String() {
		case "esc", "n":
			m.Modal = ""
			m.pendingDelete = -1
		case "y", "enter":
			return m, m.deletePending()
		}
	}
	return m, nil
}

func (m *Model) prepareVariableInput() {
	m.variableInput = textinput.New()
	m.variableInput.Prompt = "> "
	if m.varField < len(m.variables) {
		m.variableInput.SetValue(m.values[m.variables[m.varField]])
	}
	m.variableInput.Focus()
	m.variableInput.Width = max(1, min(64, m.Width-24))
}

func (m *Model) saveForm() tea.Cmd {
	m.formErrors = m.validateForm()
	if len(m.formErrors) > 0 {
		return nil
	}
	command := storage.Command{Name: strings.TrimSpace(m.form[2]), Category: strings.TrimSpace(m.form[0]), Subcategory: strings.TrimSpace(m.form[1]), Description: strings.TrimSpace(m.form[3]), Command: strings.TrimSpace(m.form[5])}
	command.Tags = splitList(m.form[4])
	if len(m.form) > 8 {
		command.Shells = splitList(m.form[6])
		command.Platforms = splitList(m.form[7])
		command.Requirements = splitList(m.form[8])
	}
	commands := append([]storage.Command(nil), m.Commands...)
	for i, existing := range commands {
		if i != m.editIndex && strings.TrimSpace(existing.Command) == command.Command {
			m.duplicateIndex, m.Modal = i, "duplicate"
			return nil
		}
	}
	if m.editIndex >= 0 {
		command.Favorite = commands[m.editIndex].Favorite
		command.Usage = commands[m.editIndex].Usage
		commands[m.editIndex] = command
	} else {
		commands = append(commands, command)
	}
	if err := storage.Save(commands); err != nil {
		return m.failureCmd("Could not save command: " + err.Error())
	}
	m.Commands = commands
	m.formErrors = nil
	m.Modal = ""
	return m.successCmd("[OK] Command saved")
}

func splitList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func (m Model) validateForm() map[int]string {
	errors := make(map[int]string)
	if len(m.form) < 6 {
		errors[0] = "Form data is incomplete"
		return errors
	}
	category := strings.TrimSpace(m.form[0])
	validCategory := false
	for _, known := range storage.Categories() {
		if category == known {
			validCategory = true
			break
		}
	}
	if !validCategory {
		errors[0] = "Use a listed section"
	}
	if strings.TrimSpace(m.form[1]) == "" {
		errors[1] = "Subcategory is required"
	}
	if strings.TrimSpace(m.form[2]) == "" {
		errors[2] = "Name is required"
	}
	if strings.TrimSpace(m.form[3]) == "" {
		errors[3] = "Description is required"
	}
	command := strings.TrimSpace(m.form[5])
	if command == "" {
		errors[5] = "Command is required"
	} else if _, err := bridge.Parse(command); err != nil {
		errors[5] = "Command template is invalid: " + err.Error()
	}
	return errors
}

func (m *Model) openDuplicate() tea.Cmd {
	index := m.duplicateIndex
	m.Modal = ""
	m.duplicateIndex = -1
	m.Query = ""
	m.searchInput.SetValue("")
	m.Mode = "tree"
	if index >= 0 && index < len(m.Commands) {
		m.Tab = m.Commands[index].Category
		m.selectTreeCommand(index)
		m.syncDetails()
		return m.successCmd("Opened the existing command")
	}
	return nil
}

func (m *Model) importPendingHistory() tea.Cmd {
	line := strings.TrimSpace(m.pendingImport)
	if line == "" {
		m.Modal = "history"
		return nil
	}
	for _, command := range m.Commands {
		if strings.TrimSpace(command.Command) == line {
			m.Modal = "history"
			return m.noticeCmd("This command is already in your library")
		}
	}
	category := m.Tab
	if category == "All" {
		category = "Linux/Bash"
	}
	commands := append([]storage.Command(nil), m.Commands...)
	commands = append(commands, storage.Command{Name: "History - " + truncate(line, 32), Category: category, Subcategory: "Imported History", Description: "Imported from shell history", Tags: []string{"history"}, Command: line})
	if err := storage.Save(commands); err != nil {
		m.Modal = "history_preview"
		return m.failureCmd("Could not save command: " + err.Error())
	}
	m.Commands = commands
	m.pendingImport = ""
	m.Modal = ""
	return m.successCmd("[OK] Imported command from shell history")
}

func (m *Model) beginRun() tea.Cmd {
	m.variableAction = "run"
	_, command, ok := m.selectedCommand()
	if !ok {
		return nil
	}
	var info templateInfo
	parsed, err := bridge.Parse(command.Command)
	if err == nil {
		err = json.Unmarshal([]byte(parsed), &info)
	}
	if err != nil {
		return m.failureCmd("Could not parse command template: " + err.Error())
	}
	if len(info.Variables) > 0 {
		m.variableAction = "run"
		m.variables = make([]string, 0, len(info.Variables))
		m.values = make(map[string]string, len(info.Variables))
		for _, variable := range info.Variables {
			m.variables = append(m.variables, variable.Name)
			m.values[variable.Name] = variable.Default
		}
		m.varField = 0
		m.Modal = "variables"
		m.prepareVariableInput()
		return nil
	}
	return m.confirmOrRun(command.Command, info.Dangerous)
}

func (m *Model) beginCopy() tea.Cmd {
	m.variableAction = "copy"
	index, command, ok := m.selectedCommand()
	if !ok {
		return nil
	}
	var info templateInfo
	parsed, err := bridge.Parse(command.Command)
	if err == nil {
		err = json.Unmarshal([]byte(parsed), &info)
	}
	if err != nil {
		return m.failureCmd("Could not parse command template: " + err.Error())
	}
	if len(info.Variables) > 0 {
		m.variableAction = "copy"
		m.variables = make([]string, 0, len(info.Variables))
		m.values = make(map[string]string, len(info.Variables))
		for _, variable := range info.Variables {
			m.variables = append(m.variables, variable.Name)
			m.values[variable.Name] = variable.Default
		}
		m.varField, m.Modal = 0, "variables"
		m.prepareVariableInput()
		return nil
	}
	if err := clipboard.WriteAll(command.Command); err != nil {
		return m.failureCmd("Clipboard error: " + err.Error())
	}
	return m.recordCopy(index)
}

func (m *Model) finishVariables() tea.Cmd {
	index, command, ok := m.selectedCommand()
	if !ok {
		m.Modal = ""
		return nil
	}
	rendered, err := bridge.Render(command.Command, m.values)
	if err != nil {
		m.Modal = ""
		return m.failureCmd("Could not render command: " + err.Error())
	}
	if err = json.Unmarshal([]byte(rendered), &m.pending); err != nil {
		m.Modal = ""
		return m.failureCmd("Could not decode rendered command")
	}
	var info templateInfo
	parsed, err := bridge.Parse(m.pending)
	if err != nil {
		m.Modal = ""
		return m.failureCmd("Could not check command safety: " + err.Error())
	}
	if err = json.Unmarshal([]byte(parsed), &info); err != nil {
		m.Modal = ""
		return m.failureCmd("Could not check command safety: " + err.Error())
	}
	if m.variableAction == "copy" {
		m.variableAction = ""
		m.Modal = ""
		if err := clipboard.WriteAll(m.pending); err != nil {
			return m.failureCmd("Clipboard error: " + err.Error())
		}
		return m.recordCopy(index)
	}
	m.variableAction = ""
	return m.confirmOrRun(m.pending, info.Dangerous)
}

func (m *Model) confirmOrRun(command string, dangerous bool) tea.Cmd {
	m.pending = command
	if dangerous {
		m.Modal = "confirm"
	} else {
		m.Modal = "preview_run"
	}
	return nil
}

func activeShellName() string {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("pwsh"); err == nil {
			return "PowerShell 7 (pwsh)"
		}
		return "Windows PowerShell"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}
	return "POSIX shell (/bin/sh)"
}

func currentWorkingDirectory() string {
	if directory, err := os.Getwd(); err == nil {
		return directory
	}
	return "Unavailable"
}

func (m *Model) execute(command string) tea.Cmd {
	usageErr := ""
	if index, _, ok := m.selectedCommand(); ok {
		if err := m.recordUsage(index); err != nil {
			usageErr = err.Error()
		}
	}
	m.Modal = ""
	return tea.ExecProcess(newShellCommand(command), func(err error) tea.Msg { return runMsg{err: err, usageErr: usageErr} })
}

func newShellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		shell := "powershell"
		if _, err := exec.LookPath("pwsh"); err == nil {
			shell = "pwsh"
		}
		return exec.Command(shell, "-NoProfile", "-Command", command)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return exec.Command(shell, "-c", command)
}

func (b *cappedBuffer) Write(value []byte) (int, error) {
	const limit = 64 * 1024
	remaining := limit - b.Len()
	if remaining > 0 {
		_, _ = b.Buffer.Write(value[:min(remaining, len(value))])
	}
	if len(value) > remaining {
		b.truncated = true
	}
	return len(value), nil
}

func cleanOutput(value string) string {
	value = terminalEscape.ReplaceAllString(value, "")
	value = strings.Map(func(char rune) rune {
		if (char < 32 && char != '\n' && char != '\t') || char == 127 {
			return -1
		}
		return char
	}, value)
	return strings.Join(strings.Fields(value), " ")
}
