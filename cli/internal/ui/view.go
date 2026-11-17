package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/snipbox/cli/internal/bridge"
)

func (m Model) View() string {
	if m.Width < 72 || m.Height < 16 {
		return Base.Render("SNIPBOX\n\n" + tr(m.Language, "resize"))
	}
	hits := m.matches()
	var b strings.Builder
	b.WriteString(Title.Render("SNIPBOX") + "    ")
	for _, tab := range []string{"All", "Git", "Docker", "System", "Network", "Tools"} {
		label := tab
		if tab == "All" && m.Language == russian {
			label = "Все"
		}
		if tab == m.Tab {
			b.WriteString(Selected.Render("[ " + label + " ] "))
		} else {
			b.WriteString(CategoryStyle(tab).Render(" " + label + " "))
		}
	}
	b.WriteString(Selected.Render(fmt.Sprintf("    %d / %d", len(hits), len(m.Commands))))
	b.WriteString("\n" + SectionTitle.Render(tr(m.Language, "search")) + ": " + m.Query + "\n\n")
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
		right.WriteString(Muted.Render(tr(m.Language, "tags")+": ") + strings.Join(command.Tags, ", ") + "\n\n")
		right.WriteString(SectionTitle.Render(tr(m.Language, "description")) + "\n" + command.Description + "\n\n")
		right.WriteString(SectionTitle.Render(tr(m.Language, "command")) + "\n" + CommandBlock.Render(highlightTemplate(command.Command)) + "\n\n")
		right.WriteString(SectionTitle.Render(tr(m.Language, "variables")) + "\n")
		parsed, err := bridge.Parse(command.Command)
		if err == nil {
			var info templateInfo
			if json.Unmarshal([]byte(parsed), &info) == nil {
				if len(info.Variables) == 0 {
					right.WriteString(tr(m.Language, "none") + "\n")
				}
				for _, variable := range info.Variables {
					fmt.Fprintf(&right, "• {{%s:%s}}\n", variable.Name, variable.Default)
				}
				if info.Dangerous {
					right.WriteString("\n" + Danger.Render("⚠ "+tr(m.Language, "danger")) + "\n")
				}
			}
		}
	}
	// Width includes each panel's border and padding; keep the complete row inside the terminal.
	leftWidth := (m.Width - 2) * 42 / 100
	rightWidth := m.Width - leftWidth - 2
	contentHeight := max(1, m.Height-9)
	if len(hits) == 0 {
		left.WriteString(tr(m.Language, "no_results") + "\n")
	}
	leftPanel := Panel.Width(max(1, leftWidth-4)).Height(contentHeight).MaxHeight(contentHeight).Render(SectionTitle.Render(tr(m.Language, "commands")) + "\n" + left.String())
	rightPanel := Panel.Width(max(1, rightWidth-4)).Height(contentHeight).MaxHeight(contentHeight).Render(SectionTitle.Render(tr(m.Language, "details")) + "\n" + right.String())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel))
	b.WriteString("\n\n" + Muted.Render(tr(m.Language, "footer")))
	if m.Status != "" {
		style := Success
		status := strings.ToLower(m.Status)
		if strings.Contains(status, "error") || strings.Contains(status, "failed") || strings.Contains(status, "could not") || strings.Contains(status, "ошибка") || strings.Contains(status, "не удалось") || strings.Contains(status, "ошибкой") {
			style = Danger
		}
		b.WriteString("\n" + style.Render(m.Status))
	}
	if m.Modal != "" {
		return Base.Render(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, Panel.BorderForeground(purple).Render(m.modalView())))
	}
	return Base.Width(m.Width).MaxWidth(m.Width).Height(m.Height).MaxHeight(m.Height).Render(b.String())
}

func (m Model) modalView() string {
	switch m.Modal {
	case "help":
		return strings.Join([]string{tr(m.Language, "help_title"), "", "Up / Down    " + tr(m.Language, "navigate"), tr(m.Language, "type_search"), "Enter        " + tr(m.Language, "copy"), "Ctrl+R       " + tr(m.Language, "run"), "Tab          " + tr(m.Language, "tabs"), "Ctrl+A       " + tr(m.Language, "new"), "Ctrl+E       " + tr(m.Language, "edit"), "Ctrl+H       " + tr(m.Language, "history_key"), "Ctrl+L       " + tr(m.Language, "language_key"), "Ctrl+O       " + tr(m.Language, "about_key"), "Esc          " + tr(m.Language, "close"), "", tr(m.Language, "press_close")}, "\n")
	case "about":
		return strings.Join([]string{tr(m.Language, "about_title"), "", tr(m.Language, "about_line1"), tr(m.Language, "about_line2"), "", tr(m.Language, "author"), tr(m.Language, "stack"), tr(m.Language, "version"), "", "Ctrl+L  " + tr(m.Language, "language_name"), tr(m.Language, "press_close")}, "\n")
	case "history":
		var b strings.Builder
		b.WriteString(tr(m.Language, "history_title") + "\n\n")
		for i, line := range m.history {
			mark := "  "
			if i == m.Selected {
				mark = "> "
			}
			fmt.Fprintf(&b, "%s%s\n", mark, truncate(line, 76))
		}
		b.WriteString("\n" + tr(m.Language, "history_footer"))
		return b.String()
	case "form":
		labels := []string{tr(m.Language, "category"), tr(m.Language, "name"), tr(m.Language, "description"), tr(m.Language, "tags_form"), tr(m.Language, "command")}
		var b strings.Builder
		if m.editIndex < 0 {
			b.WriteString(tr(m.Language, "new_title") + "\n\n")
		} else {
			b.WriteString(tr(m.Language, "edit_title") + "\n\n")
		}
		for i, label := range labels {
			marker := "  "
			if i == m.formField {
				marker = "> "
			}
			fmt.Fprintf(&b, "%s%s: %s\n", marker, label, m.form[i])
		}
		b.WriteString("\n" + tr(m.Language, "tab_advances"))
		return b.String()
	case "variables":
		var b strings.Builder
		b.WriteString(tr(m.Language, "fill_variables") + "\n\n")
		for i, name := range m.variables {
			mark := "  "
			if i == m.varField {
				mark = "> "
			}
			fmt.Fprintf(&b, "%s%s: %s\n", mark, name, m.values[name])
		}
		b.WriteString("\n" + tr(m.Language, "enter_advances"))
		return b.String()
	case "confirm":
		return tr(m.Language, "confirm_title") + "\n\n⚠ " + tr(m.Language, "danger_explainer") + "\n" + tr(m.Language, "review_command") + "\n\n" + m.pending + "\n\n" + tr(m.Language, "confirm_keys")
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
