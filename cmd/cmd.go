package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "cli-of-life",
		RunE: run,
		Args: cobra.NoArgs,

		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cmd.Flags().StringP(config.FileFlag, "f", "", "Loads a pattern file on startup")
	if err := cmd.RegisterFlagCompletionFunc(config.FileFlag,
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{".rle", ".cells"}, cobra.ShellCompDirectiveFilterFileExt
		},
	); err != nil {
		panic(err)
	}

	cmd.Flags().String(config.CompletionFlag, "", "Output command-line completion code for the specified shell. Can be 'bash', 'zsh', 'fish', or 'powershell'.")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if shell := cmd.Flag(config.CompletionFlag).Value.String(); shell != "" {
		return completion(cmd, shell)
	}

	var tiles [][]int
	if file := cmd.Flag(config.FileFlag).Value.String(); file != "" {
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
