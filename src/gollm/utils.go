package gollm

import (
	"log/slog"
	"os"
)

func defaultLogger(level slog.Leveler) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
