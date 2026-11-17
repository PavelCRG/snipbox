package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
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
	output string
	err    error
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
	Status   string
	Modal    string
	Language language
	Width    int
	Height   int

	form      []string
	formField int
	editIndex int
	history   []string
	variables []string
	values    map[string]string
	varField  int
	pending   string
}

func New(commands []storage.Command, query string) Model {
	return Model{Commands: commands, Query: query, Tab: "Git", Language: english, editIndex: -1}
}

func (m Model) Init() tea.Cmd { return nil }

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
	hits := m.matches()
	if len(hits) == 0 {
		return 0, storage.Command{}, false
	}
	index := min(max(m.Selected, 0), len(hits)-1)
	commandIndex := hits[index].Index
	return commandIndex, m.Commands[commandIndex], true
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = v.Width, v.Height
	case runMsg:
		if v.err != nil {
			m.Status = tr(m.Language, "failed") + v.err.Error() + " " + truncate(strings.TrimSpace(v.output), 100)
		} else {
			m.Status = tr(m.Language, "completed") + ": " + truncate(strings.TrimSpace(v.output), 100)
		}
		return m, nil
	case tea.KeyMsg:
		if m.Modal != "" {
			return m.updateModal(v)
		}
		switch v.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			}
		case "down", "j":
			if m.Selected < len(m.matches())-1 {
				m.Selected++
			}
		case "tab":
			tabs := []string{"All", "Git", "Docker", "System", "Network", "Tools"}
			for i, tab := range tabs {
				if tab == m.Tab {
					m.Tab = tabs[(i+1)%len(tabs)]
					break
				}
			}
			m.Selected = 0
		case "?":
			m.Modal = "help"
		case "ctrl+l":
			if m.Language == english {
				m.Language = russian
			} else {
				m.Language = english
			}
		case "ctrl+o":
			m.Modal = "about"
		case "enter":
			_, command, ok := m.selectedCommand()
			if ok {
				if err := clipboard.WriteAll(command.Command); err != nil {
					m.Status = tr(m.Language, "clipboard_error") + err.Error()
				} else {
					m.Status = tr(m.Language, "copied")
				}
			}
		case "ctrl+r":
			cmd := m.beginRun()
			return m, cmd
		case "ctrl+a":
			m.form = []string{"Tools", "", "", "", ""}
			m.formField = 0
			m.editIndex = -1
			m.Modal = "form"
		case "ctrl+e":
			if index, command, ok := m.selectedCommand(); ok {
				m.form = []string{command.Category, command.Name, command.Description, strings.Join(command.Tags, ", "), command.Command}
				m.formField = 0
				m.editIndex = index
				m.Modal = "form"
			}
		case "ctrl+h":
			m.history = history.Recent()
			if len(m.history) == 0 {
				m.Status = tr(m.Language, "no_history")
			} else {
				m.Selected = 0
				m.Modal = "history"
			}
		case "backspace":
			if len(m.Query) > 0 {
				m.Query = dropLastRune(m.Query)
				m.Selected = 0
			}
		default:
			if len(v.Runes) > 0 {
				m.Query += string(v.Runes)
				m.Selected = 0
			}
		}
	}
	return m, nil
}

func (m Model) updateModal(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.Modal {
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
			}
		case "down", "j":
			if m.Selected < len(m.history)-1 {
				m.Selected++
			}
		case "enter":
			if len(m.history) > 0 {
				line := m.history[m.Selected]
				m.Commands = append(m.Commands, storage.Command{Name: "History · " + truncate(line, 32), Category: "Tools", Description: "Imported from shell history", Tags: []string{"history"}, Command: line})
				if err := storage.Save(m.Commands); err != nil {
					m.Status = tr(m.Language, "save_error") + err.Error()
				} else {
					m.Status = tr(m.Language, "imported")
				}
				m.Modal = ""
			}
		}
	case "form":
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "tab", "enter":
			if m.formField < len(m.form)-1 {
				m.formField++
			} else {
				m.saveForm()
			}
		case "shift+tab":
			if m.formField > 0 {
				m.formField--
			}
		case "ctrl+s":
			m.saveForm()
		case "backspace":
			if len(m.form[m.formField]) > 0 {
				m.form[m.formField] = dropLastRune(m.form[m.formField])
			}
		default:
			if len(key.Runes) > 0 {
				m.form[m.formField] += string(key.Runes)
			}
		}
	case "variables":
		switch key.String() {
		case "esc":
			m.Modal = ""
		case "tab", "enter":
			if m.varField < len(m.variables)-1 {
				m.varField++
			} else {
				cmd := m.finishVariables()
				return m, cmd
			}
		case "backspace":
			name := m.variables[m.varField]
			if len(m.values[name]) > 0 {
				m.values[name] = dropLastRune(m.values[name])
			}
		default:
			if len(key.Runes) > 0 {
				name := m.variables[m.varField]
				m.values[name] += string(key.Runes)
			}
		}
	case "confirm":
		switch key.String() {
		case "esc", "n":
			m.Modal = ""
			m.Status = tr(m.Language, "canceled")
		case "y", "enter":
			cmd := m.execute(m.pending)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) saveForm() {
	if strings.TrimSpace(m.form[1]) == "" || strings.TrimSpace(m.form[4]) == "" {
		m.Status = tr(m.Language, "required")
		return
	}
	command := storage.Command{Name: strings.TrimSpace(m.form[1]), Category: strings.TrimSpace(m.form[0]), Description: strings.TrimSpace(m.form[2]), Command: strings.TrimSpace(m.form[4])}
	for _, tag := range strings.Split(m.form[3], ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			command.Tags = append(command.Tags, tag)
		}
	}
	commands := append([]storage.Command(nil), m.Commands...)
	if m.editIndex >= 0 {
		command.Favorite = commands[m.editIndex].Favorite
		command.Usage = commands[m.editIndex].Usage
		commands[m.editIndex] = command
	} else {
		commands = append(commands, command)
	}
	if err := storage.Save(commands); err != nil {
		m.Status = tr(m.Language, "save_error") + err.Error()
		return
	}
	m.Commands = commands
	m.Modal = ""
	m.Status = tr(m.Language, "saved")
}

func (m *Model) beginRun() tea.Cmd {
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
		m.Status = tr(m.Language, "parse_error") + err.Error()
		return nil
	}
	if len(info.Variables) > 0 {
		m.variables = make([]string, 0, len(info.Variables))
		m.values = make(map[string]string, len(info.Variables))
		for _, variable := range info.Variables {
			m.variables = append(m.variables, variable.Name)
			m.values[variable.Name] = variable.Default
		}
		m.varField = 0
		m.Modal = "variables"
		return nil
	}
	return m.confirmOrRun(command.Command, info.Dangerous)
}

func (m *Model) finishVariables() tea.Cmd {
	_, command, ok := m.selectedCommand()
	if !ok {
		m.Modal = ""
		return nil
	}
	rendered, err := bridge.Render(command.Command, m.values)
	if err != nil {
		m.Status = tr(m.Language, "render_error") + err.Error()
		m.Modal = ""
		return nil
	}
	if err = json.Unmarshal([]byte(rendered), &m.pending); err != nil {
		m.Status = tr(m.Language, "decode_error")
		m.Modal = ""
		return nil
	}
	var info templateInfo
	parsed, err := bridge.Parse(m.pending)
	if err != nil {
		m.Status = tr(m.Language, "safety_error") + err.Error()
		m.Modal = ""
		return nil
	}
	if err = json.Unmarshal([]byte(parsed), &info); err != nil {
		m.Status = tr(m.Language, "safety_error") + err.Error()
		m.Modal = ""
		return nil
	}
	return m.confirmOrRun(m.pending, info.Dangerous)
}

func (m *Model) confirmOrRun(command string, dangerous bool) tea.Cmd {
	m.pending = command
	if dangerous {
		m.Modal = "confirm"
		return nil
	}
	return m.execute(command)
}

func (m *Model) execute(command string) tea.Cmd {
	if index, _, ok := m.selectedCommand(); ok {
		m.Commands[index].Usage++
		if err := storage.Save(m.Commands); err != nil {
			m.Status = tr(m.Language, "usage_error") + err.Error()
		}
	}
	m.Modal = ""
	return func() tea.Msg {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			shell := "powershell"
			if _, err := exec.LookPath("pwsh"); err == nil {
				shell = "pwsh"
			}
			cmd = exec.Command(shell, "-NoProfile", "-NonInteractive", "-Command", command)
		} else {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}
			cmd = exec.Command(shell, "-c", command)
		}
		var output cappedBuffer
		cmd.Stdout, cmd.Stderr = &output, &output
		err := cmd.Run()
		text := cleanOutput(output.String())
		if output.truncated {
			text += " [output truncated]"
		}
		return runMsg{output: text, err: err}
	}
}

func dropLastRune(value string) string {
	_, size := utf8.DecodeLastRuneInString(value)
	return value[:len(value)-size]
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
