package http

import (
	"encoding/json"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/api/http/aggregator"
	config "github.com/elkoshar/bookcabin/configs"
	"github.com/elkoshar/bookcabin/pkg/helpers"
	"github.com/elkoshar/bookcabin/pkg/logger"
	"github.com/elkoshar/bookcabin/pkg/panics"
)

func root(w http.ResponseWriter, r *http.Request) {
	app := map[string]interface{}{
		"name":        "bookcabin",
		"description": "Flight Search and Aggregation API",
	}

	data, _ := json.Marshal(app)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handler(checker api.HealthChecker, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(panics.HTTPRecoveryMiddleware)
	r.Use(middleware.Timeout(cfg.HttpInboundTimeout))

	r.Get("/application/health", checker.HealthChi)
	r.Get("/", root)
	r.Handle("/metrics", promhttp.Handler())
	if helpers.GetEnvString() != helpers.EnvProduction {
		r.Get("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.json")
		}))

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger.json"),
		))

	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestLogger(&logger.CustomLogFormatter{Logger: logger.NewSlogWrapper(cfg)}))

		cors := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		})
		r.Use(cors.Handler)

		r.Handle("/panics", panics.CaptureHandler(func(w http.ResponseWriter, r *http.Request) {
			panic("Panics from /test/panics endpoint")
		}))

		r.With(api.InterceptorRequest()).Route("/bookcabin", func(r chi.Router) {
			r.Use(api.NewMetricMiddleware())

			r.Route("/flight/search", func(r chi.Router) {
				r.Post("/", aggregator.Search)
			})

		})
	})

	return r
}
