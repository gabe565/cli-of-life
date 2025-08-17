package config

import (
	"bytes"
	"log/slog"

	"gabe565.com/utils/must"
	"github.com/spf13/cobra"
)

const (
	RuleStringFlag = "rule-string"
	PlayFlag       = "play"
	CacheLimitFlag = "cache-limit"

	// Deprecated: Pass file as positional argument instead.
	FileFlag = "file"
	// Deprecated: Pass URL as positional argument instead.
	URLFlag = "url"
)

func (c *Config) RegisterFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.StringVar(&c.RuleString, RuleStringFlag, c.RuleString,
		"Rule string to use. This will be ignored if a pattern file is loaded.",
	)
	fs.BoolVar(&c.Play, PlayFlag, c.Play, "Play on startup")
	fs.IntVar(&c.CacheLimit, CacheLimitFlag, c.CacheLimit,
		"Maximum number of entries to keep cached. Higher values will use more memory, but less CPU.",
	)

	fs.StringVarP(&c.Pattern, FileFlag, "f", c.Pattern, "Load a pattern file")
	fs.StringVar(&c.Pattern, URLFlag, c.Pattern, "Load a pattern URL")
	must.Must(fs.MarkDeprecated(FileFlag, "pass file as positional argument instead."))
	must.Must(fs.MarkDeprecated(URLFlag, "pass URL as positional argument instead."))
	fs.SetOutput(DeprecatedWriter{})
}

type DeprecatedWriter struct{}

func (d DeprecatedWriter) Write(b []byte) (int, error) {
	slog.Warn(string(bytes.TrimSpace(b)))
	return len(b), nil
}
