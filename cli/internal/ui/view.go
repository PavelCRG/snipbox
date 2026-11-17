package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/bridge"
)

func (m Model) View() string {
	hits := m.matches()
	var b strings.Builder
	b.WriteString(Title.Render("⚡ SNIPBOX") + "    ")
	for _, tab := range []string{"All", "Git", "Docker", "System", "Network", "Tools"} {
		if tab == m.Tab {
			b.WriteString(Selected.Render("[ " + tab + " ] "))
		} else {
			b.WriteString(CategoryStyle(tab).Render(" " + tab + " "))
		}
	}
	b.WriteString(Selected.Render(fmt.Sprintf("    %d / %d", len(hits), len(m.Commands))))
	b.WriteString("\n" + SectionTitle.Render("Search") + ": " + m.Query + "\n\n")
	var left, right strings.Builder
	rows := max(1, m.Height-15)
	start := max(0, m.Selected-rows+1)
	for i := start; i < min(start+rows, len(hits)); i++ {
		hit := hits[i]
		command := m.Commands[hit.Index]
		prefix, star := "  ", ""
		if i == m.Selected {
			prefix = "> "
		}
		if command.Favorite {
			star = "★ "
		}
		fmt.Fprintf(&left, "%s%s%s %s (%dx)\n", prefix, star, CategoryStyle(command.Category).Render("["+command.Category+"]"), highlight(command.Name, hit.Indices), command.Usage)
	}
	if start > 0 {
		left.WriteString("↑ more results above\n")
	}
	if start+rows < len(hits) {
		left.WriteString("↓ more results below\n")
	}
	if len(hits) > 0 {
		hit := hits[min(max(m.Selected, 0), len(hits)-1)]
		command := m.Commands[hit.Index]
		right.WriteString(SectionTitle.Render(highlight(command.Name, hit.Indices)) + "\n")
		right.WriteString(Muted.Render("Tags: ") + strings.Join(command.Tags, ", ") + "\n\n")
		right.WriteString(SectionTitle.Render("Description") + "\n" + command.Description + "\n\n")
		right.WriteString(SectionTitle.Render("Command") + "\n" + CommandBlock.Render(highlightTemplate(command.Command)) + "\n\n")
		right.WriteString(SectionTitle.Render("Variables") + "\n")
		parsed, err := bridge.Parse(command.Command)
		if err == nil {
			var info templateInfo
			if json.Unmarshal([]byte(parsed), &info) == nil {
				if len(info.Variables) == 0 {
					right.WriteString("None\n")
				}
				for _, variable := range info.Variables {
					fmt.Fprintf(&right, "• {{%s:%s}}\n", variable.Name, variable.Default)
				}
				if info.Dangerous {
					right.WriteString("\n" + Danger.Render("⚠ Dangerous command") + "\n")
				}
			}
		}
	}
	width := m.Width * 42 / 100
	if width < 24 {
		width = 36
	}
	rightWidth := max(24, m.Width-width-6)
	leftPanel := Panel.Width(width).Render(SectionTitle.Render("Commands") + "\n" + left.String())
	rightPanel := Panel.Width(rightWidth).Render(SectionTitle.Render("Command Details") + "\n" + right.String())
	panelHeight := max(lipgloss.Height(leftPanel), lipgloss.Height(rightPanel))
	gap := strings.TrimSuffix(strings.Repeat(" \n", panelHeight), "\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, gap, rightPanel))
	b.WriteString("\n\n" + Muted.Render("[Enter] Copy  [Ctrl+R] Run  [Tab] Tabs  [Ctrl+A] New  [Ctrl+E] Edit  [Ctrl+H] History  [?] Help  [Esc] Quit"))
	if m.Status != "" {
		style := Success
		if strings.Contains(strings.ToLower(m.Status), "error") || strings.Contains(strings.ToLower(m.Status), "failed") || strings.Contains(strings.ToLower(m.Status), "could not") {
			style = Danger
		}
		b.WriteString("\n" + style.Render(m.Status))
	}
	if m.Modal != "" {
		return Base.Render(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, Panel.BorderForeground(purple).Render(m.modalView())))
	}
	return Base.Render(b.String())
}

func (m Model) modalView() string {
	switch m.Modal {
	case "help":
		return "Keyboard Shortcuts\n\n↑ / ↓  Navigate commands\nType   Search\nEnter  Copy command\nCtrl+R Run command\nTab    Switch category\nCtrl+A Create command\nCtrl+E Edit command\nCtrl+H Import shell history\nEsc    Close / quit\n\nPress Esc to close"
	case "history":
		var b strings.Builder
		b.WriteString("Shell History — select a command\n\n")
		for i, line := range m.history {
			mark := "  "
			if i == m.Selected {
				mark = "> "
			}
			fmt.Fprintf(&b, "%s%s\n", mark, truncate(line, 76))
		}
		b.WriteString("\nEnter imports  ·  Esc closes")
		return b.String()
	case "form":
		labels := []string{"Category", "Name", "Description", "Tags (comma separated)", "Command"}
		var b strings.Builder
		if m.editIndex < 0 {
			b.WriteString("New Command\n\n")
		} else {
			b.WriteString("Edit Command\n\n")
		}
		for i, label := range labels {
			marker := "  "
			if i == m.formField {
				marker = "> "
			}
			fmt.Fprintf(&b, "%s%s: %s\n", marker, label, m.form[i])
		}
		b.WriteString("\nTab advances  ·  Ctrl+S saves  ·  Esc cancels")
		return b.String()
	case "variables":
		var b strings.Builder
		b.WriteString("Fill Variables\n\n")
		for i, name := range m.variables {
			mark := "  "
			if i == m.varField {
				mark = "> "
			}
			fmt.Fprintf(&b, "%s%s: %s\n", mark, name, m.values[name])
		}
		b.WriteString("\nEnter advances  ·  Esc cancels")
		return b.String()
	case "confirm":
		return "Confirm Execution\n\n⚠ This command can make destructive changes.\nReview it before continuing:\n\n" + m.pending + "\n\nPress Y to run  ·  N / Esc to cancel"
	default:
		return m.Modal
	}
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:max(0, width-1)]) + "…"
}

func highlight(value string, positions []int) string {
	matched := make(map[int]bool, len(positions))
	for _, position := range positions {
		matched[position] = true
	}
	var b strings.Builder
	for i, r := range []rune(value) {
		if matched[i] {
			b.WriteString(Highlight.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func highlightTemplate(value string) string {
	var b strings.Builder
	for len(value) > 0 {
		start := strings.Index(value, "{{")
		if start < 0 {
			b.WriteString(value)
			break
		}
		b.WriteString(value[:start])
		value = value[start:]
		end := strings.Index(value, "}}")
		if end < 0 {
			b.WriteString(value)
			break
		}
		end += 2
		b.WriteString(Variable.Render(value[:end]))
		value = value[end:]
	}
	return b.String()
}
