// Package tests holds black-box tests for the internal/server package.
package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/config"
	"engflex-api/internal/server"
)

// TestSwaggerRouting covers which URLs reach the Swagger UI.
//
// gin-swagger v1.6.1 gates every request on a regexp that only accepts an
// allowlist of asset names (index.html, doc.json, the swagger-ui bundles). A
// bare /swagger or /swagger/ matches none of them, so it 404s from *inside*
// the handler — gin's route tree is never consulted and its own trailing-slash
// redirect sends browsers straight into that 404. Both bare forms must be
// routed to the UI entry point here or the documented /swagger URL is dead.
//
// The DB is nil because RegisterRoutes only stores the handle; no test here
// reaches a repository.
func TestSwaggerRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantTo   string
	}{
		{
			name:     "bare path lands on the UI entry point",
			path:     "/swagger",
			wantCode: http.StatusMovedPermanently,
			wantTo:   "/swagger/index.html",
		},
		{
			name:     "trailing slash lands on the UI entry point",
			path:     "/swagger/",
			wantCode: http.StatusMovedPermanently,
			wantTo:   "/swagger/index.html",
		},
		{
			name:     "UI entry point renders",
			path:     "/swagger/index.html",
			wantCode: http.StatusOK,
		},
		{
			name:     "spec is served",
			path:     "/swagger/doc.json",
			wantCode: http.StatusOK,
		},
		{
			name:     "health check is served",
			path:     "/health",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The CORS middleware rejects an empty allowlist at
			// construction (gin-contrib/cors panics), so pass one
			// origin here; these routes carry no CORS assertions.
			cfg := config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost:3000"}}}
			r := server.NewRouter(nil, cfg)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			require.Equal(t, tt.wantCode, w.Code)
			if tt.wantTo != "" {
				assert.Equal(t, tt.wantTo, w.Header().Get("Location"))
			}
		})
	}
}

// TestCORSAllowedOrigins pins the production switch away from AllowAllOrigins:
// only configured origins get an Access-Control-Allow-Origin header.
func TestCORSAllowedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const configured = "http://engflex-web.192-168-1-50.sslip.io"

	tests := []struct {
		name       string
		origin     string
		wantHeader string
	}{
		{name: "configured origin is echoed", origin: configured, wantHeader: configured},
		{name: "unconfigured origin is not allowed", origin: "http://evil.example", wantHeader: ""},
		{name: "no origin header stays headerless", origin: "", wantHeader: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{configured}}}
			r := server.NewRouter(nil, cfg)

			req := httptest.NewRequest(http.MethodOptions, "/api/v1/lessons", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
				req.Header.Set("Access-Control-Request-Method", http.MethodGet)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantHeader, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
