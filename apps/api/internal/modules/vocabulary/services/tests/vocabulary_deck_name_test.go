package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestVocabularyDeckName pins the name contract: trimmed on write, blank is a
// 400, and the partial unique index (user_id, lower(name)) surfaces as a 409
// through the existing FromDBError mapping.
func TestVocabularyDeckName(t *testing.T) {
	duplicate := &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}

	tests := []struct {
		name string
		run  func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository)
	}{
		{
			name: "create trims the name before writing",
			run: func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository) {
				repo.On("CountByUser", mock.Anything, "u1").Return(int64(0), nil).Once()
				repo.On("Create", mock.Anything, mock.MatchedBy(func(d *models.VocabularyDeck) bool {
					return d.Name == "Focus words" && d.UserID != nil && *d.UserID == "u1"
				})).Return(nil).Once()
				svc := services.NewVocabularyDeckService(repo, decks)
				deck, err := svc.Create(context.Background(), "u1", requests.CreateVocabularyDeck{Name: "  Focus words  "})
				require.NoError(t, err)
				require.Equal(t, "Focus words", deck.Name)
			},
		},
		{
			name: "create rejects a whitespace-only name",
			run: func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository) {
				svc := services.NewVocabularyDeckService(repo, decks)
				_, err := svc.Create(context.Background(), "u1", requests.CreateVocabularyDeck{Name: "   "})
				requireAppError(t, err, http.StatusBadRequest)
			},
		},
		{
			name: "create maps a duplicate name to 409",
			run: func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository) {
				repo.On("CountByUser", mock.Anything, "u1").Return(int64(0), nil).Once()
				repo.On("Create", mock.Anything, mock.Anything).Return(duplicate).Once()
				svc := services.NewVocabularyDeckService(repo, decks)
				_, err := svc.Create(context.Background(), "u1", requests.CreateVocabularyDeck{Name: "Focus words"})
				requireAppError(t, err, http.StatusConflict)
			},
		},
		{
			name: "rename trims and maps a duplicate to 409",
			run: func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository) {
				owner := "u1"
				repo.On("GetByID", mock.Anything, "d1").Return(&models.VocabularyDeck{ID: "d1", UserID: &owner, Name: "Old"}, nil).Once()
				repo.On("Update", mock.Anything, mock.MatchedBy(func(d *models.VocabularyDeck) bool {
					return d.Name == "Focus words"
				})).Return(duplicate).Once()
				svc := services.NewVocabularyDeckService(repo, decks)
				name := "  Focus words  "
				_, err := svc.Update(context.Background(), "u1", "d1", requests.UpdateVocabularyDeck{Name: &name})
				requireAppError(t, err, http.StatusConflict)
			},
		},
		{
			name: "rename rejects a whitespace-only name",
			run: func(t *testing.T, repo *mocks.MockVocabularyDeckRepository, decks *mocks.MockVocabularyDeckItemRepository) {
				owner := "u1"
				repo.On("GetByID", mock.Anything, "d1").Return(&models.VocabularyDeck{ID: "d1", UserID: &owner, Name: "Old"}, nil).Once()
				svc := services.NewVocabularyDeckService(repo, decks)
				name := "   "
				_, err := svc.Update(context.Background(), "u1", "d1", requests.UpdateVocabularyDeck{Name: &name})
				requireAppError(t, err, http.StatusBadRequest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockVocabularyDeckRepository(t)
			decks := mocks.NewMockVocabularyDeckItemRepository(t)
			tt.run(t, repo, decks)
		})
	}
}
