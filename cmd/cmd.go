package cmd

import (
	"context"

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

	conf := config.New()
	conf.RegisterFlags(cmd.Flags())
	if err := config.RegisterCompletion(cmd); err != nil {
		panic(err)
	}
	cmd.SetContext(config.NewContext(context.Background(), conf))
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	conf, ok := config.FromContext(cmd.Context())
	if !ok {
		panic("command missing config")
	}

	if conf.Completion != "" {
		return completion(cmd, conf.Completion)
	}

	var rule pattern.Rule
	if err := rule.UnmarshalText([]byte(conf.RuleString)); err != nil {
		return err
	}

	pat := pattern.Pattern{
		Rule: rule,
	}
	if conf.File != "" {
		var err error
		if pat, err = pattern.UnmarshalFile(conf.File, pattern.Format(conf.FileFormat)); err != nil {
			return err
		}
	}

	g := game.New(
		game.WithPattern(pat),
		game.WithDimensions(conf.Width, conf.Height),
		game.WithPlay(conf.Play),
	)

	_, err := tea.NewProgram(g, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
	return err
}
