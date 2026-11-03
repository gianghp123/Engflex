// Package llm is the Go-side OpenAI-compatible chat client for text-only
// scoring. Strategy is strict json_schema first with a tolerant
// extract-and-repair fallback: proxies may accept response_format and still
// answer in markdown (observed on the voice path).
package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"engflex-api/config"
	"github.com/kaptinlin/jsonrepair"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// LLMClient talks to an OpenAI-compatible chat endpoint.
type LLMClient interface {
	CompleteText(ctx context.Context, prompt string) (string, error)
	CompleteStructured(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) error
}

type httpLLMClient struct {
	sdk     *openai.Client
	model   string
	timeout time.Duration
}

// NewHTTPLLMClient builds the client over the shared SDK handle.
func NewHTTPLLMClient(cfg config.LLMConfig) LLMClient {
	sdk := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)
	return &httpLLMClient{sdk: &sdk, model: cfg.Model, timeout: cfg.Timeout}
}

func (c *httpLLMClient) CompleteText(ctx context.Context, prompt string) (string, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.sdk.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(c.model),
		Messages:    []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
		Temperature: openai.Float(0.2),
	})
	status := "ok"
	if err != nil {
		status = "error"
		slog.InfoContext(ctx, "llm call",
			"model", c.model, "op", "complete_text",
			"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
			"latency_ms", time.Since(start).Milliseconds(), "status", status)
		return "", err
	}
	if len(resp.Choices) == 0 {
		slog.InfoContext(ctx, "llm call", "model", c.model, "op", "complete_text",
			"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
			"latency_ms", time.Since(start).Milliseconds(), "status", "empty_choices")
		return "", fmt.Errorf("no completion choices returned")
	}
	slog.InfoContext(ctx, "llm call",
		"model", c.model, "op", "complete_text",
		"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
		"prompt_tokens", resp.Usage.PromptTokens, "completion_tokens", resp.Usage.CompletionTokens,
		"latency_ms", time.Since(start).Milliseconds(), "status", status)
	return resp.Choices[0].Message.Content, nil
}

func (c *httpLLMClient) CompleteStructured(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.sdk.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(c.model),
		Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   schemaName,
					Schema: schema,
					Strict: openai.Bool(true),
				},
			},
		},
		Temperature: openai.Float(0.2),
	})
	if err != nil {
		slog.InfoContext(ctx, "llm call",
			"model", c.model, "op", "complete_structured", "schema", schemaName,
			"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
			"latency_ms", time.Since(start).Milliseconds(), "status", "error")
		return err
	}
	if len(resp.Choices) == 0 {
		slog.InfoContext(ctx, "llm call",
			"model", c.model, "op", "complete_structured", "schema", schemaName,
			"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
			"latency_ms", time.Since(start).Milliseconds(), "status", "empty_choices")
		return fmt.Errorf("no completion choices returned")
	}
	content := resp.Choices[0].Message.Content
	logArgs := []any{
		"model", c.model, "op", "complete_structured", "schema", schemaName,
		"prompt_hash", promptHash(prompt), "prompt_len", len(prompt),
		"prompt_tokens", resp.Usage.PromptTokens, "completion_tokens", resp.Usage.CompletionTokens,
		"latency_ms", time.Since(start).Milliseconds(),
	}
	if err := json.Unmarshal([]byte(content), output); err == nil {
		slog.InfoContext(ctx, "llm call", append(logArgs, "status", "ok")...)
		return nil
	}
	repaired, rerr := jsonrepair.Repair(extractJSONObject(content))
	if rerr != nil {
		slog.InfoContext(ctx, "llm call", append(logArgs, "status", "unusable_json")...)
		return fmt.Errorf("llm reply held no usable JSON: %w", rerr)
	}
	if err := json.Unmarshal([]byte(repaired), output); err != nil {
		slog.InfoContext(ctx, "llm call", append(logArgs, "status", "invalid_json")...)
		return fmt.Errorf("llm reply failed validation: %w", err)
	}
	slog.InfoContext(ctx, "llm call", append(logArgs, "status", "repaired")...)
	return nil
}

// promptHash is a 16-hex-char prefix of the prompt's sha256. Logs carry the
// hash and length for cost/latency correlation, never the raw prompt text.
func promptHash(p string) string {
	h := sha256.Sum256([]byte(p))
	return hex.EncodeToString(h[:])[:16]
}

// extractJSONObject returns the first balanced {...} block, else "".
func extractJSONObject(raw string) string {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '{' {
			continue
		}
		depth, inStr, esc := 0, false, false
		for j := i; j < len(raw); j++ {
			ch := raw[j]
			if inStr {
				if esc {
					esc = false
				} else if ch == '\\' {
					esc = true
				} else if ch == '"' {
					inStr = false
				}
				continue
			}
			switch ch {
			case '"':
				inStr = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return raw[i : j+1]
				}
			}
		}
	}
	return ""
}
