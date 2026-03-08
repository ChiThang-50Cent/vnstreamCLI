package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ChiThang-50Cent/vnstream/internal/config"
	"github.com/ChiThang-50Cent/vnstream/internal/ui"
)

func main() {
	cfg := config.Default()
	initialQuery := ""
	if len(os.Args) > 1 {
		initialQuery = strings.TrimSpace(strings.Join(os.Args[1:], " "))
	}

	app, err := ui.NewAppModel(cfg, initialQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		os.Exit(1)
	}
}
