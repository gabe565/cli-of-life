package config

import "github.com/gabe565/cli-of-life/internal/pattern"

type Config struct {
	File       string
	FileFormat string
	RuleString string
	Play       bool

	Completion string
}

func New() *Config {
	return &Config{
		FileFormat: "auto",
		RuleString: pattern.GameOfLife().String(),
	}
}
