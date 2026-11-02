package tests

import (
	"context"
	"os"
	"testing"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresVideoDetailCategoryJoin(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	for _, hasCategory := range []bool{false, true} {
		t.Run(map[bool]string{false: "nil relation", true: "joined category"}[hasCategory], func(t *testing.T) {
			video := &models.VideoExercise{Title: "Join regression", VideoURL: "test"}
			if hasCategory {
				category := &models.VideoCategory{Slug: uuid.NewString(), Name: "Join regression"}
				require.NoError(t, db.Create(category).Error)
				t.Cleanup(func() { require.NoError(t, db.Delete(category).Error) })
				video.CategoryID = &category.ID
			}
			require.NoError(t, db.Create(video).Error)
			t.Cleanup(func() { require.NoError(t, db.Delete(video).Error) })
			loaded, err := repositories.NewVideoExerciseRepository(db).GetDetail(context.Background(), video.ID)
			require.NoError(t, err)
			require.Equal(t, video.ID, loaded.ID)
			require.Empty(t, loaded.Transcripts)
			if hasCategory {
				require.NotNil(t, loaded.Category)
				require.Equal(t, *video.CategoryID, loaded.Category.ID)
			} else {
				require.Nil(t, loaded.Category)
			}
		})
	}
}
