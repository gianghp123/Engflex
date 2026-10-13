package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3"

	"engflex-api/internal/server/middleware"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		rate         string
		requests     int
		wantLastCode int
		wantErr      bool
	}{
		{name: "invalid rate errors", rate: "not-a-rate", requests: 1, wantErr: true},
		{name: "empty rate errors", rate: "", requests: 1, wantErr: true},
		{name: "allows under limit", rate: "5-M", requests: 2, wantLastCode: http.StatusOK},
		{name: "rejects over limit", rate: "2-M", requests: 3, wantLastCode: http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := middleware.NewInMemoryRateLimitMiddleware(tt.rate)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, h)

			r := gin.New()
			r.GET("/ping", h, func(c *gin.Context) { c.Status(http.StatusOK) })
			rec := httptest.NewRecorder()
			for i := 0; i < tt.requests; i++ {
				rec = httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/ping", nil)
				req.RemoteAddr = "10.0.0.9:1234"
				r.ServeHTTP(rec, req)
			}
			assert.Equal(t, tt.wantLastCode, rec.Code)
		})
	}
}

func TestNewRedisRateLimitMiddlewareBadRate(t *testing.T) {
	tests := []struct {
		name string
		rate string
	}{
		{name: "garbage", rate: "zzz"},
		{name: "empty", rate: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// nil client is safe here: rate parsing fails before the client is touched.
			_, err := middleware.NewRedisRateLimitMiddleware(nil, "engflex:rl:test", tt.rate)
			require.Error(t, err)
		})
	}
}

func TestConversationsSessionLimiterShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		rate     string
		requests int
		wantLast int
	}{
		{name: "sona parity 5-M trips on 6th", rate: "5-M", requests: 6, wantLast: http.StatusTooManyRequests},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := middleware.NewInMemoryRateLimitMiddleware(tt.rate)
			require.NoError(t, err)
			r := gin.New()
			// Mirrors module.go order: limiter BEFORE auth (IP-keyed, sona parity).
			r.POST("/conversations", h, func(c *gin.Context) { c.Status(http.StatusCreated) })
			var rec *httptest.ResponseRecorder
			for i := 0; i < tt.requests; i++ {
				rec = httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/conversations", nil)
				req.RemoteAddr = "10.0.0.7:1234"
				r.ServeHTTP(rec, req)
			}
			assert.Equal(t, tt.wantLast, rec.Code)
		})
	}
}

// errStore fails every read: a Redis outage after boot surfaces as store
// errors (not as missing middleware), and requests must still be served.
type errStore struct{}

func (errStore) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, assert.AnError
}

func (errStore) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, assert.AnError
}

func (errStore) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, assert.AnError
}

func (errStore) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, assert.AnError
}

func TestApplyProxyTrust(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		trusted    []string
		remoteAddr string
		xff        string
		wantIP     string
		wantErr    bool
	}{
		{
			name:       "trust none ignores spoofed header",
			trusted:    nil,
			remoteAddr: "1.2.3.4:1234",
			xff:        "9.9.9.9",
			wantIP:     "1.2.3.4",
		},
		{
			name:       "trusted proxy honors header",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "10.1.2.3:1234",
			xff:        "9.9.9.9",
			wantIP:     "9.9.9.9",
		},
		{
			name:       "untrusted remote ignores header",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "1.2.3.4:1234",
			xff:        "9.9.9.9",
			wantIP:     "1.2.3.4",
		},
		{
			name:       "invalid cidr errors",
			trusted:    []string{"not-a-cidr"},
			remoteAddr: "1.2.3.4:1234",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			err := middleware.ApplyProxyTrust(r, tt.trusted)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var got string
			r.GET("/whoami", func(c *gin.Context) {
				got = c.ClientIP()
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.wantIP, got)
		})
	}
}

func TestRateLimitStoreErrorFailsOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		rate string
	}{
		{name: "store outage serves request", rate: "1-M"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := middleware.NewRateLimitMiddleware(errStore{}, tt.rate)
			require.NoError(t, err)
			r := gin.New()
			r.GET("/ping", h, func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = "10.0.0.9:1234"
			rec := httptest.NewRecorder()
			require.NotPanics(t, func() { r.ServeHTTP(rec, req) })
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestRateLimitExceededEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, err := middleware.NewInMemoryRateLimitMiddleware("1-M")
	require.NoError(t, err)
	r := gin.New()
	r.GET("/ping", h, func(c *gin.Context) { c.Status(http.StatusOK) })
	rec := httptest.NewRecorder()
	for i := 0; i < 2; i++ {
		rec = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.9:1234"
		r.ServeHTTP(rec, req)
	}
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "rate limit exceeded", body["message"])
}

func TestBypassRateLimitForPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		path         string
		requests     int
		wantLastCode int
	}{
		{name: "internal prefix bypasses limiter", path: "/api/v1/internal/job", requests: 3, wantLastCode: http.StatusOK},
		{name: "public path still limited", path: "/api/v1/scenarios", requests: 2, wantLastCode: http.StatusTooManyRequests},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim, err := middleware.NewInMemoryRateLimitMiddleware("1-M")
			require.NoError(t, err)
			r := gin.New()
			r.Use(middleware.BypassRateLimitForPrefix("/api/v1/internal/", lim))
			v1 := r.Group("/api/v1")
			v1.GET("/internal/job", func(c *gin.Context) { c.Status(http.StatusOK) })
			v1.GET("/scenarios", func(c *gin.Context) { c.Status(http.StatusOK) })
			var rec *httptest.ResponseRecorder
			for i := 0; i < tt.requests; i++ {
				rec = httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, tt.path, nil)
				req.RemoteAddr = "10.0.0.9:1234"
				r.ServeHTTP(rec, req)
			}
			assert.Equal(t, tt.wantLastCode, rec.Code)
		})
	}
}

func TestRateLimitKeyPrefersCFHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		remoteAddr string
		cfHeader   string
		xff        string
		wantKey    string
	}{
		{name: "cf header wins over client ip", remoteAddr: "10.0.1.9:1234", cfHeader: "203.0.113.7", xff: "198.51.100.3", wantKey: "203.0.113.7"},
		{name: "cf header trimmed", remoteAddr: "10.0.1.9:1234", cfHeader: "  203.0.113.7  ", wantKey: "203.0.113.7"},
		{name: "empty cf falls back to client ip", remoteAddr: "10.0.1.9:1234", cfHeader: "   ", xff: "198.51.100.3", wantKey: "198.51.100.3"},
		{name: "absent cf falls back to client ip", remoteAddr: "10.0.1.9:1234", xff: "198.51.100.3", wantKey: "198.51.100.3"},
		{name: "no headers falls back to remote addr", remoteAddr: "192.168.1.50:1234", wantKey: "192.168.1.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			require.NoError(t, middleware.ApplyProxyTrust(r, []string{"10.0.0.0/8"}))
			var got string
			r.GET("/whoami", func(c *gin.Context) {
				got = middleware.RateLimitKey(c)
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.cfHeader != "" {
				req.Header.Set("CF-Connecting-IP", tt.cfHeader)
			}
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.wantKey, got)
		})
	}
}

func TestRateLimitBucketsByCFHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	newBucket := func(t *testing.T) gin.HandlerFunc {
		h, err := middleware.NewInMemoryRateLimitMiddleware("2-M")
		require.NoError(t, err)
		return h
	}
	t.Run("same cf shares one bucket across remote addrs", func(t *testing.T) {
		r := gin.New()
		r.GET("/ping", newBucket(t), func(c *gin.Context) { c.Status(http.StatusOK) })
		var last int
		for i := 0; i < 3; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = "10.0.1.9:1234"
			req.Header.Set("CF-Connecting-IP", "203.0.113.7")
			r.ServeHTTP(rec, req)
			last = rec.Code
		}
		assert.Equal(t, http.StatusTooManyRequests, last)
	})
	t.Run("distinct cf values bucket separately", func(t *testing.T) {
		r := gin.New()
		r.GET("/ping", newBucket(t), func(c *gin.Context) { c.Status(http.StatusOK) })
		for i := 0; i < 2; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = "10.0.1.9:1234"
			req.Header.Set("CF-Connecting-IP", "203.0.113.7")
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.1.9:1234"
		req.Header.Set("CF-Connecting-IP", "203.0.113.99")
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
