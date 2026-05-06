package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	env := os.Getenv("APP_ENV")

	var handler slog.Handler

	if env == "production" {
		// JSON estruturado em produção
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Texto legível em desenvolvimento
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
