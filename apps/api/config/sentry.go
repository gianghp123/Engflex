package config

import (
	"strconv"

	"engflex-api/internal/utils"
)

// SentryConfig holds the Sentry error-monitoring settings. An empty DSN
// disables Sentry entirely: dev, CI, and tests run without it and the
// server boots normally. Only SENTRY_DSN carries a secret.
type SentryConfig struct {
	// DSN is the Sentry project DSN. Empty means disabled.
	DSN string
	// Environment is reported on every event (defaults to APP_ENV).
	Environment string
	// TracesSampleRate is the fraction of requests traced (0.0-1.0).
	TracesSampleRate float64
	// Debug enables SDK debug messages (startup troubleshooting only).
	Debug bool
}

// LoadSentryConfig reads the Sentry settings from the environment. env is
// the runtime environment (see LoadServerConfig), used as the default event
// environment. A malformed or out-of-range sample rate fails closed to 0
// (error monitoring only, no tracing).
func LoadSentryConfig(env string) SentryConfig {
	rate, err := strconv.ParseFloat(utils.GetEnv("SENTRY_TRACES_SAMPLE_RATE", "0"), 64)
	if err != nil || rate < 0 || rate > 1 {
		rate = 0
	}
	return SentryConfig{
		DSN:              utils.GetEnv("SENTRY_DSN", ""),
		Environment:      utils.GetEnv("SENTRY_ENVIRONMENT", env),
		TracesSampleRate: rate,
		Debug:            utils.GetEnvBool("SENTRY_DEBUG", false),
	}
}
