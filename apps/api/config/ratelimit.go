package config

import "engflex-api/internal/utils"

// RateLimitConfig holds request throttling settings.
// Rates use ulule/limiter formatted strings ("<limit>-<period>",
// S/M/H/D, e.g. "300-M"). A rate is opt-in: unset or empty means no limiter
// for that scope. There is deliberately no hidden default, so an undeclared
// budget can never silently throttle traffic (the failure mode that shipped
// 60-M/5-M to prod). The Redis store is selected via RedisConfig; an empty
// Redis URL keeps the in-memory store (single-replica dev/CI).
// TrustedProxies is a comma-separated list of CIDRs whose X-Forwarded-For
// the limiter trusts; empty trusts none (spoofed headers ignored, buckets
// key on RemoteAddr). Set the platform LB CIDR in prod.
type RateLimitConfig struct {
	GlobalRate     string
	SessionRate    string
	TrustedProxies string
}

// LoadRateLimitConfig reads rate limiting settings from the environment.
// Unset or empty rates declare no limiter for that scope.
func LoadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRate:     utils.GetEnv("RATE_LIMIT_GLOBAL", ""),
		SessionRate:    utils.GetEnv("RATE_LIMIT_SESSION", ""),
		TrustedProxies: utils.GetEnv("RATE_LIMIT_TRUSTED_PROXIES", ""),
	}
}
