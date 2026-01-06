package utils

import (
	"log/slog"
	"os"
)

var Slogger *slog.Logger

func InitLogger(verbose bool) {
	if verbose {
		Slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		Slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
