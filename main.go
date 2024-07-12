package main

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/game"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/pflag"
)

func main() {
	var file string
	pflag.StringVarP(&file, "file", "f", "", "Loads a pattern file on startup")
	pflag.Parse()

	var tiles [][]int
	if file != "" {
		var err error
		if tiles, err = pattern.UnmarshalFile(file); err != nil {
			slog.Error("Failed to load tiles", "error", err.Error())
			os.Exit(1)
		}
	}

	program := tea.NewProgram(
		game.New(tiles),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := program.Run(); err != nil {
		slog.Error("Error running game", "error", err.Error())
		os.Exit(1)
	}
}
