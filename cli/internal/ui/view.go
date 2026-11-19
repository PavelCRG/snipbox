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
	total := len(m.Commands)
	if m.Tab != "All" {
		total = 0
		for _, command := range m.Commands {
			if command.Category == m.Tab {
				total++
			}
		}
	}
	count := Selected.Render(fmt.Sprintf("%d / %d · %s", len(m.matches()), total, m.Tab))
	search.Width = max(1, contentWidth-lipgloss.Width(search.Prompt)-lipgloss.Width(count)-2)
	searchView := search.View()
	second := lipgloss.JoinHorizontal(lipgloss.Top, searchView, strings.Repeat(" ", max(1, contentWidth-lipgloss.Width(searchView)-lipgloss.Width(count))), count)
	purpose := Muted.Render("A personal library for reusable terminal commands")
	content := lipgloss.JoinVertical(lipgloss.Left, ansi.Truncate(top, contentWidth, "…"), ansi.Truncate(second, contentWidth, "…"), ansi.Truncate(purpose, contentWidth, "…"))
	return Header.Width(innerWidth).Render(content)
}

func panelWidths(width int) (leftOuter, rightOuter, leftContent, rightContent int) {
	leftOuter = max(1, int(float64(width)*0.42))
	rightOuter = max(1, width-leftOuter-2)
	leftContent = max(1, leftOuter-4)
	rightContent = max(1, rightOuter-4)
	return
}

func listStart(selected, visible, total int) int {
	return min(max(0, selected-visible+1), max(0, total-visible))
}

func treeBranch(rows []treeRow, index int) string {
	row := rows[index]
	if row.Kind == "subcategory" {
		for i := index + 1; i < len(rows); i++ {
			if rows[i].Kind == "category" || (rows[i].Kind == "subcategory" && rows[i].Category == row.Category) {
				if rows[i].Kind == "subcategory" {
					return "├─"
				}
				break
			}
		}
		return "└─"
	}
	if row.Kind != "command" {
		return ""
	}
	lastCommand := true
	laterSubcategory := false
	for i := index + 1; i < len(rows); i++ {
		next := rows[i]
		if next.Kind == "category" || (next.Kind == "subcategory" && next.Category != row.Category) {
			break
		}
		if next.Kind == "subcategory" {
			laterSubcategory = next.Category == row.Category
			break
		}
		if next.Kind == "command" && next.Category == row.Category && next.Subcategory == row.Subcategory {
			lastCommand = false
			break
		}
	}
	spine := "   "
	if laterSubcategory {
		spine = "│  "
	}
	branch := "├─"
	if lastCommand {
		branch = "└─"
	}
	return spine + branch
}

func (m Model) commandList(width, height int) string {
	var b strings.Builder
	rows := m.matches()
	visibleRows := max(1, height-1) // The panel title occupies one row.
	if m.Mode == "tree" {
		tree := m.treeRows()
		if len(tree) == 0 {
			message := "No commands in your library."
			if m.Query != "" {
				message = "No commands match your search."
			} else if len(m.Commands) > 0 {
				message = "No commands in this category yet."
			}
			b.WriteString(Muted.Render(message))
		}
		start := listStart(m.Selected, visibleRows, len(tree))
		for i := start; i < min(start+visibleRows, len(tree)); i++ {
			row := tree[i]
			selected := i == m.Selected
			marker := "  "
			if selected {
				marker = "\u25b6 "
			}
			line := ""
			switch row.Kind {
			case "category":
				arrow := "\u25b8"
				if m.categoryExpanded[row.Category] || m.Query != "" {
					arrow = "\u25be"
				}
				if selected {
					line = SelectedRow.Width(width).Render(marker + arrow + " " + row.Category + fmt.Sprintf(" (%d)", row.MatchCount))
				} else {
					line = marker + arrow + " " + CategoryStyle(row.Category).Render(row.Category) + Muted.Render(fmt.Sprintf(" (%d)", row.MatchCount))
				}
			case "subcategory":
				path := row.Category + "\x00" + row.Subcategory
				arrow := "\u25b8"
				if m.subcategoryExpanded[path] || m.Query != "" {
					arrow = "\u25be"
				}
				label := "  " + treeBranch(tree, i) + " " + arrow + " " + row.Subcategory
				if selected {
					line = SelectedRow.Width(width).Render(marker + label)
				} else {
					line = marker + CategoryTextStyle(row.Category).Render(label)
				}
			case "command":
				command := m.Commands[row.Index]
				star := " "
				if command.Favorite {
					star = "\u2605"
				}
				label := "  " + treeBranch(tree, i) + " " + star + " " + command.Name
				usage := fmt.Sprintf("%dx", command.Usage)
				labelWidth := max(1, width-lipgloss.Width(marker)-lipgloss.Width(usage)-1)
				label = ansi.Truncate(label, labelWidth, "\u2026")
				padding := strings.Repeat(" ", max(1, width-lipgloss.Width(marker)-lipgloss.Width(label)-lipgloss.Width(usage)))
				if selected {
					line = SelectedRow.Width(width).Render(marker + label + padding + usage)
				} else {
					line = marker + Muted.Render(label) + padding + Muted.Render(usage)
				}
			}
			b.WriteString(ansi.Truncate(line, width, "\u2026") + "\n")
		}
	} else {
		start := listStart(m.Selected, visibleRows, len(rows))
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
	viewport.Height = max(1, height-2)
	viewport.SetContent(wrapLines(details, max(1, width)))
	content := SectionTitle.Render("Command Details") + "\n" + viewport.View()
	if viewport.TotalLineCount() > viewport.Height {
		hint := "↓ More details below"
		if viewport.AtTop() {
			hint = "↓ More details below"
		} else if viewport.AtBottom() {
			hint = "↑ More details above"
		} else {
			hint = "↑/↓ More details above and below"
		}
		content += "\n" + Muted.Render(hint)
	}
	return DetailPanel.Width(width + 2).Height(height).Render(content)
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
	b.WriteString(SectionTitle.Render("Environment") + "\n")
	b.WriteString("Shells: " + metadataText(command.Shells) + "\n")
	b.WriteString("Platforms: " + metadataText(command.Platforms) + "\n")
	workingDirectory := command.WorkingDirectory
	if workingDirectory == "" {
		workingDirectory = "Current directory"
	}
	b.WriteString("Working directory: " + workingDirectory + "\n")
	tools := metadataText(command.Requirements)
	if tools == "Not specified" {
		tools = "None listed"
	}
	b.WriteString("Tools: " + tools + "\n\n")
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

func metadataText(values []string) string {
	if len(values) == 0 {
		return "Not specified"
	}
	return strings.Join(values, ", ")
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
		if _, command, ok := m.selectedCommand(); ok {
			b.WriteString(Muted.Render(command.Name+" · "+command.Category) + "\n\n")
		}
		for i, label := range m.actionLabels() {
			label = "[" + strings.ToUpper(actionKeys[i]) + "] " + label
			if i == m.actionSelected {
				fmt.Fprintf(&b, "%s %s\n", Selected.Render("▶"), Selected.Render(label))
			} else {
				fmt.Fprintf(&b, "  %s\n", label)
			}
		}
		b.WriteString("\nC/R/F/E/D action  •  Up/Down move  •  Enter choose  •  Esc close")
		return b.String()
	case "help":
		return m.helpView.FullHelpView(keys.FullHelp())
	case "about":
		return strings.Join([]string{
			"About Snipbox",
			"",
			"Find and reuse reliable terminal commands.",
			"Inspect, customize templates, then copy or run.",
			"Rust fuzzy search. Go terminal UI. Your data stays local.",
			"History import and Git sync are opt-in.",
			"Created by Pavel Slepcov.",
			"Press M or right-click for command actions.",
			"",
			"Press Esc to close.",
		}, "\n")
	case "history":
		return "Shell History\n\n" + m.historyView.View() + "\nEnter imports  •  Esc closes"
	case "history_preview":
		return "Review History Import\n\nCommand:\n" + m.pendingImport + "\n\nEnter saves to your library  •  Esc returns to history"
	case "preview_run":
		return strings.Join([]string{
			"Run Command",
			"",
			"Shell: " + activeShellName(),
			"Working directory: " + m.runWorkingDirectory(),
			"",
			"Command:",
			m.pending,
			"",
			"Enter runs the command  •  Esc cancels",
		}, "\n")
	case "form":
		labels := []string{"Category", "Subcategory", "Name", "Description", "Tags (comma separated)", "Command", "Shells (comma separated)", "Platforms (comma separated)", "Required tools (comma separated)", "Working directory (optional)"}
		var b strings.Builder
		if m.editIndex < 0 {
			b.WriteString("New Command\n\n")
		} else {
			b.WriteString("Edit Command\n\n")
		}
		visible := max(1, m.Height-12)
		start := min(max(0, m.formField-visible+1), max(0, len(m.form)-visible))
		end := min(len(m.form), start+visible)
		if start > 0 {
			b.WriteString(Muted.Render("... earlier fields") + "\n")
		}
		for i := start; i < end && i < len(labels); i++ {
			label := labels[i]
			prefix, value := "  ", m.form[i]
			if i == m.formField {
				prefix, value = "> ", m.formInput.View()
			}
			fmt.Fprintf(&b, "%s%s: %s", prefix, label, value)
			if message := m.formErrors[i]; message != "" {
				b.WriteString(Danger.Render("  ! " + message))
			}
			b.WriteByte('\n')
		}
		if end < len(m.form) {
			b.WriteString(Muted.Render("... more fields below") + "\n")
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
		return "Confirm Execution\n\n[!] Review this potentially destructive command.\n" + activeShellName() + " • " + m.runWorkingDirectory() + "\n\n" + m.pending + "\n\nPress Y to run  •  N / Esc to cancel"
	case "confirm_delete":
		name := "the selected command"
		if m.pendingDelete >= 0 && m.pendingDelete < len(m.Commands) {
			name = m.Commands[m.pendingDelete].Name
		}
		return "Delete Command?\n\n" + name + " will be permanently removed from your local library.\n\nPress Y or Enter to delete  •  N / Esc to cancel"
	case "duplicate":
		if m.duplicateIndex >= 0 && m.duplicateIndex < len(m.Commands) {
			command := m.Commands[m.duplicateIndex]
			return "Command Already Saved\n\n" + command.Name + " · " + command.Category + "\n" + command.Command + "\n\nEnter opens the saved command  •  Esc returns to editing"
		}
		return "Command Already Saved\n\nEnter returns to editing  •  Esc closes"
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
