package tests

import (
	"encoding/json"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/profiles/dtos/responses"
	"engflex-api/internal/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
	"sync"
	"testing"
)

func TestPreferencesEmbeddedColumnsAndNestedWireShape(t *testing.T) {
	parsed, err := schema.Parse(&models.UserProfile{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	require.Contains(t, parsed.FieldsByDBName, "level")
	require.Contains(t, parsed.FieldsByDBName, "goals")
	require.Contains(t, parsed.FieldsByDBName, "topics")
	require.NotContains(t, parsed.FieldsByDBName, "preferences")
	require.True(t, parsed.FieldsByDBName["clerk_user_id"].PrimaryKey)
	var model models.UserProfile
	payload := `{"userId":"user-test","preferences":{"level":"intermediate","goals":["work","exams"],"dailyCommitmentMin":15,"topics":["business","technology"]}}`
	require.NoError(t, json.Unmarshal([]byte(payload), &model))
	var dto responses.Profile
	require.NoError(t, utils.Map(&dto, &model))
	require.Equal(t, string(model.Preferences.Level), string(dto.Preferences.Level))
	require.Len(t, dto.Preferences.Goals, 2)
	require.Len(t, dto.Preferences.Topics, 2)
}
