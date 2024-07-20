package config

import (
	"errors"

	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/rule"
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
		cmd.RegisterFlagCompletionFunc(URLFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{"https://"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			},
		),
		cmd.RegisterFlagCompletionFunc(RuleStringFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return []string{rule.GameOfLife().String(), rule.HighLife().String()}, cobra.ShellCompDirectiveNoFileComp
			},
		),
		cmd.RegisterFlagCompletionFunc(PlayFlag, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(CacheLimitFlag, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(CompletionFlag,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return shells(), cobra.ShellCompDirectiveNoFileComp
			},
		),
	)
}
