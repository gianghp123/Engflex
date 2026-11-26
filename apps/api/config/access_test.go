package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAccessConfigDefaultsAndOverrides(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "")
	t.Setenv("CONTENT_PREMIUM_GATING", "")
	t.Setenv("APP_TIMEZONE", "")
	assert.False(t, LoadAccessConfig().Enforce, "enforcement must ship off (shadow mode)")
	assert.False(t, LoadAccessConfig().ContentGating, "Option A: content gating ships off")
	assert.Equal(t, "Asia/Ho_Chi_Minh", LoadAccessConfig().Timezone)

	t.Setenv("ACCESS_ENFORCE", "true")
	t.Setenv("CONTENT_PREMIUM_GATING", "1")
	t.Setenv("APP_TIMEZONE", "Asia/Bangkok")
	cfg := LoadAccessConfig()
	assert.True(t, cfg.Enforce)
	assert.True(t, cfg.ContentGating)
	assert.Equal(t, "Asia/Bangkok", cfg.Timezone)

	t.Setenv("ACCESS_ENFORCE", "garbage")
	assert.False(t, LoadAccessConfig().Enforce, "unparseable falls back, never fails boot")
}

func TestAccessConfigLocationFallsBackToUTC(t *testing.T) {
	loc := AccessConfig{Timezone: "Not/AZone"}.Location()
	require.NotNil(t, loc)
	assert.Equal(t, time.UTC, loc)
	assert.Equal(t, "Asia/Ho_Chi_Minh", AccessConfig{}.Location().String())
}
