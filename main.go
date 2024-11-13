package main

import (
	"log/slog"
	"net/http"
	"os"

	"gabe565.com/cli-of-life/cmd"
	"gabe565.com/utils/cobrax"
	"gabe565.com/utils/httpx"
)

var version = "beta"

func main() {
	root := cmd.New(cobrax.WithVersion(version))
	http.DefaultTransport = httpx.NewUserAgentTransport(nil, cobrax.BuildUserAgent(root))
	if err := root.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
