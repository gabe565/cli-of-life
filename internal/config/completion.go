package config

import (
	"errors"

	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/cobra"
)

const (
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellFish       = "fish"
	ShellPowerShell = "powershell"
)

func shells() []string {
	return []string{ShellBash, ShellZsh, ShellFish, ShellPowerShell}
}

func RegisterCompletion(cmd *cobra.Command) error {
	return errors.Join(
		cmd.RegisterFlagCompletionFunc(FileFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{pattern.ExtRLE, pattern.ExtPlaintext}, cobra.ShellCompDirectiveFilterFileExt
			},
		),
		cmd.RegisterFlagCompletionFunc(RuleStringFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{pattern.GameOfLife().String(), pattern.HighLife().String()}, cobra.ShellCompDirectiveNoFileComp
			},
		),
	)
}
