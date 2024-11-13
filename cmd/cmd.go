package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/game"
	"gabe565.com/cli-of-life/internal/pattern"
	"gabe565.com/cli-of-life/internal/pprof"
	"gabe565.com/cli-of-life/internal/quadtree"
	"gabe565.com/utils/cobrax"
	"gabe565.com/utils/must"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func New(opts ...cobrax.Option) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli-of-life [file | url]",
		Short: "Play Conway's Game of Life in your terminal",
		RunE:  run,
		Args:  cobra.MaximumNArgs(1),

		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return pattern.Extensions(), cobra.ShellCompDirectiveFilterFileExt
		},
		DisableAutoGenTag: true,
		SilenceErrors:     true,
	}

	config.InitLog(slog.LevelInfo)
	conf := config.New()
	conf.RegisterFlags(cmd.Flags())
	must.Must(config.RegisterCompletion(cmd))
	cmd.SetContext(config.NewContext(context.Background(), conf))

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if pprof.Enabled {
		go func() {
			if err := pprof.ListenAndServe(); err != nil {
				slog.Error("Failed to start debug server", "error", err.Error())
			}
		}()
	}

	conf, ok := config.FromContext(cmd.Context())
	if !ok {
		panic("command missing config")
	}

	if conf.Completion != "" {
		return completion(cmd, conf.Completion)
	}

	slog.Info("cli-of-life", "version", cobrax.GetVersion(cmd), "commit", cobrax.GetCommit(cmd))

	if len(args) == 1 {
		conf.Pattern = args[0]
	}

	if conf.CacheLimit > 0 {
		quadtree.SetMaxCache(conf.CacheLimit)
	}

	program := tea.NewProgram(
		game.New(conf),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
		tea.WithoutCatchPanics(),
	)

	defer func() {
		if err := recover(); err != nil {
			program.Kill()
			_ = program.ReleaseTerminal()
			//nolint:forbidigo
			fmt.Printf("Caught panic:\n\n%s\n\nRestoring terminal...\n\n", err)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	slog.Info("Starting game")
	config.InitLog(slog.LevelWarn)
	_, err := program.Run()
	slog.Info("Quitting game")
	return err
}
