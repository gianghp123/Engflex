package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/lessons/activityconfig"
	"engflex-api/internal/modules/lessons/prompts"
)

func seededWritingConfig() activityconfig.WritingConfig {
	return activityconfig.WritingConfig{
		Task:         "email",
		Instructions: "Write a short self-introduction email to your new team.",
		Stimulus:     "You joined the Osaka office this month.",
		MinWords:     40,
		MaxWords:     60,
		ModelAnswer:  "Hello everyone, my name is Kenji Sato and I am a sales assistant in the Osaka office.",
		Checklist:    []string{"You greeted the team", "You gave your name and role"},
	}
}

func TestBuildWritingScorePrompt(t *testing.T) {
	cfg := seededWritingConfig()
	text := "Hello everyone, my name is Kenji Sato and I is a sales assistant."
	got := prompts.BuildWritingScorePrompt(cfg, text, 55)

	tests := []struct {
		name   string
		substr string
	}{
		// Dynamic brief fields.
		{name: "contains task", substr: "email"},
		{name: "contains instructions", substr: "Write a short self-introduction email to your new team."},
		{name: "contains stimulus", substr: "You joined the Osaka office this month."},
		{name: "contains word range", substr: "40-60"},
		{name: "contains word count", substr: "55 words"},
		{name: "contains checklist item 1", substr: "You greeted the team"},
		{name: "contains checklist item 2", substr: "You gave your name and role"},
		{name: "labels the reference", substr: "Reference answer"},
		{name: "contains model answer", substr: "Hello everyone, my name is Kenji Sato and I am a sales assistant in the Osaka office."},
		{name: "contains learner text", substr: "Hello everyone, my name is Kenji Sato and I is a sales assistant."},
		// Six rule markers, one per section.
		{name: "rule task verdicts", substr: "task.verdicts"},
		{name: "rule grammar no invention", substr: "never invent errors"},
		{name: "rule phrasing preference", substr: "Never label a stylistic preference"},
		{name: "rule expression slots", substr: "... slots"},
		{name: "rule suggested not a copy", substr: "never a copy of the reference answer"},
		{name: "rule tip no repeat", substr: "do not repeat a span reason"},
		// Exact six-section key list.
		{name: "exact key list", substr: "task, grammar, phrasing, expressions, suggested, tip"},
		// Example markers: trio present, perfect case empty, anti-example.
		{name: "examples marker", substr: "Examples:"},
		{name: "scored example span", substr: `"occurrence": 1`},
		{name: "scored example answers prompt", substr: `"answersPrompt"`},
		{name: "perfect example empty grammar", substr: `"grammar": []`},
		{name: "perfect example empty phrasing", substr: `"phrasing": []`},
		{name: "perfect example tip", substr: "Clean email covering every point."},
		{name: "anti-example marker", substr: "NOT ("},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, got, tt.substr)
		})
	}
	require.Contains(t, got, text)
	assert.NotContains(t, got, `"corrected"`)
}
