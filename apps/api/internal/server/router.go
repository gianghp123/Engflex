package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"engflex-api/config"
	_ "engflex-api/docs"
	"engflex-api/internal/modules/conversations"
	"engflex-api/internal/modules/lessons"
	"engflex-api/internal/modules/scenarios"
	"engflex-api/internal/modules/videos"
	"engflex-api/internal/modules/vocabulary"
	"engflex-api/internal/server/middleware"
)

// NewRouter builds the Gin engine with shared middleware and module routes.
// Add domain modules here as they are implemented:
// profiles.RegisterRoutes(v1, db), contents.RegisterRoutes(v1, db), ...
func NewRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	r := gin.New()
	// Recovery mounts outermost (gin.Default semantics): sentrygin captures
	// the panic first and repanics into Recovery, which writes the
	// {message} envelope. Mounted the other way round, Recovery would
	// swallow the panic before sentrygin sees it and nothing is reported.
	r.Use(
		middleware.Recovery(),
		sentrygin.New(sentrygin.Options{Repanic: true, WaitForDelivery: false}),
		middleware.RequestID(), middleware.RequestLogger(),
	)
	// Trust X-Forwarded-For only from configured proxies so the IP-keyed
	// limiter buckets by real client IP (sona parity). Empty
	// RATE_LIMIT_TRUSTED_PROXIES trusts none: spoofed headers are ignored.
	// Gin defaults to trust-all, which lets any client rotate the header
	// and dodge the limiter — hence the explicit call.
	r.ForwardedByClientIP = true
	proxies := []string{}
	for _, p := range strings.Split(cfg.RateLimit.TrustedProxies, ",") {
		if p = strings.TrimSpace(p); p != "" {
			proxies = append(proxies, p)
		}
	}
	if err := middleware.ApplyProxyTrust(r, proxies); err != nil {
		slog.Error("invalid trusted proxies, trusting none", "value", cfg.RateLimit.TrustedProxies, "error", err)
		if terr := middleware.ApplyProxyTrust(r, nil); terr != nil {
			slog.Error("failed to reset proxy trust", "error", terr)
		}
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Cors.AllowedOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// Kept from origin/dev so existing probes keep working.
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// gin-swagger gates every request on a regexp that only accepts an
	// allowlist of asset names (index.html, doc.json, the swagger-ui bundles),
	// so a bare /swagger or /swagger/ is refused from *inside* the handler —
	// gin's own trailing-slash redirect lands straight on that 404. Route both
	// bare forms to the UI entry point so the URL people type works.
	swaggerIndex := func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	}
	// PersistAuthorization keeps whatever token is pasted into the Authorize
	// dialog in localStorage, so the long-lived SWAGGER_TESTING_JWT (or any
	// freshly minted session token) is entered once instead of on every
	// reload. The doc JSON is unaffected — this is UI state only.
	swaggerAssets := ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.PersistAuthorization(true))
	r.GET("/swagger", swaggerIndex)
	r.GET("/swagger/*any", func(c *gin.Context) {
		// The catch-all also matches the bare directory, so screen it here:
		// gin cannot route /swagger/ separately (it panics on the conflict
		// with *any), and c.Param("any") is "/" rather than "" for it.
		if c.Request.URL.Path == "/swagger/" {
			swaggerIndex(c)
			return
		}
		swaggerAssets(c)
	})

	v1 := r.Group("/api/v1")
	// buildLimiter selects the Redis store when REDIS_URL is set (shared
	// buckets across replicas) and falls back to the in-memory store
	// otherwise. Any failure degrades loudly but never blocks boot.
	buildLimiter := func(kind, rate, prefix string) gin.HandlerFunc {
		if !cfg.RateLimit.Enabled {
			return nil
		}
		if cfg.Redis.URL == "" {
			h, err := middleware.NewInMemoryRateLimitMiddleware(rate)
			if err != nil {
				slog.Error("rate limiting disabled: invalid rate", "kind", kind, "rate", rate, "error", err)
				return nil
			}
			return h
		}
		memoryFallback := func(reason string, err error) gin.HandlerFunc {
			slog.Warn("rate limiting degraded to memory store", "kind", kind, "reason", reason, "error", err)
			h, rerr := middleware.NewInMemoryRateLimitMiddleware(rate)
			if rerr != nil {
				slog.Error("rate limiting disabled: invalid rate", "kind", kind, "rate", rate, "error", rerr)
				return nil
			}
			return h
		}
		opt, err := redis.ParseURL(cfg.Redis.URL)
		if err != nil {
			return memoryFallback("bad REDIS_URL", err)
		}
		client := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			return memoryFallback("redis unreachable", err)
		}
		h, err := middleware.NewRedisRateLimitMiddleware(
			client, cfg.Redis.KeyPrefix+":rl:"+prefix, rate)
		if err != nil {
			slog.Error("rate limiting disabled", "kind", kind, "error", err)
			return nil
		}
		return h
	}
	if global := buildLimiter("global", cfg.RateLimit.GlobalRate, "global"); global != nil {
		// Machine-to-machine engine callbacks must never 429: they share
		// the v1 group (and its per-IP bucket) with user traffic.
		v1.Use(middleware.BypassRateLimitForPrefix("/api/v1/internal/", global))
	}
	sessionLimiter := buildLimiter("session", cfg.RateLimit.SessionRate, "session")
	conversations.RegisterRoutes(v1, db, cfg.Voice, sessionLimiter)
	lessons.RegisterRoutes(v1, db, cfg.Voice, cfg.LLM, cfg.Media, sessionLimiter)
	scenarios.RegisterRoutes(v1, db)
	videos.RegisterRoutes(v1, db)
	vocabulary.RegisterRoutes(v1, db)

	return r
}
