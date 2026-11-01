package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/modules/exams/services"
	"engflex-api/internal/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestFinalizeEstimatedScores(t *testing.T) {
	raw, err := os.ReadFile("../../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	tests := []struct {
		name                          string
		parts                         []int64
		listening, reading            int
		wantL, wantR, wantTotal       float64 // -1 means no scaled score
		expired, lateSave, finalBatch bool
	}{
		{"all skipped", []int64{1, 2, 3, 4, 5, 6, 7}, 0, 0, 5, 5, 10, false, false, false},
		{"all incorrect", []int64{1, 2, 3, 4, 5, 6, 7}, -1, -1, 5, 5, 10, false, false, false},
		{"perfect", []int64{1, 2, 3, 4, 5, 6, 7}, 100, 100, 495, 495, 990, false, false, false},
		{"mixed", []int64{1, 2, 3, 4, 5, 6, 7}, 50, 75, 245, 370, 615, false, false, false},
		{"listening only", []int64{1, 2, 3, 4}, 100, 0, 495, -1, -1, false, false, false},
		{"reading only", []int64{5, 6, 7}, 0, 100, -1, 495, -1, false, false, false},
		{"subset", []int64{1}, 6, 0, -1, -1, -1, false, false, false},
		{"expiry grades saved only", []int64{1, 2, 3, 4, 5, 6, 7}, 50, 75, 245, 370, 615, true, false, true},
		{"late save finalizes", []int64{1, 2, 3, 4, 5, 6, 7}, 50, 75, 245, 370, 615, true, true, false},
		{"final batch counted", []int64{1, 2, 3, 4}, 99, 0, 495, -1, -1, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := toeic.ParseContent(raw)
			require.NoError(t, err)
			selected := map[int64]bool{}
			for _, n := range tt.parts {
				selected[n] = true
			}
			result := toeic.Result{Revision: 1, Answers: []toeic.Answer{}}
			seenL, seenR := 0, 0
			var batch []requests.Answer
			for _, part := range content.Parts {
				if !selected[int64(part.PartNumber)] {
					continue
				}
				for _, item := range part.Items {
					for _, q := range item.Questions {
						a := toeic.Answer{QuestionID: q.ID, Flagged: true}
						seen, correct := &seenR, tt.reading
						if part.Section == "listening" {
							seen, correct = &seenL, tt.listening
						}
						(*seen)++
						if *seen <= correct {
							key := *q.CorrectKey
							a.SelectedKey = &key
						}
						if correct == -1 { // Deliberately answer every question incorrectly.
							for _, option := range q.Options {
								if option.Key != *q.CorrectKey {
									key := option.Key
									a.SelectedKey = &key
									break
								}
							}
						}
						if part.Section == "listening" && *seen == 100 && tt.finalBatch {
							key := *q.CorrectKey
							batch = append(batch, requests.Answer{QuestionID: q.ID, SelectedKey: &key, Flagged: true})
						}
						result.Answers = append(result.Answers, a)
					}
				}
			}
			encoded, err := json.Marshal(result)
			require.NoError(t, err)
			start := time.Now().Add(-time.Minute)
			mode := enums.ExamModePractice
			if tt.expired {
				mode = enums.ExamModeExam
			}
			m := &models.Attempt{ID: "run", UserID: "owner", Status: enums.AttemptStatusInProgress, Mode: &mode, SelectedParts: tt.parts, StartedAt: &start, Result: datatypes.JSON(encoded), ExamTest: &models.ExamTest{Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(raw)}}
			if tt.expired {
				deadline := time.Now().Add(-time.Second)
				m.DeadlineAt = &deadline
			}
			repo := repomocks.NewMockAttemptRepository(t)
			testRepo := repomocks.NewMockExamTestRepository(t)
			transactions := repomocks.NewMockTransactionManager(t)
			tx := &gorm.DB{}
			transactions.On("Run", mock.Anything, mock.Anything).Return(func(_ context.Context, fn func(*gorm.DB) error) error { return fn(tx) })
			repo.On("WithTx", tx).Return(repo)
			repo.On("GetExamAttemptForUpdate", mock.Anything, "owner", "run").Return(m, nil)
			repo.On("SaveExamAttempt", mock.Anything, m, 1).Return(nil).Once()
			svc := services.NewAttemptService(repo, testRepo, transactions)
			got, err := svc.Patch(context.Background(), "owner", "run", requests.PatchAttempt{Revision: 1, Answers: batch, Submit: !tt.lateSave})
			if tt.lateSave {
				var app *common.AppError
				require.ErrorAs(t, err, &app)
				require.Equal(t, http.StatusConflict, app.Status)
				got = m
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, enums.AttemptStatusCompleted, got.Status)
			parsed, err := toeic.ParseResult(got.Result)
			require.NoError(t, err)
			require.Equal(t, toeic.ScoringVersion, *parsed.ScoringVersion)
			require.Nil(t, parsed.ScaledScore)
			for i, section := range []*toeic.SectionResult{parsed.Sections.Listening, parsed.Sections.Reading} {
				want := []float64{tt.wantL, tt.wantR}[i]
				if want < 0 {
					if section != nil {
						require.Nil(t, section.ScaledScore)
					}
				} else {
					require.NotNil(t, section)
					require.Equal(t, want, *section.ScaledScore)
				}
			}
			if tt.wantTotal < 0 {
				require.Nil(t, got.Score)
			} else {
				require.Equal(t, tt.wantTotal, *got.Score)
			}
			var dto responses.ReviewAttempt
			require.NoError(t, utils.Map(&dto, got))
			require.Equal(t, parsed.Sections.Listening != nil, dto.ParsedResult.Sections.Listening != nil)
			// Repeated submit must return the existing snapshot without another save.
			before := string(m.Result)
			_, err = svc.Patch(context.Background(), "owner", "run", requests.PatchAttempt{Revision: 1, Submit: true})
			require.NoError(t, err)
			require.Equal(t, before, string(m.Result))
		})
	}
}

func TestStatsMapsDatabaseErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{{"unavailable", errors.New("unavailable")}, {"success", nil}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockAttemptRepository(t)
			testRepo := repomocks.NewMockExamTestRepository(t)
			tx := repomocks.NewMockTransactionManager(t)
			repo.On("GetExamStats", mock.Anything, "owner").Return(&models.ExamAttemptStats{}, tt.err)
			_, err := services.NewAttemptService(repo, testRepo, tx).Stats(context.Background(), "owner")
			if tt.err != nil {
				var app *common.AppError
				require.ErrorAs(t, err, &app)
				require.Equal(t, http.StatusInternalServerError, app.Status)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
