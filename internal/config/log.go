package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func InitLog(level slog.Level) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))
}
