package config

import (
	"gabe565.com/cli-of-life/internal/rule"
)

type Config struct {
	Pattern string

	PatternFormat string
	RuleString    string
	Play          bool
	CacheLimit    int

	Completion string
}

func New() *Config {
	return &Config{
		PatternFormat: "auto",
		RuleString:    rule.GameOfLife().String(),
		CacheLimit:    10_000_000,
	}
}
