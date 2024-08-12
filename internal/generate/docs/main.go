package main

import (
	"os"

	"github.com/gabe565/cli-of-life/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	output := "./docs"

	if err := os.MkdirAll(output, 0o755); err != nil {
		panic(err)
	}

	root := cmd.New(cmd.WithVersion("beta"))
	if err := doc.GenMarkdownTree(root, output); err != nil {
		panic(err)
	}
}
