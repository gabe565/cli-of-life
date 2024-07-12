package cmd

import (
	"errors"
	"fmt"

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

var ErrInvalidShell = errors.New("invalid shell")

func completion(cmd *cobra.Command, shell string) error {
	switch shell {
	case ShellBash:
		return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
	case ShellZsh:
		return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
	case ShellFish:
		return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
	case ShellPowerShell:
		return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
	default:
		return fmt.Errorf("%w: %s", ErrInvalidShell, shell)
	}
}
