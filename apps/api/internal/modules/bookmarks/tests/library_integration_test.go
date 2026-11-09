package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/bookmarks/dtos/responses"
	"engflex-api/internal/utils"
)

func TestSavedLibraryPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })

	ctx := context.Background()
	userID, otherUser := uuid.NewString(), uuid.NewString()
	story := &models.Story{Slug: uuid.NewString(), Title: "Bookmark story", AudioURL: "fixture"}
	require.NoError(t, tx.Create(story).Error)
	word := &models.VocabularyItem{Term: uuid.NewString(), CEFR: enums.CEFR("B1"), Status: enums.PublishStatusPublished, IsPremium: true}
	require.NoError(t, tx.Select("Term", "CEFR", "Status", "IsPremium").Create(word).Error)
	created := time.Now().UTC().Truncate(time.Second)
	due := created.Add(time.Hour)
	vocab := repositories.NewUserVocabularyRepository(tx)
	require.NoError(t, vocab.Upsert(ctx, &models.UserVocabulary{
		UserID: userID, ItemID: word.ID, SourceType: "manual", Mastered: true, SRSDueAt: &due, CreatedAt: created,
	}))
	bookmarks := repositories.NewBookmarkRepository(tx)
	storyBookmark := &models.Bookmark{UserID: userID, ContentID: story.ID, ContentType: enums.ContentTypeStory, CreatedAt: created}
	require.NoError(t, bookmarks.Create(ctx, storyBookmark))
	require.NoError(t, bookmarks.Create(ctx, &models.Bookmark{UserID: otherUser, ContentID: story.ID, ContentType: enums.ContentTypeStory}))
	require.NoError(t, bookmarks.Create(ctx, &models.Bookmark{UserID: userID, ContentID: uuid.NewString(), ContentType: enums.ContentTypeVideo, CreatedAt: created}))

	note := "Personal reminder"
	require.NoError(t, vocab.UpdateNote(ctx, userID, word.ID, &note))
	require.ErrorIs(t, vocab.UpdateNote(ctx, otherUser, word.ID, &note), gorm.ErrRecordNotFound)
	state, err := vocab.Get(ctx, userID, word.ID)
	require.NoError(t, err)
	require.True(t, state.Mastered)
	require.True(t, due.Equal(*state.SRSDueAt))
	require.NoError(t, bookmarks.UpdateNote(ctx, userID, "story", story.ID, &note))
	require.NoError(t, bookmarks.Create(ctx, storyBookmark))
	reloaded, err := bookmarks.Get(ctx, userID, "story", story.ID)
	require.NoError(t, err)
	require.Equal(t, &note, reloaded.Note)
	require.True(t, created.Equal(reloaded.CreatedAt))

	library := repositories.NewSavedItemRepository(tx)
	target, err := library.GetTarget(ctx, "story", story.ID)
	require.NoError(t, err)
	require.NotNil(t, target.Story)
	stats, err := library.LibraryStats(ctx, userID)
	require.NoError(t, err)
	require.EqualValues(t, 3, stats.Total)
	require.EqualValues(t, 2, stats.WithNotes)

	tests := []struct {
		name, kind, search string
		premium            bool
		offset             int
		total              int64
		wantKind           string
		visible            bool
	}{
		{name: "stable first page across both sources", total: 3, wantKind: "story", visible: true},
		{name: "missing video remains removable", offset: 1, total: 3, wantKind: "video"},
		{name: "premium title excluded from search", search: word.Term},
		{name: "premium word visible with access", search: word.Term, premium: true, total: 1, wantKind: "vocab", visible: true},
		{name: "private note visible without premium", kind: "vocab", search: "REMINDER", total: 1, wantKind: "vocab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, err := library.CountLibrary(ctx, userID, tt.kind, tt.search, false, tt.premium)
			require.NoError(t, err)
			require.Equal(t, tt.total, total)
			items, err := library.ListLibrary(ctx, userID, tt.kind, tt.search, false, tt.premium, 1, tt.offset)
			require.NoError(t, err)
			if tt.wantKind == "" {
				require.Empty(t, items)
				return
			}
			require.Len(t, items, 1)
			item := items[0]
			require.Equal(t, tt.wantKind, item.ContentType)
			require.Equal(t, tt.visible, item.Story != nil || item.VideoExercise != nil || item.VocabularyItem != nil)
			var dto responses.SavedLibrary
			require.NoError(t, utils.Map(&dto, &models.SavedLibrary{Items: items}))
			require.Len(t, dto.Items, 1)
			require.Equal(t, item.ContentID, dto.Items[0].ContentID)
			require.Equal(t, tt.visible, dto.Items[0].Story != nil || dto.Items[0].VideoExercise != nil || dto.Items[0].VocabularyItem != nil)
		})
	}
}
