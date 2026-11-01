package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/videos/services"
)

// fakePremiumChecker defaults to "not premium" — most tests exercise
// free-content behavior and don't care about the subscription lookup.
type fakePremiumChecker struct {
	isPremium bool
}

func (f fakePremiumChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	return f.isPremium, nil
}

func setupExercise(t *testing.T) (services.VideoExerciseService, *mocks.MockVideoExerciseRepository) {
	t.Helper()
	return setupExerciseWithPremium(t, false)
}

func setupExerciseWithPremium(t *testing.T, isPremium bool) (services.VideoExerciseService, *mocks.MockVideoExerciseRepository) {
	t.Helper()
	mockRepo := mocks.NewMockVideoExerciseRepository(t)
	svc := services.NewVideoExerciseService(mockRepo, fakePremiumChecker{isPremium: isPremium})
	return svc, mockRepo
}

func TestVideoExerciseService_GetByID(t *testing.T) {
	exerciseID := "11111111-1111-1111-1111-111111111111"
	graph := &models.VideoExercise{
		ID:         exerciseID,
		CategoryID: strPtr("22222222-2222-2222-2222-222222222222"),
		Title:      "Trailer",
		Status:     enums.PublishStatusPublished,
		Category:   &models.VideoCategory{ID: "22222222-2222-2222-2222-222222222222", Slug: "trailer", Name: "Trailer"},
		Transcripts: []*models.VideoTranscript{
			{ID: "33333333-3333-3333-3333-333333333333", VideoExerciseID: exerciseID, Sequence: 1, Content: "hello"},
			{ID: "44444444-4444-4444-4444-444444444444", VideoExerciseID: exerciseID, Sequence: 2, Content: "world"},
		},
	}

	tests := []struct {
		name      string
		setupMock func(m *mocks.MockVideoExerciseRepository)
		wantErr   bool
		want      *models.VideoExercise
	}{
		{
			name: "success: returns exercise with category and transcripts",
			setupMock: func(m *mocks.MockVideoExerciseRepository) {
				m.On("GetDetail", mock.Anything, exerciseID).Return(graph, nil)
			},
			wantErr: false,
			want:    graph,
		},
		{
			name: "not found: missing id returns 404",
			setupMock: func(m *mocks.MockVideoExerciseRepository) {
				m.On("GetDetail", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupExercise(t)
			tt.setupMock(mockRepo)

			id := exerciseID
			if tt.wantErr {
				id = "missing"
			}
			res, err := svc.GetByID(context.Background(), id, "user_1", false)

			if tt.wantErr {
				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotNil(t, res.Category)
				assert.Equal(t, "trailer", res.Category.Slug)
				require.Len(t, res.Transcripts, 2)
				assert.Equal(t, "hello", res.Transcripts[0].Content)
				assert.Equal(t, "world", res.Transcripts[1].Content)
			}
		})
	}
}

func TestVideoExerciseService_GetByID_DBError(t *testing.T) {
	svc, mockRepo := setupExercise(t)
	mockRepo.On("GetDetail", mock.Anything, mock.Anything).
		Return(nil, errors.New("db connection failure"))

	res, err := svc.GetByID(context.Background(), "11111111-1111-1111-1111-111111111111", "user_1", false)
	requireAppError(t, err, http.StatusInternalServerError)
	assert.Nil(t, res)
}

// A draft exercise is invisible to non-admins (404) but visible to admins —
// same rule as vocabulary items/decks.
func TestVideoExerciseService_GetByID_DraftHiddenFromNonAdmin(t *testing.T) {
	exerciseID := "11111111-1111-1111-1111-111111111111"
	draft := &models.VideoExercise{ID: exerciseID, Status: enums.PublishStatusDraft}

	svc, mockRepo := setupExercise(t)
	mockRepo.On("GetDetail", mock.Anything, exerciseID).Return(draft, nil)

	res, err := svc.GetByID(context.Background(), exerciseID, "user_1", false)
	requireAppError(t, err, http.StatusNotFound)
	assert.Nil(t, res)
}

func TestVideoExerciseService_GetByID_DraftVisibleToAdmin(t *testing.T) {
	exerciseID := "11111111-1111-1111-1111-111111111111"
	draft := &models.VideoExercise{ID: exerciseID, Status: enums.PublishStatusDraft}

	svc, mockRepo := setupExercise(t)
	mockRepo.On("GetDetail", mock.Anything, exerciseID).Return(draft, nil)

	res, err := svc.GetByID(context.Background(), exerciseID, "admin_1", true)
	require.NoError(t, err)
	require.NotNil(t, res)
}

// A published premium exercise stays visible as far as existence goes, but
// is blocked (403) for a non-premium, non-admin caller.
func TestVideoExerciseService_GetByID_PremiumGating(t *testing.T) {
	exerciseID := "11111111-1111-1111-1111-111111111111"
	premium := &models.VideoExercise{ID: exerciseID, Status: enums.PublishStatusPublished, IsPremium: true}

	t.Run("blocked for non-premium caller", func(t *testing.T) {
		svc, mockRepo := setupExerciseWithPremium(t, false)
		mockRepo.On("GetDetail", mock.Anything, exerciseID).Return(premium, nil)

		res, err := svc.GetByID(context.Background(), exerciseID, "user_1", false)
		requireAppError(t, err, http.StatusForbidden)
		assert.Nil(t, res)
	})

	t.Run("allowed for premium caller", func(t *testing.T) {
		svc, mockRepo := setupExerciseWithPremium(t, true)
		mockRepo.On("GetDetail", mock.Anything, exerciseID).Return(premium, nil)

		res, err := svc.GetByID(context.Background(), exerciseID, "user_1", false)
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}

func strPtr(s string) *string { return &s }
