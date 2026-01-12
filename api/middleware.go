package api

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type requestBody struct {
	Vertical string `json:"vertical"`
}

func InterceptorRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			slog.Debug(fmt.Sprintf("Request header = %v , Request URL = %v , Request Body = %v", r.Header, r.URL, r.Body))

			ctx := r.Context()

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func GetIP(r *http.Request) string {
	// First, check for the X-Forwarded-For header (used in proxies)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, the client IP is the first one
		ip = strings.Split(ip, ",")[0]
		return strings.TrimSpace(ip)
	}

	// If X-Forwarded-For is empty, fall back to r.RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	return ip
}

func newStatusRecoder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func GetPathName(r *http.Request) string {
	path := r.URL.Path
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		path = rctx.RoutePattern()
	}
	return fmt.Sprintf("%s:%s", r.Method, path)
}

func NewMetricMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			recorder := newStatusRecoder(w)

			next.ServeHTTP(recorder, r)

			if strings.HasPrefix(r.URL.Path, "/application/health") {
				return
			}

		})
	}
}
