//go:build pprof

package pprof

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const Enabled = true

func ListenAndServe() error {
	addr := "127.0.0.1:6060"
	server := http.Server{
		Addr:        addr,
		ReadTimeout: 3 * time.Second,
	}
	slog.Info("Starting debug sever", "address", addr)
	return server.ListenAndServe()
}
