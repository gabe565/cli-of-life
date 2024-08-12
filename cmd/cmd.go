package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/pprof"
	"github.com/gabe565/cli-of-life/internal/quadtree"
	"github.com/gabe565/cli-of-life/internal/rule"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func New(opts ...Option) *cobra.Command {
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

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
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

	var r rule.Rule
	if err := r.UnmarshalText([]byte(conf.RuleString)); err != nil {
		return err
	}

	var pat pattern.Pattern
	var err error
	switch {
	case conf.File != "":
		slog.Info("Loading pattern file", "path", conf.File)
		if pat, err = pattern.UnmarshalFile(conf.File); err != nil {
			return err
		}
	case conf.URL != "":
		slog.Info("Loading pattern URL", "url", conf.URL)
		if pat, err = pattern.UnmarshalURL(context.Background(), conf.URL); err != nil {
			return err
		}
	case !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()):
		slog.Info("Loading pattern from stdin")
		if pat, err = pattern.Unmarshal(os.Stdin); err != nil {
			return err
		}
	default:
		pat = pattern.Pattern{
			Rule: r,
			Tree: quadtree.New(),
		}
	}
	if pat.Name != "" {
		slog.Info("Loaded pattern", "pattern", pat)
	}

	pat.Tree.SetMaxCache(conf.CacheLimit)

	program := tea.NewProgram(
		game.New(game.WithPattern(pat), game.WithConfig(conf)),
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
	_, err = program.Run()
	slog.Info("Quitting game")
	return err
}
