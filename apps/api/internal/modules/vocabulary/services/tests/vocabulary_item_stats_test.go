package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestVocabularyItemService_Stats pins the hub stat cards with a fixed clock
// (the service takes now as an argument, never time.Now()) and proves a
// failing count surfaces the mapped common error status. `dayStart` is the
// caller's local midnight, so "saved today" is that local day, not UTC's.
func TestVocabularyItemService_Stats(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	// UTC+7 local midnight: the window the web sends.
	dayStart := time.Date(2026, 10, 15, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		dayStart  time.Time
		setupMock func(items *mocks.MockVocabularyItemRepository, state *mocks.MockUserVocabularyRepository)
		want      services.VocabularyStats
		status    int
	}{
		{
			name:     "maps every count",
			dayStart: dayStart,
			setupMock: func(items *mocks.MockVocabularyItemRepository, state *mocks.MockUserVocabularyRepository) {
				state.On("CountByUser", mock.Anything, "user_1").Return(int64(10), nil)
				state.On("CountMastered", mock.Anything, "user_1").Return(int64(4), nil)
				state.On("CountDue", mock.Anything, "user_1", now).Return(int64(3), nil)
				state.On("CountSavedBetween", mock.Anything, "user_1", dayStart, dayStart.AddDate(0, 0, 1)).Return(int64(5), nil)
				items.On("CountPrivate", mock.Anything, "user_1").Return(int64(2), nil)
			},
			want: services.VocabularyStats{
				TotalSaved: 10, Mastered: 4, NeedsReview: 3, CustomAdditions: 2, DueToday: 3, SavedToday: 5,
			},
		},
		{
			name: "count failure surfaces mapped status",
			setupMock: func(items *mocks.MockVocabularyItemRepository, state *mocks.MockUserVocabularyRepository) {
				state.On("CountByUser", mock.Anything, "user_1").Return(int64(0), gorm.ErrRecordNotFound)
			},
			status: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mocks.NewMockVocabularyItemRepository(t)
			state := mocks.NewMockUserVocabularyRepository(t)
			tt.setupMock(items, state)
			svc := services.NewVocabularyItemService(items, state, mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

			got, err := svc.Stats(context.Background(), "user_1", now, tt.dayStart)

			if tt.status != 0 {
				requireAppError(t, err, tt.status)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
