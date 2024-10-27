package main

import (
	"log/slog"
	"os"

	"gabe565.com/cli-of-life/cmd"
)

var version = "beta"

func main() {
	root := cmd.New(cmd.WithVersion(version))
	if err := root.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
