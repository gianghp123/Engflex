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

func TestVideoDetailPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })

	category := &models.VideoCategory{Slug: uuid.NewString(), Name: "Video detail fixture"}
	require.NoError(t, tx.Create(category).Error)
	video := &models.VideoExercise{CategoryID: &category.ID, Title: "Fixture", VideoURL: "fixture"}
	require.NoError(t, tx.Create(video).Error)
	require.NoError(t, tx.Create(&models.VideoTranscript{VideoExerciseID: video.ID, Sequence: 1, Content: "Caption"}).Error)

	detail, err := repositories.NewVideoExerciseRepository(tx).GetDetail(context.Background(), video.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Category)
	require.Equal(t, category.ID, detail.Category.ID)
	require.Len(t, detail.Transcripts, 1)
	require.Equal(t, "Caption", detail.Transcripts[0].Content)
}
