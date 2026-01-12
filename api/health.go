package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

type HealthChecker struct {
}

func (h *HealthChecker) health(ctx context.Context) map[string]interface{} {
	OK := "OK"

	applicationStatus := OK

	resp := map[string]interface{}{
		"name": os.Args[0],
		"status": map[string]string{
			"application": applicationStatus,
		},
	}

	return resp
}

func (h *HealthChecker) HealthChi(w http.ResponseWriter, r *http.Request) {
	resp := h.health(r.Context())
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *HealthChecker) HealthEcho(c echo.Context) error {
	resp := h.health(c.Request().Context())
	return c.JSON(200, resp)
}
