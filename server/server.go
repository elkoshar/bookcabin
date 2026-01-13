package server

import (
	"github.com/elkoshar/bookcabin/api"
	httpapi "github.com/elkoshar/bookcabin/api/http"
	config "github.com/elkoshar/bookcabin/configs"
	"github.com/elkoshar/bookcabin/service/aggregator"
	"github.com/elkoshar/bookcabin/service/airasia"
	"github.com/elkoshar/bookcabin/service/batik"
	"github.com/elkoshar/bookcabin/service/garuda"
	"github.com/elkoshar/bookcabin/service/lion"
)

func InitHttp(config *config.Config) error {

	garudaProvider := garuda.New(
		config.GarudaPath,
	)
	lionProvider := lion.New(config.LionPath)
	airAsiaProvider := airasia.New(config.AirAsiaPath)
	batikProvider := batik.New(config.BatikPath)

	aggregator := aggregator.NewAggregator(
		config.AggregatorTimeout,
		garudaProvider,
		lionProvider,
		airAsiaProvider,
		batikProvider,
	)

	httpserver := httpapi.Server{
		Cfg:         config,
		HealthCheck: api.HealthChecker{},
		Aggregator:  aggregator,
	}

	return runHTTPServer(httpserver, config.ServerHttpPort)
}
