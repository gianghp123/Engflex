package config

import (
	"time"

	"engflex-api/internal/utils"
)

// AccessConfig holds entitlement enforcement switches. Both booleans ship
// OFF at launch (spec §15): content gating is disabled under Option A, and
// paid-AI enforcement starts in shadow mode (Decide runs and logs, nothing
// is reserved).
type AccessConfig struct {
	// Enforce gates paid AI actions on the access service (ACCESS_ENFORCE).
	// False = shadow mode: decisions are evaluated and logged, no hold.
	Enforce bool
	// ContentGating restores per-item is_premium checks
	// (CONTENT_PREMIUM_GATING). Option A keeps static content free.
	ContentGating bool
	// Timezone is the location quota periods are cut in (APP_TIMEZONE).
	// Stored timestamps stay UTC; only period keys and resetAt use it.
	Timezone string
}

// LoadAccessConfig reads the access configuration from the environment.
func LoadAccessConfig() AccessConfig {
	return AccessConfig{
		Enforce:       utils.GetEnvBool("ACCESS_ENFORCE", false),
		ContentGating: utils.GetEnvBool("CONTENT_PREMIUM_GATING", false),
		Timezone:      utils.GetEnv("APP_TIMEZONE", "Asia/Ho_Chi_Minh"),
	}
}

// Location parses Timezone, falling back to UTC when the configured value
// is invalid. The caller logs the fallback once at boot (spec §7.4).
func (c AccessConfig) Location() *time.Location {
	tz := c.Timezone
	if tz == "" {
		tz = "Asia/Ho_Chi_Minh"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}
