package config

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"

	"github.com/spf13/pflag"
)

const (
	RuleStringFlag = "rule-string"
	PlayFlag       = "play"
	CacheLimitFlag = "cache-limit"
	CompletionFlag = "completion"

	// Deprecated: Pass file as positional argument instead
	FileFlag = "file"
	// Deprecated: Pass URL as positional argument instead
	URLFlag = "url"
)

func (c *Config) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.RuleString, RuleStringFlag, c.RuleString, "Rule string to use. This will be ignored if a pattern file is loaded.")
	fs.BoolVar(&c.Play, PlayFlag, c.Play, "Play on startup")
	fs.IntVar(&c.CacheLimit, CacheLimitFlag, c.CacheLimit, "Maximum number of entries to keep cached. Higher values will use more memory, but less CPU.")
	fs.StringVar(&c.Completion, CompletionFlag, c.Completion, "Output command-line completion code for the specified shell (one of: "+strings.Join(shells(), ", ")+")")

	fs.StringVarP(&c.Pattern, FileFlag, "f", c.Pattern, "Load a pattern file")
	fs.StringVar(&c.Pattern, URLFlag, c.Pattern, "Load a pattern URL")
	if err := errors.Join(
		fs.MarkDeprecated(FileFlag, "pass file as positional argument instead."),
		fs.MarkDeprecated(URLFlag, "pass URL as positional argument instead."),
	); err != nil {
		panic(err)
	}
	fs.SetOutput(DeprecatedWriter{})
}

type DeprecatedWriter struct{}

func (d DeprecatedWriter) Write(b []byte) (int, error) {
	slog.Warn(string(bytes.TrimSpace(b)))
	return len(b), nil
}
