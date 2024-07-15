package config

import (
	"strings"

	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/spf13/pflag"
)

const (
	FileFlag       = "file"
	FileFormatFlag = "file-format"
	RuleStringFlag = "rule-string"
	PlayFlag       = "play"
	WidthFlag      = "width"
	HeightFlag     = "height"
	CompletionFlag = "completion"
)

func (c *Config) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.File, FileFlag, "f", c.File, "Loads a pattern file on startup")
	fs.StringVar(&c.FileFormat, FileFormatFlag, c.FileFormat, "File format (one of: "+strings.Join(pattern.FormatStrings(), ", ")+")")
	fs.StringVar(&c.RuleString, RuleStringFlag, c.RuleString, "Rule string to use. This will be ignored if a pattern file is loaded.")
	fs.BoolVar(&c.Play, PlayFlag, c.Play, "Play on startup")
	fs.UintVar(&c.Width, WidthFlag, c.Width, "Board width. Will be ignored if a larger pattern is loaded.")
	fs.UintVar(&c.Height, HeightFlag, c.Height, "Board height. Will be ignored if a larger pattern is loaded.")
	fs.StringVar(&c.Completion, CompletionFlag, c.Completion, "Output command-line completion code for the specified shell (one of: "+strings.Join(shells(), ", ")+")")
}
