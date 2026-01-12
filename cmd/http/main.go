package main

import (
	"fmt"
	"log/slog"
	"os"

	config "github.com/elkoshar/bookcabin/configs"
	"github.com/elkoshar/bookcabin/pkg/logger"
	"github.com/elkoshar/bookcabin/server"
)

// @title Flight Search and Aggregation API
// @version 0.1
// @description This service is to handle Flight Search and Aggregation API. For more detail, please visit https://github.com/elkoshar/bookcabin
// @contact.name Elko Sharhadi Eppasa
// @contact.url https://github.com/elkoshar
// @contact.email elko.s.eppasa@mail.com
// @BasePath /bookcabin
func main() {

	var (
		cfg *config.Config
	)

	// init config
	err := config.Init(
		config.WithConfigFile("config"),
		config.WithConfigType("env"),
	)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to initialize config: %v", err))
		os.Exit(1)
	}
	cfg = config.Get()

	//init logging
	logger.InitLogger(cfg)

	// // init send to Slack when panics
	// panics.SetOptions(&panics.Options{
	// 	Env: helpers.GetEnvString(),
	// })

	// init all DI for service handler implementation
	if err := server.InitHttp(cfg); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}
