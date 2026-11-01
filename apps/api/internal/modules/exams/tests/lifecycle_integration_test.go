package tests

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestExamLifecyclePostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(8)
	defer pool.Close()
	raw, err := os.ReadFile("../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	category := &models.ExamCategory{Slug: uuid.NewString(), Title: "Synthetic", Type: enums.ExamTypeTOEIC}
	require.NoError(t, db.Create(category).Error)
	exam := &models.ExamTest{CategoryID: category.ID, SourceKey: uuid.NewString(), Slug: uuid.NewString(), Title: "Synthetic", Content: datatypes.JSON(raw)}
	require.NoError(t, db.Create(exam).Error)
	defer func() {
		require.NoError(t, db.Where("exam_test_id = ?", exam.ID).Delete(&models.Attempt{}).Error)
		require.NoError(t, db.Delete(exam).Error)
		require.NoError(t, db.Delete(category).Error)
	}()
	svc := services.NewAttemptService(repositories.NewAttemptRepository(db), repositories.NewExamTestRepository(db), repositories.NewTransactionManager(db))
	ctx := context.Background()
	start := func() *models.Attempt {
		m, err := svc.Start(ctx, "owner", exam.ID, requests.StartAttempt{Mode: enums.ExamModePractice, SelectedParts: []int{2}})
		require.NoError(t, err)
		return m
	}
	key := "A"
	bad := "D"
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"save then final batch and retry", func(t *testing.T) {
			m := start()
			_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key, Flagged: true}}})
			require.NoError(t, err)
			_, err = svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1})
			var app *common.AppError
			require.ErrorAs(t, err, &app)
			require.Equal(t, http.StatusConflict, app.Status)
			result, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 2, Submit: true, Answers: []requests.Answer{{QuestionID: "q8", SelectedKey: &key}}})
			require.NoError(t, err)
			require.Equal(t, 2, *result.ParsedResult.CorrectCount)
			require.Equal(t, 8.0, *result.ParsedResult.AccuracyPct)
			require.Nil(t, result.Score)
			retry, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Submit: true, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &bad}}})
			require.NoError(t, err)
			require.Equal(t, result.ParsedResult.Revision, retry.ParsedResult.Revision)
			require.Equal(t, 2, *retry.ParsedResult.CorrectCount)
			_, err = svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 3})
			require.ErrorAs(t, err, &app)
			require.Equal(t, http.StatusConflict, app.Status)
		}},
		{"invalid batch rolls back", func(t *testing.T) {
			m := start()
			_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}, {QuestionID: "q8", SelectedKey: &bad}}})
			require.Error(t, err)
			read, err := svc.Get(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, 1, read.ParsedResult.Revision)
			require.Nil(t, read.ParsedResult.Answers[0].SelectedKey)
		}},
		{"deadline commits despite late save conflict", func(t *testing.T) {
			m := start()
			past := time.Now().Add(-time.Second)
			started := past.Add(-120 * time.Minute)
			require.NoError(t, db.Model(&models.Attempt{}).Where("id = ?", m.ID).Updates(map[string]any{"deadline_at": past, "started_at": started}).Error)
			_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
			require.Error(t, err)
			read, err := svc.Get(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, enums.AttemptStatusCompleted, read.Status)
			require.Equal(t, 0, *read.ParsedResult.CorrectCount)
			require.Equal(t, 7200, *read.DurationSec)
		}},
		{"GET finalizes expired attempt", func(t *testing.T) {
			m := start()
			past := time.Now().Add(-time.Second)
			require.NoError(t, db.Model(&models.Attempt{}).Where("id = ?", m.ID).Updates(map[string]any{"deadline_at": past, "started_at": past.Add(-time.Minute)}).Error)
			read, err := svc.Get(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, enums.AttemptStatusCompleted, read.Status)
		}},
		{"concurrent same revision only one commits", func(t *testing.T) {
			m := start()
			var wg sync.WaitGroup
			out := make(chan error, 2)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
					out <- err
				}()
			}
			wg.Wait()
			close(out)
			success := 0
			conflict := 0
			for err := range out {
				if err == nil {
					success++
				} else {
					var app *common.AppError
					require.ErrorAs(t, err, &app)
					require.Equal(t, http.StatusConflict, app.Status)
					conflict++
				}
			}
			require.Equal(t, 1, success)
			require.Equal(t, 1, conflict)
		}},
		{"clear selection and flags", func(t *testing.T) {
			m := start()
			_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key, Flagged: true}}})
			require.NoError(t, err)
			read, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 2, Answers: []requests.Answer{{QuestionID: "q7"}}})
			require.NoError(t, err)
			require.Nil(t, read.ParsedResult.Answers[0].SelectedKey)
			require.False(t, read.ParsedResult.Answers[0].Flagged)
		}},
		{"late submit ignores batch", func(t *testing.T) {
			m := start()
			past := time.Now().Add(-time.Second)
			require.NoError(t, db.Model(&models.Attempt{}).Where("id = ?", m.ID).Updates(map[string]any{"deadline_at": past, "started_at": past.Add(-time.Minute)}).Error)
			read, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Submit: true, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
			require.NoError(t, err)
			require.Equal(t, 0, *read.ParsedResult.CorrectCount)
		}},
		{"save submit race preserves one revision", func(t *testing.T) {
			m := start()
			out := make(chan error, 2)
			var wg sync.WaitGroup
			for _, submit := range []bool{false, true} {
				wg.Add(1)
				go func(submit bool) {
					defer wg.Done()
					_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Submit: submit, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
					out <- err
				}(submit)
			}
			wg.Wait()
			close(out)
			success := 0
			for err := range out {
				if err == nil {
					success++
				} else {
					var app *common.AppError
					require.ErrorAs(t, err, &app)
					require.Equal(t, http.StatusConflict, app.Status)
				}
			}
			require.Equal(t, 1, success)
			read, err := svc.Get(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, 2, read.ParsedResult.Revision)
		}},
		{"cancel retains saved answers and history", func(t *testing.T) {
			m := start()
			_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key, Flagged: true}}})
			require.NoError(t, err)
			cancelled, err := svc.Cancel(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, enums.AttemptStatusAbandoned, cancelled.Status)
			require.Equal(t, 3, cancelled.ParsedResult.Revision)
			require.Equal(t, key, *cancelled.ParsedResult.Answers[0].SelectedKey)
			require.True(t, cancelled.ParsedResult.Answers[0].Flagged)
			require.Nil(t, cancelled.SubmittedAt)
			repeated, err := svc.Cancel(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, 3, repeated.ParsedResult.Revision)
			_, err = svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 3, Submit: true})
			var app *common.AppError
			require.ErrorAs(t, err, &app)
			require.Equal(t, http.StatusConflict, app.Status)
			history, _, err := svc.List(ctx, "owner", 1, 100)
			require.NoError(t, err)
			found := false
			for _, entry := range history {
				if entry.ID == m.ID {
					found = true
					require.Equal(t, enums.AttemptStatusAbandoned, entry.Status)
					require.Equal(t, 25, entry.ParsedResult.QuestionCount)
				}
			}
			require.True(t, found)
		}},
		{"expired cancel grades saved answers and remains idempotent", func(t *testing.T) {
			m, err := svc.Start(ctx, "owner", exam.ID, requests.StartAttempt{Mode: enums.ExamModeExam, SelectedParts: []int{1, 2, 3, 4, 5, 6, 7}})
			require.NoError(t, err)
			_, err = svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
			require.NoError(t, err)
			past := time.Now().Add(-time.Second)
			require.NoError(t, db.Model(&models.Attempt{}).Where("id = ?", m.ID).Updates(map[string]any{"deadline_at": past, "started_at": past.Add(-120 * time.Minute), "mode": enums.ExamModeExam}).Error)
			completed, err := svc.Cancel(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, enums.AttemptStatusCompleted, completed.Status)
			require.Equal(t, 7200, *completed.DurationSec)
			require.Equal(t, 1, *completed.ParsedResult.CorrectCount)
			require.NotNil(t, completed.Score)
			require.Equal(t, 10.0, *completed.Score)
			again, err := svc.Cancel(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, completed.ParsedResult.Revision, again.ParsedResult.Revision)
		}},
		{"cancel save race preserves final state", func(t *testing.T) {
			m := start()
			var wg sync.WaitGroup
			errs := make(chan error, 2)
			wg.Add(2)
			go func() { defer wg.Done(); _, err := svc.Cancel(ctx, "owner", m.ID); errs <- err }()
			go func() {
				defer wg.Done()
				_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Answers: []requests.Answer{{QuestionID: "q7", SelectedKey: &key}}})
				errs <- err
			}()
			wg.Wait()
			close(errs)
			for err := range errs {
				if err != nil {
					var app *common.AppError
					require.ErrorAs(t, err, &app)
					require.Equal(t, http.StatusConflict, app.Status)
				}
			}
			var stored models.Attempt
			require.NoError(t, db.First(&stored, "id = ?", m.ID).Error)
			require.Equal(t, enums.AttemptStatusAbandoned, stored.Status)
			require.Nil(t, stored.SubmittedAt)
		}},
		{"cancel submit race closes exactly once", func(t *testing.T) {
			m := start()
			var wg sync.WaitGroup
			errs := make(chan error, 2)
			wg.Add(2)
			go func() { defer wg.Done(); _, err := svc.Cancel(ctx, "owner", m.ID); errs <- err }()
			go func() {
				defer wg.Done()
				_, err := svc.Patch(ctx, "owner", m.ID, requests.PatchAttempt{Revision: 1, Submit: true})
				errs <- err
			}()
			wg.Wait()
			close(errs)
			for err := range errs {
				if err != nil {
					var app *common.AppError
					require.ErrorAs(t, err, &app)
					require.Equal(t, http.StatusConflict, app.Status)
				}
			}
			var stored models.Attempt
			require.NoError(t, db.First(&stored, "id = ?", m.ID).Error)
			require.Contains(t, []enums.AttemptStatus{enums.AttemptStatusAbandoned, enums.AttemptStatusCompleted}, stored.Status)
			again, err := svc.Cancel(ctx, "owner", m.ID)
			require.NoError(t, err)
			require.Equal(t, 2, again.ParsedResult.Revision)
		}},
		{"owner isolation", func(t *testing.T) {
			m := start()
			_, err := svc.Patch(ctx, "other", m.ID, requests.PatchAttempt{Revision: 1, Submit: true})
			var app *common.AppError
			require.ErrorAs(t, err, &app)
			require.Equal(t, http.StatusNotFound, app.Status)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
