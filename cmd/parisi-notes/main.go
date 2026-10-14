package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"parisi-notes/internal/app"
)

func main() {
	m, err := app.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
