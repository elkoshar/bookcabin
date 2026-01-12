package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	config "github.com/elkoshar/bookcabin/configs"
	"github.com/go-chi/chi/v5/middleware"
)

type (
	SlogWrapper struct {
		logger *slog.Logger
	}
	CustomLogEntry struct {
		Logger  *SlogWrapper
		Request *http.Request
	}
	CustomLogFormatter struct {
		Logger *SlogWrapper
	}
)

func (c CustomLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &CustomLogEntry{
		Logger:  c.Logger,
		Request: r,
	}
}

func NewSlogWrapper(cfg *config.Config) *SlogWrapper {
	logger := getLogger(cfg)
	return &SlogWrapper{logger: logger}
}

func (e *CustomLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	slog.Info(
		fmt.Sprintf("Request: %s %s, Status: %d, Bytes: %d, Elapsed: %s",
			e.Request.Method,
			e.Request.URL.Path,
			status,
			bytes,
			elapsed,
		),
	)
}

func (e *CustomLogEntry) Panic(v interface{}, stack []byte) {
	slog.Info(fmt.Sprintf("Panic: %+v, Stack: %s", v, stack))
}
