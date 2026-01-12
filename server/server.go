package server

import (
	httpapi "github.com/elkoshar/bookcabin/api/http"
	config "github.com/elkoshar/bookcabin/configs"
)

func InitHttp(config *config.Config) error {

	httpserver := httpapi.NewServer(config)

	return runHTTPServer(*httpserver, config.ServerHttpPort)
}
