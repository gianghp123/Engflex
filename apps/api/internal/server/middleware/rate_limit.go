package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	mstore "github.com/ulule/limiter/v3/drivers/store/memory"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"

	"engflex-api/internal/common"
)

// ApplyProxyTrust configures which upstream proxies the engine believes for
// X-Forwarded-For / X-Real-IP parsing (gin's ClientIP, which the limiter
// keys on). Empty cidrs trusts none: spoofed headers are ignored and
// buckets key on RemoteAddr. Gin defaults to trust-all, which lets any
// client rotate X-Forwarded-For and dodge the limiter, so the router calls
// this explicitly with RATE_LIMIT_TRUSTED_PROXIES.
func ApplyProxyTrust(r *gin.Engine, cidrs []string) error {
	if len(cidrs) == 0 {
		return r.SetTrustedProxies(nil)
	}
	return r.SetTrustedProxies(cidrs)
}

// failOpenHandler keeps serving requests when the store errors (e.g. Redis
// flaps after boot). It calls c.Next() so the request continues down the
// chain; mgin aborts after OnError returns, which only stops later
// middleware — the handlers already run via Next() complete normally.
func failOpenHandler(c *gin.Context, err error) {
	slog.WarnContext(c.Request.Context(), "rate limit store error, allowing request", "error", err)
	c.Next()
}

// BypassRateLimitForPrefix skips h for routes under prefix (matched on the
// full route path, e.g. "/api/v1/internal/") and applies it elsewhere.
// Machine-to-machine callbacks (engine finalize/ingest) share the v1 group
// with user traffic; without this a busy engine could 429 its own
// delivery calls and lose turn data.
func BypassRateLimitForPrefix(prefix string, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.FullPath(), prefix) {
			c.Next()
			return
		}
		h(c)
	}
}

// limitReachedHandler answers 429s in the house envelope so API clients see
// {message} like every other error. The library default is plain-text "Limit
// exceeded", which fetch clients cannot parse into ApiError and which crashes
// error boundaries with a SyntaxError instead of a readable status.
func limitReachedHandler(c *gin.Context) {
	common.Fail(c, common.TooManyRequests("rate limit exceeded"))
}

// RateLimitKey keys the limiter on CF-Connecting-IP when present (tunnel-
// authoritative: the Cloudflare edge overwrites it, so tunnel clients cannot
// spoof each other's buckets), else gin ClientIP (LAN direct + M2M, parsed
// under the trusted-proxy rules). Trimmed; empty falls back. Pre-lockdown
// caveat: LAN-local clients can set the header (trusted network); after the
// planned inbound lockdown (tunnel-only traffic) every value is edge-set.
func RateLimitKey(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); v != "" {
		return v
	}
	return c.ClientIP()
}

// NewRateLimitMiddleware builds an IP-keyed Gin rate limiter over any
// limiter.Store. formattedRate is "<limit>-<period>" (S/M/H/D, e.g. "60-M").
// Unlike sona-voice (which panics), invalid input returns an error so the
// caller decides how to degrade (see router wiring: fail-open with a log).
func NewRateLimitMiddleware(store limiter.Store, formattedRate string) (gin.HandlerFunc, error) {
	rate, err := limiter.NewRateFromFormatted(formattedRate)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit %q: %w", formattedRate, err)
	}
	return mgin.NewMiddleware(
		limiter.New(store, rate),
		mgin.WithKeyGetter(RateLimitKey),
		mgin.WithErrorHandler(failOpenHandler),
		mgin.WithLimitReachedHandler(limitReachedHandler),
	), nil
}

// NewInMemoryRateLimitMiddleware builds a single-replica limiter (dev/CI,
// no Redis required).
func NewInMemoryRateLimitMiddleware(formattedRate string) (gin.HandlerFunc, error) {
	return NewRateLimitMiddleware(mstore.NewStore(), formattedRate)
}

// NewRedisRateLimitMiddleware builds a shared limiter for multi-replica
// prod. prefix namespaces keys per environment (e.g. "engflex:rl:global").
func NewRedisRateLimitMiddleware(redisClient *redis.Client, prefix, formattedRate string) (gin.HandlerFunc, error) {
	rate, err := limiter.NewRateFromFormatted(formattedRate)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit %q: %w", formattedRate, err)
	}
	store, err := sredis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   prefix,
		MaxRetry: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter store %q: %w", prefix, err)
	}
	return mgin.NewMiddleware(
		limiter.New(store, rate),
		mgin.WithKeyGetter(RateLimitKey),
		mgin.WithErrorHandler(failOpenHandler),
	), nil
}
