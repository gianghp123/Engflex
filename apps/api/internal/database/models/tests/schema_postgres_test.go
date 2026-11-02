package tests

import (
	"engflex-api/internal/database/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"os"
	"sync"
	"testing"
)

func TestPostgresPersistedModelColumns(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	defer pool.Close()
	persisted := []any{&models.Attempt{}, &models.Bookmark{}, &models.Conversation{}, &models.ConversationTurn{}, &models.ExamCategory{}, &models.ExamTest{}, &models.Feedback{}, &models.Lesson{}, &models.LessonActivity{}, &models.LessonSection{}, &models.PaymentTransaction{}, &models.Persona{}, &models.Plan{}, &models.PremiumFeature{}, &models.Scenario{}, &models.ScenarioTopic{}, &models.Story{}, &models.StorySentence{}, &models.Subscription{}, &models.UserProfile{}, &models.UserVocabulary{}, &models.VideoCategory{}, &models.VideoExercise{}, &models.VideoTranscript{}, &models.VocabularyCategory{}, &models.VocabularyDeck{}, &models.VocabularyDeckItem{}, &models.VocabularyItem{}}
	for _, model := range persisted {
		parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		require.NoError(t, err)
		t.Run(parsed.Table, func(t *testing.T) {
			columns, err := db.Migrator().ColumnTypes(model)
			require.NoError(t, err)
			names := map[string]bool{}
			for _, column := range columns {
				names[column.Name()] = true
			}
			for _, field := range parsed.Fields {
				if field.DBName != "" && !field.IgnoreMigration {
					require.Contains(t, names, field.DBName, "persisted model field missing from Goose schema")
				}
			}
			for name := range names {
				require.Contains(t, parsed.FieldsByDBName, name, "persisted column missing from model")
			}
		})
	}
}
