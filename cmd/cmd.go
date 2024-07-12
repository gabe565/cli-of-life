package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/game"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "cli-of-life",
		RunE: run,
	}
	cmd.Flags().StringP("file", "f", "", "Loads a pattern file on startup")
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	var tiles [][]int
	if file := cmd.Flag("file").Value.String(); file != "" {
		var err error
		if tiles, err = pattern.UnmarshalFile(file); err != nil {
			return err
		}
	}

	_, err := tea.NewProgram(
		game.New(tiles),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	).Run()
	return err
}
