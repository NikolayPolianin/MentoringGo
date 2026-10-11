package logger

import (
	"log/slog"
	"os"
)

func init() {
	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	).With("service", "One-Service").With("env", "production")

	slog.SetDefault(logger)
}
