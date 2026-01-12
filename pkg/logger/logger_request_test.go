package logger_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	config "github.com/elkoshar/bookcabin/configs"
	"github.com/elkoshar/bookcabin/pkg/logger"
)

func TestNewSlogWrapper(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "Test Success",
			config: &config.Config{
				LogLevel: "INFO",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.NewSlogWrapper(tt.config)
		})
	}
}

func TestNewLogEntry(t *testing.T) {
	cfg := &config.Config{}
	slogWrapper := logger.NewSlogWrapper(cfg)
	customLogFormater := logger.CustomLogFormatter{
		Logger: slogWrapper,
	}
	request := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
	}
	customLogFormater.NewLogEntry(request)
}

func TestWrite(t *testing.T) {
	cfg := &config.Config{}
	customLogEntry := &logger.CustomLogEntry{
		Logger: logger.NewSlogWrapper(cfg),
		Request: &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test"},
		},
	}
	status, bytes, elapsed := 200, 1234, time.Millisecond*500
	customLogEntry.Write(status, bytes, customLogEntry.Request.Header, elapsed, nil)
}

func TestPanic(t *testing.T) {
	cfg := &config.Config{}
	customLogEntry := &logger.CustomLogEntry{
		Logger: logger.NewSlogWrapper(cfg),
		Request: &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test"},
		},
	}
	panicValue := "Test panic value"
	stackTrace := []byte("stack trace here")
	customLogEntry.Panic(panicValue, stackTrace)
}
