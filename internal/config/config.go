package config

import "github.com/gabe565/cli-of-life/internal/pattern"

type Config struct {
	File       string
	FileFormat string
	RuleString string
	Play       bool
	Width      uint
	Height     uint

	Completion string
}

func New() *Config {
	return &Config{
		FileFormat: "auto",
		RuleString: pattern.GameOfLife().String(),
		Width:      400,
		Height:     400,
	}
}
