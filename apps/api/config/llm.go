package config

import (
	"time"

	"engflex-api/internal/utils"
)

// LLMConfig is the OpenAI-compatible chat endpoint for text-only scoring.
// Role-named (never vendor-named): any gateway speaking the protocol works.
type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

// LoadLLMConfig reads the LLM endpoint. The key is required: fail fast.
func LoadLLMConfig() LLMConfig {
	model := utils.GetEnv("LLM_MODEL", "")
	if model == "" {
		// Voice stack names the same setting LLM_NAME; accept both.
		model = utils.GetEnv("LLM_NAME", "gpt-4o-mini")
	}
	return LLMConfig{
		BaseURL: utils.GetEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		APIKey:  utils.GetEnv("LLM_API_KEY", ""),
		Model:   model,
		Timeout: time.Duration(utils.GetEnvInt("LLM_TIMEOUT_MS", 60000)) * time.Millisecond,
	}
}
