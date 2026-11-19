package ui

import "github.com/charmbracelet/lipgloss"

var (
	background = lipgloss.AdaptiveColor{Light: "#f7f7fb", Dark: "#1a1b26"}
	foreground = lipgloss.AdaptiveColor{Light: "#24283b", Dark: "#c0caf5"}
	muted      = lipgloss.AdaptiveColor{Light: "#707a9a", Dark: "#565f89"}
	border     = lipgloss.AdaptiveColor{Light: "#a9b1d6", Dark: "#414868"}
	blue       = lipgloss.Color("#7aa2f7")
	purple     = lipgloss.Color("#bb9af7")
	cyan       = lipgloss.Color("#2ac3de")
	category   = map[string]lipgloss.Color{
		"Git": "#ff9e64", "PowerShell": "#7dcfff", "Docker": "#7dcfff", "Linux/Bash": "#9ece6a", "Network": "#bb9af7",
	}
	Title        = lipgloss.NewStyle().Bold(true).Foreground(blue)
	Header       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(blue).Background(background).Padding(0, 1)
	Panel        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Background(background).Foreground(foreground).Padding(0, 1)
	ActivePanel  = Panel.BorderForeground(blue)
	DetailPanel  = Panel.BorderForeground(purple)
	Selected     = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	SelectedRow  = lipgloss.NewStyle().Foreground(cyan).Background(lipgloss.Color("#283457")).Bold(true)
	Highlight    = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	Variable     = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Bold(true)
	SectionTitle = lipgloss.NewStyle().Foreground(purple).Bold(true)
	Muted        = lipgloss.NewStyle().Foreground(muted)
	CommandBlock = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#587539", Dark: "#9ece6a"}).Background(lipgloss.AdaptiveColor{Light: "#e9eaf2", Dark: "#24283b"}).Padding(0, 1)
	Danger       = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Background(lipgloss.AdaptiveColor{Light: "#fff0f1", Dark: "#35212d"}).Bold(true).Padding(0, 1)
	Success      = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Bold(true).Padding(0, 1)
	Base         = lipgloss.NewStyle().Foreground(foreground).Background(background)
	ModalOuter   = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(blue).Background(background).Padding(0, 1)
	ModalInner   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Background(background).Foreground(foreground).Padding(1, 2)
)

func CategoryStyle(name string) lipgloss.Style {
	color, ok := category[name]
	if !ok {
		color = lipgloss.Color("#565f89")
	}
	return lipgloss.NewStyle().Foreground(background).Background(color).Bold(true).Padding(0, 1)
}

func CategoryTextStyle(name string) lipgloss.Style {
	color, ok := category[name]
	if !ok {
		color = lipgloss.Color("#565f89")
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true)
}
