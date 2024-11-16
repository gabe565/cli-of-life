package config

import (
	"errors"

	"gabe565.com/cli-of-life/internal/rule"
	"github.com/spf13/cobra"
)

func RegisterCompletion(cmd *cobra.Command) error {
	return errors.Join(
		cmd.RegisterFlagCompletionFunc(RuleStringFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{rule.GameOfLife().String(), rule.HighLife().String()}, cobra.ShellCompDirectiveNoFileComp
			},
		),
		cmd.RegisterFlagCompletionFunc(PlayFlag, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(CacheLimitFlag, cobra.NoFileCompletions),
	)
}
