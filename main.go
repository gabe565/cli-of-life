package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/game"
)

func main() {
	program := tea.NewProgram(
		game.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := program.Run(); err != nil {
		slog.Error("Error running game", "error", err.Error())
		os.Exit(1)
	}
}
