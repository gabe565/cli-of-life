package config

import (
	"strings"

	"github.com/spf13/pflag"
)

const (
	FileFlag       = "file"
	URLFlag        = "url"
	RuleStringFlag = "rule-string"
	PlayFlag       = "play"
	CacheLimitFlag = "cache-limit"
	CompletionFlag = "completion"
)

func (c *Config) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.File, FileFlag, "f", c.File, "Load a pattern file")
	fs.StringVar(&c.URL, URLFlag, c.URL, "Load a pattern URL")
	fs.StringVar(&c.RuleString, RuleStringFlag, c.RuleString, "Rule string to use. This will be ignored if a pattern file is loaded.")
	fs.BoolVar(&c.Play, PlayFlag, c.Play, "Play on startup")
	fs.UintVar(&c.CacheLimit, CacheLimitFlag, c.CacheLimit, "Maximum number of entries to keep cached. Higher values will use more memory, but less CPU.")
	fs.StringVar(&c.Completion, CompletionFlag, c.Completion, "Output command-line completion code for the specified shell (one of: "+strings.Join(shells(), ", ")+")")
}
