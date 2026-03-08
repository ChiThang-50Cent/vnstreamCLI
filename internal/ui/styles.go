package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	App      lipgloss.Style
	Title    lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
	Header   lipgloss.Style
	Footer   lipgloss.Style
	Notice   lipgloss.Style
	Busy     lipgloss.Style
	Help     lipgloss.Style
	Focused  lipgloss.Style
	Blurred  lipgloss.Style
	Highlight lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		App:       lipgloss.NewStyle().Background(lipgloss.Color("#1e1e2e")).Foreground(lipgloss.Color("#cdd6f4")),
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89b4fa")),
		Muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true),
		Header:    lipgloss.NewStyle().Foreground(lipgloss.Color("#74c7ec")).Bold(true),
		Footer:    lipgloss.NewStyle().Foreground(lipgloss.Color("#bac2de")),
		Notice:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
		Busy:      lipgloss.NewStyle().Foreground(lipgloss.Color("#f5c2e7")),
		Help:      lipgloss.NewStyle().Foreground(lipgloss.Color("#9399b2")),
		Focused:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true),
		Blurred:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9399b2")),
		Highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#89dceb")).Bold(true),
	}
}
