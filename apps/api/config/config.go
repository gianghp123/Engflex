package config

import (
	"engflex-api/internal/utils"
)

// Config aggregates every configuration section of the API.
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Cors      CorsConfig
	Log       LogConfig
	Voice     VoiceConfig
	LLM       LLMConfig
	Sentry    SentryConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	Media     MediaConfig
}

// Load reads .env files (if present) and then resolves every config section
// from the environment. Real environment variables always win over .env
// values. When no path is given, ".env" in the process working directory is
// tried (and silently skipped when absent); pass an explicit path in tests or
// when the binary runs from another directory (e.g. "apps/api/.env").
func Load(dotEnvPaths ...string) (Config, error) {
	if err := utils.LoadDotEnv(dotEnvPaths...); err != nil {
		return Config{}, err
	}

	server := LoadServerConfig()
	return Config{
		Server:    server,
		Database:  LoadDatabaseConfig(),
		Auth:      LoadAuthConfig(),
		Cors:      LoadCorsConfig(),
		Log:       LoadLogConfig(server.Env),
		Voice:     LoadVoiceConfig(),
		LLM:       LoadLLMConfig(),
		Sentry:    LoadSentryConfig(server.Env),
		Redis:     LoadRedisConfig(),
		RateLimit: LoadRateLimitConfig(),
		Media:     LoadMediaConfig(),
	}, nil
}
