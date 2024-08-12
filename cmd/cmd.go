package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
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
		Use:   "cli-of-life [file | url]",
		Short: "Play Conway's Game of Life in your terminal",
		RunE:  run,
		Args:  cobra.MaximumNArgs(1),

		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{pattern.ExtRLE, pattern.ExtPlaintext}, cobra.ShellCompDirectiveFilterFileExt
		},
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

	var r rule.Rule
	if err := r.UnmarshalText([]byte(conf.RuleString)); err != nil {
		return err
	}

	if len(args) == 1 {
		conf.Pattern = args[0]
	}

	var pat pattern.Pattern
	var err error
	switch {
	case conf.Pattern != "":
		u, err := url.Parse(conf.Pattern)
		if err != nil {
			return err
		}

		switch u.Scheme {
		case "http", "https":
			slog.Info("Loading pattern URL", "url", conf.Pattern)
			if pat, err = pattern.UnmarshalURL(context.Background(), conf.Pattern); err != nil {
				return err
			}
		default:
			slog.Info("Loading pattern file", "path", conf.Pattern)
			if pat, err = pattern.UnmarshalFile(conf.Pattern); err != nil {
				return err
			}
		}
		slog.Info("Loaded pattern", "pattern", pat)
	case !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()):
		slog.Info("Loading pattern from stdin")
		if pat, err = pattern.Unmarshal(os.Stdin); err != nil {
			return err
		}
		slog.Info("Loaded pattern", "pattern", pat)
	default:
		pat = pattern.Pattern{
			Rule: r,
			Tree: quadtree.New(),
		}
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
