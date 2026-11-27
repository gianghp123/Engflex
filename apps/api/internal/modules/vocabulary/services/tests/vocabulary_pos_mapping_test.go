package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/utils"
)

// part_of_speech is carried end to end through a silent-failure-prone pair:
// requests -> model goes through copier (utils.Map), and model -> response
// goes through copier again in the controller. A field-name or pointer-type
// drift on either hop yields no error and an empty value, so both hops are
// pinned here.
func TestVocabularyItem_PartOfSpeechRoundTrip(t *testing.T) {
	noun := enums.VocabularyPartOfSpeechNoun
	verb := enums.VocabularyPartOfSpeechVerb

	tests := []struct {
		name string
		pos  *enums.VocabularyPartOfSpeech
	}{
		{name: "set is carried into the model and the response DTO", pos: &noun},
		{name: "nil stays nil", pos: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupItem(t)
			var created *models.VocabularyItem
			mockRepo.On("Create", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					created = args.Get(1).(*models.VocabularyItem)
				}).
				Return(nil)

			_, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{
				Term:         "idempotent",
				CEFR:         "C1",
				PartOfSpeech: tt.pos,
			})
			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, tt.pos, created.PartOfSpeech, "request -> model")

			var dto responses.VocabularyItem
			require.NoError(t, utils.Map(&dto, created))
			assert.Equal(t, tt.pos, dto.PartOfSpeech, "model -> response DTO")
		})
	}

	t.Run("update replaces it, nil leaves it unchanged", func(t *testing.T) {
		existing := &models.VocabularyItem{
			ID: itemID, Term: "old", Status: enums.PublishStatusDraft,
			PartOfSpeech: &noun,
		}
		svc, mockRepo := setupItem(t)
		mockRepo.On("GetByID", mock.Anything, itemID).Return(existing, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		res, err := svc.Update(context.Background(), itemID, requests.UpdateVocabularyItem{
			PartOfSpeech: &verb,
		})
		require.NoError(t, err)
		require.NotNil(t, res.PartOfSpeech)
		assert.Equal(t, verb, *res.PartOfSpeech)

		res, err = svc.Update(context.Background(), itemID, requests.UpdateVocabularyItem{
			Term: strPtr("renamed"),
		})
		require.NoError(t, err)
		require.NotNil(t, res.PartOfSpeech)
		assert.Equal(t, verb, *res.PartOfSpeech, "absent field must not clear it")
		assert.Equal(t, "renamed", res.Term)
	})
}
