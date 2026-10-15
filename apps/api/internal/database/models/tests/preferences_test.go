package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

func TestPreferencesScan(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantLevel enums.Level
		wantGoals []enums.Goal
		wantMin   int
		wantTopic enums.InterestTopic
		wantEmpty bool
	}{
		{
			name:      "camelCase blob decodes every field",
			value:     []byte(`{"level":"intermediate","goals":["work","exams"],"dailyCommitmentMin":20,"topics":["business","technology"]}`),
			wantLevel: enums.LevelIntermediate,
			wantGoals: []enums.Goal{enums.GoalWork, enums.GoalExams},
			wantMin:   20,
			wantTopic: enums.InterestTopicBusiness,
		},
		{
			name:      "empty object leaves zero values",
			value:     []byte(`{}`),
			wantEmpty: true,
		},
		{
			name:      "NULL leaves the zero value",
			value:     nil,
			wantEmpty: true,
		},
		{
			name:      "malformed JSON leaves the zero value",
			value:     []byte("{not json"),
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got models.Preferences
			require.NoError(t, got.Scan(tt.value))

			if tt.wantEmpty {
				assert.Equal(t, models.Preferences{}, got)
				return
			}
			assert.Equal(t, tt.wantLevel, got.Level)
			assert.Equal(t, tt.wantGoals, got.Goals)
			assert.Equal(t, tt.wantMin, got.DailyCommitmentMin)
			require.Len(t, got.Topics, 2)
			assert.Equal(t, tt.wantTopic, got.Topics[0])
		})
	}
}

func TestPreferencesValue(t *testing.T) {
	in := models.Preferences{
		Level:              enums.LevelIntermediate,
		Goals:              []enums.Goal{enums.GoalWork, enums.GoalExams},
		DailyCommitmentMin: 20,
		Topics:             []enums.InterestTopic{enums.InterestTopicBusiness, enums.InterestTopicTechnology},
	}

	raw, err := in.Value()
	require.NoError(t, err)
	encoded, ok := raw.([]byte)
	require.True(t, ok)

	assert.JSONEq(t, `{"level":"intermediate","goals":["work","exams"],"dailyCommitmentMin":20,"topics":["business","technology"]}`, string(encoded))

	var back models.Preferences
	require.NoError(t, back.Scan(encoded))
	assert.Equal(t, in, back)
}
