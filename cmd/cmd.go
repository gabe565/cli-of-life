package cmd

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli-of-life",
		Short: "Play Conway's Game of Life in your terminal",
		RunE:  run,
		Args:  cobra.NoArgs,

		ValidArgsFunction: cobra.NoFileCompletions,
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP(config.FileFlag, "f", "", "Loads a pattern file on startup")
	cmd.Flags().String(config.FileFormatFlag, "auto", "File format (one of: "+strings.Join(pattern.FormatStrings(), ", ")+")")
	cmd.Flags().String(config.RuleStringFlag, pattern.GameOfLife().String(), "Rule string to use. This will be ignored if a pattern file is loaded.")
	cmd.Flags().String(config.CompletionFlag, "", "Output command-line completion code for the specified shell (one of: "+strings.Join(shells(), ", ")+")")

	if err := errors.Join(
		cmd.RegisterFlagCompletionFunc(config.FileFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{pattern.ExtRLE, pattern.ExtPlaintext}, cobra.ShellCompDirectiveFilterFileExt
			},
		),
		cmd.RegisterFlagCompletionFunc(config.FileFormatFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return pattern.FormatStrings(), cobra.ShellCompDirectiveNoFileComp
			},
		),
		cmd.RegisterFlagCompletionFunc(config.RuleStringFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{pattern.GameOfLife().String(), pattern.HighLife().String()}, cobra.ShellCompDirectiveNoFileComp
			},
		),
	); err != nil {
		panic(err)
	}

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if shell := cmd.Flag(config.CompletionFlag).Value.String(); shell != "" {
		return completion(cmd, shell)
	}

	var rule pattern.Rule
	if err := rule.UnmarshalText([]byte(cmd.Flag(config.RuleStringFlag).Value.String())); err != nil {
		return err
	}

	pat := pattern.Pattern{
		Rule: rule,
	}
	if file := cmd.Flag(config.FileFlag).Value.String(); file != "" {
		format := pattern.Format(cmd.Flag(config.FileFormatFlag).Value.String())
		var err error
		if pat, err = pattern.UnmarshalFile(file, format); err != nil {
			return err
		}
	}

	_, err := tea.NewProgram(
		game.New(pat),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	).Run()
	return err
}
