package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/api/http/aggregator"
	config "github.com/elkoshar/bookcabin/configs"
)

type Server struct {
	server      *http.Server
	Cfg         *config.Config
	HealthCheck api.HealthChecker
	Aggregator  api.FlightAggregator
}

var ()

func (s *Server) Serve(port string) error {

	aggregator.Init(s.Aggregator)

	s.server = &http.Server{
		ReadTimeout:  s.Cfg.HttpReadTimeout * time.Second,
		WriteTimeout: s.Cfg.HttpWriteTimeout * time.Second,
		Handler:      handler(s.HealthCheck, s.Cfg),
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
