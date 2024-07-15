package cmd

import (
	"errors"
	"fmt"

	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/spf13/cobra"
)

var ErrInvalidShell = errors.New("invalid shell")

func completion(cmd *cobra.Command, shell string) error {
	switch shell {
	case config.ShellBash:
		return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
	case config.ShellZsh:
		return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
	case config.ShellFish:
		return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
	case config.ShellPowerShell:
		return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
	default:
		return fmt.Errorf("%w: %s", ErrInvalidShell, shell)
	}
}
