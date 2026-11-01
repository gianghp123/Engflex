package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"engflex-api/config"
	"engflex-api/internal/server"
)

// TestPaymentPublicRoutesDoNotRequireAuth guards the IPN/return/plans
// routes against accidentally gaining RequireAuth: VNPay calls ipn
// server-to-server with no Clerk token, and return is a browser redirect
// target. Both rely on signature verification, not auth middleware, for
// security.
func TestPaymentPublicRoutesDoNotRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/payments/vnpay/plans"},
		{http.MethodGet, "/api/v1/payments/vnpay/ipn"},
		{http.MethodGet, "/api/v1/payments/vnpay/return"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := server.NewRouter(nil, config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost:3000"}}})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestPaymentIPN_RejectsMissingSignature confirms a callback with no
// vnp_SecureHash is rejected (ack RspCode 97) rather than panicking or
// falling through to a DB lookup with a nil connection.
func TestPaymentIPN_RejectsMissingSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := server.NewRouter(nil, config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost:3000"}}})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay/ipn?vnp_TxnRef=x", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"RspCode":"97"`)
}
