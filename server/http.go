package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpapi "github.com/elkoshar/bookcabin/api/http"
)

// runHTTPServer will run HTTP server with specified parameter and do gracefull shutdown if receive sigint or sigterm
func runHTTPServer(httpsrv httpapi.Server, port string) error {
	idleConnsClosed := make(chan struct{})
	go func() {

		signals := make(chan os.Signal, 1)

		// SIGHUP is for handling upstart reload
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		<-signals

		// when received an os signal, shut down.
		if err := httpsrv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			slog.Info(fmt.Sprintf("HTTP server Shutdown: %v", err))
		}
		close(idleConnsClosed)
	}()

	slog.Info(fmt.Sprintf("HTTP server listening on port %s", port))

	if err := httpsrv.Serve(port); err != http.ErrServerClosed {
		// Error starting or closing listener:
		return err
	}

	<-idleConnsClosed
	slog.Info("HTTP server shutdown gracefully")
	return nil
}
