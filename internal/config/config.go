package config

import (
	"github.com/gabe565/cli-of-life/internal/rule"
)

type Config struct {
	File          string
	URL           string
	PatternFormat string
	RuleString    string
	Play          bool
	Width         uint
	Height        uint
	CacheLimit    uint

	Completion string
}

func New() *Config {
	return &Config{
		PatternFormat: "auto",
		RuleString:    rule.GameOfLife().String(),
		Width:         600,
		Height:        600,
		CacheLimit:    200_000,
	}
}
