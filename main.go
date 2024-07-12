package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	program := tea.NewProgram(
		New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := program.Run(); err != nil {
		slog.Error("Error running game", "error", err.Error())
		os.Exit(1)
	}
}
