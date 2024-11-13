package config

import (
	"log/slog"
	"os"
	"time"

	"gabe565.com/utils/termx"
	"github.com/lmittmann/tint"
)

func InitLog(level slog.Level) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
			NoColor:    !termx.IsColor(os.Stderr),
		}),
	))
}
