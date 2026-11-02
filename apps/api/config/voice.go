package config

import (
	"time"

	"engflex-api/internal/utils"
)

// VoiceConfig holds the Python voice engine settings.
// VOICE_INTERNAL_SECRET is required so the server fails fast when unset.
type VoiceConfig struct {
	// ServiceURL is the engine base URL (upstream dev runner in dev).
	ServiceURL string
	// InternalSecret authenticates Go -> engine and engine -> Go callbacks.
	InternalSecret string
	// MaxDurationSec is the session time box sent to the engine.
	MaxDurationSec int
	// StartTimeout bounds the engine /start call.
	StartTimeout time.Duration
	// OfferTimeout bounds the engine offer/ICE calls.
	OfferTimeout time.Duration
	// AnalyzeTimeout bounds the engine /analyze call (one LLM pass).
	AnalyzeTimeout time.Duration
	// PronounceURL is the OpenPronounce container base URL (models baked into
	// the image, loaded lazily once per container start).
	PronounceURL string
	// PronounceTimeout bounds both container calls (/pronunciation then
	// /phonemes); it covers a cold container start.
	PronounceTimeout time.Duration
}

// LoadVoiceConfig reads the voice configuration from the environment.
func LoadVoiceConfig() VoiceConfig {
	return VoiceConfig{
		ServiceURL:       utils.GetEnv("VOICE_SERVICE_URL", "http://localhost:7860"),
		InternalSecret:   utils.GetEnv("VOICE_INTERNAL_SECRET", ""),
		MaxDurationSec:   utils.GetEnvInt("VOICE_MAX_DURATION_SEC", 300),
		StartTimeout:     time.Duration(utils.GetEnvInt("VOICE_START_TIMEOUT_MS", 15000)) * time.Millisecond,
		OfferTimeout:     time.Duration(utils.GetEnvInt("VOICE_OFFER_TIMEOUT_MS", 10000)) * time.Millisecond,
		AnalyzeTimeout:   time.Duration(utils.GetEnvInt("VOICE_ANALYZE_TIMEOUT_MS", 60000)) * time.Millisecond,
		PronounceURL:     utils.GetEnv("PRONOUNCE_SERVICE_URL", "http://localhost:8001"),
		PronounceTimeout: time.Duration(utils.GetEnvInt("PRONOUNCE_TIMEOUT_MS", 120000)) * time.Millisecond,
	}
}
