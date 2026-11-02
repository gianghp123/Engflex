package tests

import (
	"context"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"os"
	"testing"
)

func TestPostgresProfileCheckedPreferences(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to migrated PostgreSQL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	defer pool.Close()
	profile := &models.UserProfile{UserID: "profile-" + uuid.NewString(), Preferences: models.Preferences{Level: enums.LevelIntermediate, Goals: []enums.Goal{enums.GoalWork}, Topics: []enums.InterestTopic{enums.InterestTopicBusiness, enums.InterestTopicTechnology}, DailyCommitmentMin: 15}}
	repo := repositories.NewUserProfileRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.Upsert(ctx, profile))
	defer db.Where("clerk_user_id = ?", profile.UserID).Delete(&models.UserProfile{})
	saved, err := repo.GetByUserID(ctx, profile.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, saved.ID)
	require.Equal(t, profile.Preferences, saved.Preferences)
	profile.Preferences.Level = enums.LevelAdvanced
	require.NoError(t, repo.Upsert(ctx, profile))
	saved, err = repo.GetByUserID(ctx, profile.UserID)
	require.NoError(t, err)
	require.Equal(t, enums.LevelAdvanced, saved.Preferences.Level)
	for column, value := range map[string]any{"level": "unknown", "goals": "{unknown}", "topics": "{unknown}", "daily_commitment_min": 17} {
		require.Error(t, db.Model(&models.UserProfile{}).Where("clerk_user_id = ?", profile.UserID).Update(column, value).Error)
	}
}
