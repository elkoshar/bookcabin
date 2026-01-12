package logger

import (
	"log/slog"
	"os"
	"strings"

	config "github.com/elkoshar/bookcabin/configs"
)

func InitLogger(lc *config.Config) {
	logger := getLogger(lc)
	slog.SetDefault(logger)
}

func getLevel(logLevelConfig string) slog.Level {
	switch strings.ToLower(logLevelConfig) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getLogger(c *config.Config) *slog.Logger {
	level := getLevel(c.LogLevel)
	//handler := tixlog.NewStdoutHandler(level)
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		AddSource: true, // Recommended: adds the file and line number to the log
		Level:     level,
	}

	switch strings.ToLower(c.LogFormat) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json": // Recommended for production environments
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		// Default to JSON for consistency
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
