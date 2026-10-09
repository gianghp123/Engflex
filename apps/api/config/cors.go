package config

import (
	"strings"

	"engflex-api/internal/utils"
)

// CorsConfig holds the CORS settings.
// Values are read from the CORS_ORIGINS environment variable
// (comma-separated, e.g. "https://app.example,http://localhost:5173").
//
// NOTE (dev): the router restricts origins to CorsConfig.AllowedOrigins;
// keep CORS_ORIGINS covering every local dev/e2e web port.
type CorsConfig struct {
	// AllowedOrigins are the origins allowed to call the API.
	AllowedOrigins []string
}

// LoadCorsConfig reads the CORS configuration from the environment.
// Defaults cover local Vite development.
func LoadCorsConfig() CorsConfig {
	raw := utils.GetEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:5199")
	origins := make([]string, 0, 2)
	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return CorsConfig{AllowedOrigins: origins}
}
