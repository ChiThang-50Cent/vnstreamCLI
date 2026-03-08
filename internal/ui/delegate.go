package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Separator interface — any item can implement this to be rendered as a section header.
type SeparatorItem interface {
	IsSeparator() bool
}

// CompactDelegate renders each list item as a single line (like fzf).
type CompactDelegate struct {
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	sepStyle      lipgloss.Style
}

func NewCompactDelegate() CompactDelegate {
	return CompactDelegate{
		selectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		normalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")),
		sepStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")).
			Italic(true),
	}
}

func (d CompactDelegate) Height() int                             { return 1 }
func (d CompactDelegate) Spacing() int                            { return 0 }
func (d CompactDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d CompactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title := item.FilterValue()
	if i, ok := item.(interface{ Title() string }); ok {
		title = i.Title()
	}

	if sep, ok := item.(SeparatorItem); ok && sep.IsSeparator() {
		fmt.Fprint(w, d.sepStyle.Render(title))
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, d.selectedStyle.Render("> "+title))
	} else {
		fmt.Fprint(w, d.normalStyle.Render("  "+title))
	}
}
