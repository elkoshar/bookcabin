package http

import (
	"context"
	"net"
	"net/http"

	"github.com/elkoshar/bookcabin/api"
	config "github.com/elkoshar/bookcabin/configs"
)

// Server struct
type Server struct {
	server *http.Server
	Cfg    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		server: &http.Server{
			Handler: handler(api.HealthChecker{}, cfg),
		},
		Cfg: cfg,
	}
}

var ()

// Serve will run an HTTP server
func (s *Server) Serve(port string) error {

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

// Shutdown will tear down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
