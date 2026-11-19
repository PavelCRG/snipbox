package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/snipbox/cli/internal/bridge"
	"github.com/snipbox/cli/internal/storage"
)

func (m Model) View() string {
	if m.Width < 70 || m.Height < 18 {
		width := max(1, m.Width)
		hint := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(wrapLines("Terminal window too small (min 70x18). Please resize.", width))
		return fit(Base, width, m.Height, hint)
	}
	if m.Modal != "" {
		return m.renderModal()
	}
	return m.browseView()
}

func (m Model) browseView() string {
	header := m.headerView()
	helpView := m.helpView
	helpView.Width = max(1, m.Width-4)
	helpLine := Muted.Render(helpFooter(helpView, max(1, m.Width-4)))
	status := ""
	if m.Status != "" {
		style := Success
		if strings.Contains(strings.ToLower(m.Status), "error") || strings.Contains(strings.ToLower(m.Status), "failed") || strings.Contains(strings.ToLower(m.Status), "could not") {
			style = Danger
		}
		status = style.Render(ansi.Truncate(m.Status, max(1, m.Width-2), "…"))
	}
	headerHeight, footerHeight := lipgloss.Height(header), lipgloss.Height(helpLine)
	if status != "" {
		footerHeight += lipgloss.Height(status)
	}
	bodyHeight := max(3, m.Height-headerHeight-footerHeight)
	_, _, leftContent, rightContent := panelWidths(m.Width)
	contentHeight := max(1, bodyHeight-2)
	left := m.commandList(leftContent, contentHeight)
	right := m.detailsPanel(rightContent, contentHeight)
	split := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", 2), right)
	view := lipgloss.JoinVertical(lipgloss.Left, header, split, helpLine)
	if status != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, view, status)
	}
	return fit(Base, m.Width, m.Height, view)
}

func commandCountLabel(count int) string {
	if count == 1 {
		return "1 command"
	}
	return fmt.Sprintf("%d commands", count)
}

func helpFooter(model help.Model, width int) string {
	var rows []string
	var current []key.Binding
	for _, binding := range keys.ShortHelp() {
		candidate := append(append([]key.Binding(nil), current...), binding)
		model.Width = 0
		if len(current) > 0 && lipgloss.Width(model.ShortHelpView(candidate)) > width {
			model.Width = width
			rows = append(rows, model.ShortHelpView(current))
			current = []key.Binding{binding}
		} else {
			current = candidate
		}
	}
	if len(current) > 0 {
		model.Width = width
		rows = append(rows, model.ShortHelpView(current))
	}
	return strings.Join(rows, "\n")
}

func (m Model) headerView() string {
	innerWidth := max(1, m.Width-2)
	contentWidth := max(1, innerWidth-2)
	brand := Title.Render("SNIPBOX")
	tabs := append([]string{"All"}, storage.Categories()...)
	var tabParts []string
	for _, tab := range tabs {
		style := Muted
		if tab == m.Tab {
			style = CategoryStyle(tab)
		}
		tabParts = append(tabParts, style.Render(tab))
	}
	tabRow := strings.Join(tabParts, " ")
	top := lipgloss.JoinHorizontal(lipgloss.Top, brand, strings.Repeat(" ", max(1, contentWidth-lipgloss.Width(brand)-lipgloss.Width(tabRow))), tabRow)
	top = ansi.Truncate(top, contentWidth, "…")
	search := m.searchInput
	search.Prompt = "> Search: "
	count := Selected.Render(fmt.Sprintf("%d / %d", len(m.matches()), len(m.Commands)))
	search.Width = max(1, contentWidth-lipgloss.Width(search.Prompt)-lipgloss.Width(count)-2)
	searchView := search.View()
	second := lipgloss.JoinHorizontal(lipgloss.Top, searchView, strings.Repeat(" ", max(1, contentWidth-lipgloss.Width(searchView)-lipgloss.Width(count))), count)
	content := lipgloss.JoinVertical(lipgloss.Left, ansi.Truncate(top, contentWidth, "…"), ansi.Truncate(second, contentWidth, "…"))
	return Header.Width(innerWidth).Render(content)
}

func panelWidths(width int) (leftOuter, rightOuter, leftContent, rightContent int) {
	leftOuter = max(1, int(float64(width)*0.42))
	rightOuter = max(1, width-leftOuter-2)
	leftContent = max(1, leftOuter-4)
	rightContent = max(1, rightOuter-4)
	return
}

func (m Model) commandList(width, height int) string {
	var b strings.Builder
	rows := m.matches()
	visibleRows := max(1, height-1) // The panel title occupies one row.
	if m.Mode == "tree" {
		tree := m.treeRows()
		start := min(max(0, m.Selected-visibleRows+1), max(0, len(tree)-visibleRows))
		for i := start; i < min(start+visibleRows, len(tree)); i++ {
			row := tree[i]
			marker := "  "
			if i == m.Selected {
				marker = Selected.Render("┃ ")
			}
			label := ""
			switch row.Kind {
			case "category":
				arrow := "▸"
				if m.categoryExpanded[row.Category] || m.Query != "" {
					arrow = "▾"
				}
				label = arrow + " " + row.Category + fmt.Sprintf(" (%d)", row.MatchCount)
			case "subcategory":
				key := row.Category + "\x00" + row.Subcategory
				arrow := "▸"
				if m.subcategoryExpanded[key] || m.Query != "" {
					arrow = "▾"
				}
				label = "  " + arrow + " " + row.Subcategory
			case "command":
				c := m.Commands[row.Index]
				star := " "
				if c.Favorite {
					star = "★"
				}
				label = "      " + star + " " + c.Name
				usage := Muted.Render(fmt.Sprintf("%dx", c.Usage))
				labelWidth := max(1, width-lipgloss.Width(marker)-lipgloss.Width(usage)-1)
				label = ansi.Truncate(label, labelWidth, "…")
				padding := strings.Repeat(" ", max(1, width-lipgloss.Width(marker)-lipgloss.Width(label)-lipgloss.Width(usage)))
				style := Muted
				if i == m.Selected {
					style = Selected
				}
				b.WriteString(ansi.Truncate(marker+style.Render(label)+padding+usage, width, "…") + "\n")
				continue
			}
			style := Muted
			if i == m.Selected {
				style = Selected
			}
			b.WriteString(ansi.Truncate(marker+style.Render(label), width, "…") + "\n")
		}
	} else {
		start := min(max(0, m.Selected-visibleRows+1), max(0, len(rows)-visibleRows))
		if len(rows) == 0 {
			b.WriteString(Muted.Render("No commands match this search."))
		}
		for i := start; i < min(start+visibleRows, len(rows)); i++ {
			hit := rows[i]
			c := m.Commands[hit.Index]
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(foreground)
			if i == m.Selected {
				cursor = Selected.Render("┃ ▶")
				nameStyle = Selected
			}
			favorite := "  "
			if c.Favorite {
				favorite = "★ "
			}
			badge := CategoryStyle(c.Category).Render(c.Category)
			usage := Muted.Render(fmt.Sprintf("%dx", c.Usage))
			nameWidth := max(1, width-lipgloss.Width(cursor)-lipgloss.Width(favorite)-lipgloss.Width(badge)-lipgloss.Width(usage)-4)
			name := ansi.Truncate(c.Name, nameWidth, "…")
			row := lipgloss.JoinHorizontal(lipgloss.Left, cursor, favorite, badge, " ", nameStyle.Render(highlight(name, hit.Indices)), " ", usage)
			b.WriteString(ansi.Truncate(row, width, "…") + "\n")
		}
	}
	return ActivePanel.Width(width + 2).Height(height).Render(SectionTitle.Render("Commands") + "\n" + strings.TrimSuffix(b.String(), "\n"))
}

func (m Model) detailsPanel(width, height int) string {
	details := m.detailsContent()
	if m.Mode == "tree" {
		if row, ok := m.selectedTreeRow(); ok && row.Kind != "command" {
			switch row.Kind {
			case "category":
				details = "Select a subcategory to browse its commands."
			case "subcategory":
				details = fmt.Sprintf("%s in this subcategory. Expand to browse.", commandCountLabel(row.MatchCount))
			}
		}
	}
	viewport := m.detailsView
	viewport.Width = max(1, width)
	viewport.Height = max(1, height-1)
	viewport.SetContent(wrapLines(details, max(1, width)))
	return DetailPanel.Width(width + 2).Height(height).Render(SectionTitle.Render("Command Details") + "\n" + viewport.View())
}

func (m Model) detailsContent() string {
	_, command, ok := m.selectedCommand()
	if !ok {
		return "Select a command to see its details."
	}
	var b strings.Builder
	b.WriteString(Title.Render(highlight(command.Name, m.selectedIndices())) + "\n")
	var tags []string
	for _, tag := range command.Tags {
		tags = append(tags, "#"+tag)
	}
	b.WriteString(Muted.Render(strings.Join(tags, "  ")) + "\n\n")
	b.WriteString(SectionTitle.Render("Description") + "\n" + command.Description + "\n\n")
	b.WriteString(SectionTitle.Render("Command") + "\n" + CommandBlock.Render(highlightTemplate(command.Command)) + "\n\n")
	b.WriteString(SectionTitle.Render("Parameters") + "\n")
	parsed, err := bridge.Parse(command.Command)
	if err != nil {
		b.WriteString("None\n")
		return b.String()
	}
	var info templateInfo
	if json.Unmarshal([]byte(parsed), &info) != nil {
		b.WriteString("None\n")
		return b.String()
	}
	if len(info.Variables) == 0 {
		b.WriteString(Muted.Render("None") + "\n")
	} else {
		for _, v := range info.Variables {
			fmt.Fprintf(&b, "• {{%s:%s}}\n", v.Name, v.Default)
		}
	}
	if info.Dangerous {
		b.WriteString("\n" + Danger.Render("[!] DANGEROUS COMMAND"))
	}
	return b.String()
}

func (m Model) selectedIndices() []int {
	if m.Mode == "tree" {
		if row, ok := m.selectedTreeRow(); ok && row.Kind == "command" {
			return row.Indices
		}
		return nil
	}
	rows := m.matches()
	if m.Selected < 0 || m.Selected >= len(rows) {
		return nil
	}
	return rows[m.Selected].Indices
}

func (m Model) renderModal() string {
	modalWidth := max(1, min(68, m.Width-4))
	content := wrapLines(m.modalView(), max(1, modalWidth-10))
	lines := strings.Split(content, "\n")
	maxHeight := max(1, m.Height-6)
	if len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}
	content = strings.Join(lines, "\n")
	outer := modalBox(modalWidth, content)
	return fit(Base, m.Width, m.Height, lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, outer))
}

func modalBox(width int, content string) string {
	width = max(1, width)
	innerWidth := max(1, width-10)
	inner := ModalInner.Width(max(1, width-6)).Render(wrapLines(content, innerWidth))
	return ModalOuter.Width(max(1, width-2)).Render(inner)
}

func (m Model) modalView() string {
	switch m.Modal {
	case "actions":
		var b strings.Builder
		b.WriteString("Command Actions\n\n")
		for i, label := range m.actionLabels() {
			if i == m.actionSelected {
				fmt.Fprintf(&b, "%s %s\n", Selected.Render("▶"), Selected.Render(label))
			} else {
				fmt.Fprintf(&b, "  %s\n", label)
			}
		}
		b.WriteString("\nUp/Down move  •  Enter choose  •  Esc close")
		return b.String()
	case "help":
		return m.helpView.FullHelpView(keys.FullHelp())
	case "about":
		return strings.Join([]string{
			"About Snipbox",
			"",
			"Search commands by name, description, tags, or text.",
			"The Rust engine ranks matches; Go powers the terminal interface.",
			"Commands are saved in local TOML. Shell history is imported only on request.",
			"Created by Pavel Slepcov.",
			"",
			"Press Esc to close.",
		}, "\n")
	case "history":
		return "Shell History\n\n" + m.historyView.View() + "\nEnter imports  •  Esc closes"
	case "form":
		labels := []string{"Category", "Subcategory", "Name", "Description", "Tags (comma separated)", "Command"}
		var b strings.Builder
		if m.editIndex < 0 {
			b.WriteString("New Command\n\n")
		} else {
			b.WriteString("Edit Command\n\n")
		}
		for i, l := range labels {
			if i == m.formField {
				fmt.Fprintf(&b, "┃ %s: %s\n", l, m.formInput.View())
			} else if i < len(m.form) {
				fmt.Fprintf(&b, "  %s: %s\n", l, m.form[i])
			}
		}
		b.WriteString("\nTab moves forward  •  Ctrl+S saves  •  Esc cancels")
		return b.String()
	case "variables":
		var b strings.Builder
		b.WriteString("Fill Variables\n\n")
		for i, n := range m.variables {
			v := m.values[n]
			if i == m.varField {
				v = m.variableInput.View()
				fmt.Fprintf(&b, "┃ {{%s}}: %s\n", n, v)
			} else {
				fmt.Fprintf(&b, "  {{%s}}: %s\n", n, v)
			}
		}
		b.WriteString("\nEnter moves forward  •  Ctrl+R runs  •  Esc cancels")
		return b.String()
	case "confirm":
		return "Confirm Execution\n\n[!] Review this potentially destructive command:\n\n" + m.pending + "\n\nPress Y to run  •  N / Esc to cancel"
	case "confirm_delete":
		return "Delete Command?\n\nThis permanently removes the selected command from your local library.\n\nPress Y or Enter to delete  •  N / Esc to cancel"
	}
	return m.Modal
}

func fit(style lipgloss.Style, width, height int, value string) string {
	width = max(1, width)
	height = max(1, height)
	lines := strings.Split(value, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "…")
	}
	return style.Width(width).MaxWidth(width).Height(height).MaxHeight(height).Render(strings.Join(lines, "\n"))
}
func truncate(value string, width int) string { return ansi.Truncate(value, max(0, width), "…") }
func wrapLines(value string, width int) string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = ansi.Wrap(lines[i], max(1, width), " ")
	}
	return strings.Join(lines, "\n")
}
func highlight(value string, positions []int) string {
	matched := make(map[int]bool, len(positions))
	for _, p := range positions {
		matched[p] = true
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
