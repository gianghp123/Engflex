package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/voice/pronunciation"
)

func TestStartAttemptGatingFollowsFlag(t *testing.T) {
	tests := []struct {
		name     string
		gatingOn bool
		lesson   *models.Lesson
		wantErr  int // 0 = nil
	}{
		{name: "gating off starts attempts on premium lessons", gatingOn: false, lesson: &models.Lesson{IsPremium: true}},
		{name: "gating on blocks a premium lesson for a free user", gatingOn: true, lesson: &models.Lesson{IsPremium: true}, wantErr: 403},
		{name: "gating on passes free lessons", gatingOn: true, lesson: &models.Lesson{IsPremium: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			lessons := repomocks.NewMockLessonRepository(t)
			if tt.gatingOn {
				lessons.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000001").
					Return(tt.lesson, nil).Once()
			}
			attempts := repomocks.NewMockAttemptRepository(t)
			attempts.EXPECT().GetOpenAttempt(mock.Anything, "user_1", "00000000-0000-0000-0000-000000000001").
				Return(nil, gorm.ErrRecordNotFound).Maybe()
			if tt.wantErr == 0 {
				attempts.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := services.NewLessonActivityService(
				repomocks.NewMockLessonActivityRepository(t), unusedScorer{}, unusedLLM{}, attempts, lessons, fakeNotPremium{},
			)
			m, err := svc.StartAttempt(context.Background(), "user_1", "00000000-0000-0000-0000-000000000001")
			if tt.wantErr == 0 {
				require.NoError(t, err)
				require.NotNil(t, m)
			} else {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantErr, appErr.Status)
			}
		})
	}
}

// fakeNotPremium satisfies common.PremiumChecker as "not premium" so the
// gating branch under test is deterministic.
type fakeNotPremium struct{}

func (fakeNotPremium) IsPremium(ctx context.Context, userID string) (bool, error) { return false, nil }

// unusedScorer/unusedLLM satisfy the scorer/LLM interfaces for constructor
// completeness on paths this file never exercises (the house rule bans nil
// into constructors).
type unusedScorer struct{}

func (unusedScorer) Score(ctx context.Context, text string, audio []byte, mime string) (*pronunciation.Result, error) {
	return nil, errors.New("unused")
}

type unusedLLM struct{}

func (unusedLLM) CompleteText(ctx context.Context, prompt string) (string, error) {
	return "", errors.New("unused")
}

func (unusedLLM) CompleteStructured(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) error {
	return errors.New("unused")
}
