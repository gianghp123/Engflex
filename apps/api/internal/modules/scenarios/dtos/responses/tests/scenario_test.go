package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/scenarios/dtos/responses"
)

func TestScenarioDecodeObjectives(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []responses.ScenarioObjective
	}{
		{
			name: "decodes the jsonb array",
			raw:  `[{"id":1,"label":"Greet and order","hint":"Say hello first"},{"id":2,"label":"Ask the price","hint":""}]`,
			want: []responses.ScenarioObjective{
				{ID: 1, Label: "Greet and order", Hint: "Say hello first"},
				{ID: 2, Label: "Ask the price"},
			},
		},
		{
			name: "empty column yields an empty non-nil list",
			raw:  "",
			want: []responses.ScenarioObjective{},
		},
		{
			name: "null yields an empty non-nil list",
			raw:  "null",
			want: []responses.ScenarioObjective{},
		},
		{
			name: "malformed blob degrades to empty",
			raw:  `[{"id":`,
			want: []responses.ScenarioObjective{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dto responses.Scenario
			dto.DecodeObjectives([]byte(tt.raw))
			require.NotNil(t, dto.Objectives)
			assert.Equal(t, tt.want, dto.Objectives)
		})
	}
}
