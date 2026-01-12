package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elkoshar/bookcabin/api"
	config "github.com/elkoshar/bookcabin/configs"
	"github.com/elkoshar/bookcabin/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareHttp(t *testing.T) {
	logger.InitLogger(&config.Config{})

	tests := []struct {
		name       string
		method     string
		path       string
		key        string
		header     map[string]string
		wantResult bool
		res        int
	}{
		{
			name:   "Test success",
			method: "GET",
			path:   "/users/10",
			key:    "id",
			header: map[string]string{
				"X-Language": "en",
			},
			wantResult: true,
		}, {
			name:       "Test success no lang",
			method:     "GET",
			path:       "/users/10",
			key:        "id",
			header:     map[string]string{"X-Language": ""},
			wantResult: true,
		}, {
			name:       "Test success no header",
			method:     "GET",
			path:       "/users/10",
			key:        "id",
			header:     map[string]string{},
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.With(api.InterceptorRequest()).Route("/users", func(r chi.Router) {
				r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
				})
			})

			ts := httptest.NewServer(r)
			defer ts.Close()

			err := testRequest(t, ts, tt.method, tt.path, tt.header)
			if tt.wantResult {
				assert.NoError(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header map[string]string) error {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	if err != nil {
		return err
	}
	for key, val := range header {
		req.Header.Set(key, val)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return err
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "With X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "Without X-Forwarded-For header, with RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "10.0.0.1",
		},
		{
			name:       "Invalid RemoteAddr format",
			headers:    map[string]string{},
			remoteAddr: "invalid-addr",
			expectedIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header:     make(http.Header),
				RemoteAddr: tt.remoteAddr,
			}

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Call the function
			got := api.GetIP(req)
			if got != tt.expectedIP {
				t.Errorf("getIP() = %v, want %v", got, tt.expectedIP)
			}
		})
	}
}
