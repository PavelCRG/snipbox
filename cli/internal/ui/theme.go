package ui

import "github.com/charmbracelet/lipgloss"

var (
	background     = lipgloss.Color("#1a1b26")
	foreground     = lipgloss.Color("#c0caf5")
	purple         = lipgloss.Color("#bb9af7")
	cyan           = lipgloss.Color("#7dcfff")
	muted          = lipgloss.Color("#565f89")
	categoryColors = map[string]lipgloss.Color{"Git": "#f7768e", "Docker": "#7dcfff", "System": "#9ece6a", "Network": "#e0af68", "Tools": "#bb9af7"}
	Title          = lipgloss.NewStyle().Bold(true).Foreground(purple)
	Panel          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(muted).Padding(0, 1)
	Selected       = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	Highlight      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9e64")).Bold(true)
	Variable       = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true)
	Base           = lipgloss.NewStyle().Foreground(foreground).Background(background)
)

func CategoryStyle(category string) lipgloss.Style {
	color, ok := categoryColors[category]
	if !ok {
		color = muted
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true)
}
