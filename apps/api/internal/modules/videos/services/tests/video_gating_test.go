package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/services"
)

const gatingVideoID = "00000000-0000-0000-0000-000000000001"

func gatingExercise(premium bool) *models.VideoExercise {
	return &models.VideoExercise{ID: gatingVideoID, Status: enums.PublishStatusPublished, IsPremium: premium}
}

// runVideoAttemptTx wires the transaction plumbing for a successful start:
// the locked exercise read, the open-attempt miss, and the insert.
func runVideoAttemptTx(t *testing.T, exercises *mocks.MockVideoExerciseRepository, attempts *mocks.MockAttemptRepository, txMgr *mocks.MockTransactionManager, exercise *models.VideoExercise) {
	t.Helper()
	exercisesTx := mocks.NewMockVideoExerciseRepository(t)
	attemptsTx := mocks.NewMockAttemptRepository(t)
	exercises.EXPECT().WithTx(mock.Anything).Return(exercisesTx).Once()
	exercisesTx.EXPECT().GetByIDForUpdate(mock.Anything, gatingVideoID).Return(exercise, nil).Once()
	attempts.EXPECT().WithTx(mock.Anything).Return(attemptsTx).Once()
	attemptsTx.EXPECT().GetVideoAttempt(mock.Anything, "user_1", gatingVideoID, enums.AttemptTypeDictation).
		Return(nil, gorm.ErrRecordNotFound).Once()
	attemptsTx.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
	txMgr.EXPECT().Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(*gorm.DB) error) error { return fn(nil) }).Once()
}

func TestVideoAttemptStartGatingFollowsFlag(t *testing.T) {
	tests := []struct {
		name     string
		gatingOn bool
		exercise *models.VideoExercise
		wantErr  int // 0 = nil
	}{
		{name: "gating off starts attempts on premium exercises", gatingOn: false, exercise: gatingExercise(true)},
		{name: "gating on blocks a premium exercise for a free user", gatingOn: true, exercise: gatingExercise(true), wantErr: 403},
		{name: "gating on passes free exercises", gatingOn: true, exercise: gatingExercise(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			exercises := mocks.NewMockVideoExerciseRepository(t)
			if tt.gatingOn {
				exercises.EXPECT().GetByID(mock.Anything, gatingVideoID).Return(tt.exercise, nil).Once()
			}
			attempts := mocks.NewMockAttemptRepository(t)
			txMgr := mocks.NewMockTransactionManager(t)
			if tt.wantErr == 0 {
				runVideoAttemptTx(t, exercises, attempts, txMgr, tt.exercise)
			}
			svc := services.NewVideoAttemptService(
				attempts, exercises, mocks.NewMockVideoTranscriptRepository(t), txMgr, fakePremiumChecker{},
			)

			m, err := svc.StartAttempt(context.Background(), "user_1", gatingVideoID,
				enums.AttemptTypeDictation, requests.StartVideoAttempt{Type: enums.AttemptTypeDictation})
			if tt.wantErr == 0 {
				require.NoError(t, err)
				require.NotNil(t, m)
			} else {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, m)
			}
		})
	}
}

func TestVideoTranscriptReadGatingFollowsFlag(t *testing.T) {
	lines := []*models.VideoTranscript{
		{ID: "00000000-0000-0000-0000-000000000002", VideoExerciseID: gatingVideoID, Sequence: 1, Content: "hello"},
	}
	tests := []struct {
		name     string
		gatingOn bool
		exercise *models.VideoExercise
		wantErr  int // 0 = nil
	}{
		{name: "gating off reads transcripts on premium exercises", gatingOn: false, exercise: gatingExercise(true)},
		{name: "gating on blocks transcripts on a premium exercise", gatingOn: true, exercise: gatingExercise(true), wantErr: 403},
		{name: "gating on passes transcripts on free exercises", gatingOn: true, exercise: gatingExercise(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			exercises := mocks.NewMockVideoExerciseRepository(t)
			if tt.gatingOn {
				exercises.EXPECT().GetByID(mock.Anything, gatingVideoID).Return(tt.exercise, nil).Once()
			}
			transcripts := mocks.NewMockVideoTranscriptRepository(t)
			if tt.wantErr == 0 {
				transcripts.EXPECT().GetByVideoExerciseID(mock.Anything, gatingVideoID).Return(lines, nil).Once()
			}
			svc := services.NewVideoTranscriptService(transcripts, exercises, fakePremiumChecker{})

			got, err := svc.GetByVideoExerciseID(context.Background(), gatingVideoID, "user_1")
			if tt.wantErr == 0 {
				require.NoError(t, err)
				require.Len(t, got, 1)
			} else {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, got)
			}
		})
	}
}

func TestVideoTranscriptGetByIDGatingFollowsFlag(t *testing.T) {
	transcriptID := "00000000-0000-0000-0000-000000000003"
	line := &models.VideoTranscript{ID: transcriptID, VideoExerciseID: gatingVideoID, Sequence: 1, Content: "hello"}
	tests := []struct {
		name     string
		gatingOn bool
		exercise *models.VideoExercise
		wantErr  int // 0 = nil
	}{
		{name: "gating off reads single transcripts on premium exercises", gatingOn: false, exercise: gatingExercise(true)},
		{name: "gating on blocks a single transcript on a premium exercise", gatingOn: true, exercise: gatingExercise(true), wantErr: 403},
		{name: "gating on passes single transcripts on free exercises", gatingOn: true, exercise: gatingExercise(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			transcripts := mocks.NewMockVideoTranscriptRepository(t)
			transcripts.EXPECT().GetByID(mock.Anything, transcriptID).Return(line, nil).Once()
			exercises := mocks.NewMockVideoExerciseRepository(t)
			if tt.gatingOn {
				exercises.EXPECT().GetByID(mock.Anything, gatingVideoID).Return(tt.exercise, nil).Once()
			}
			svc := services.NewVideoTranscriptService(transcripts, exercises, fakePremiumChecker{})

			got, err := svc.GetByID(context.Background(), transcriptID, "user_1")
			if tt.wantErr == 0 {
				require.NoError(t, err)
				require.NotNil(t, got)
			} else {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, got)
			}
		})
	}
}
