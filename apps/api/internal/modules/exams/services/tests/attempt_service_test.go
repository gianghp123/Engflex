package tests

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/services"
)

func TestStartRejectsInvalidChoices(t *testing.T) {
	tests := []struct {
		name string
		req  requests.StartAttempt
	}{
		{"invalid mode", requests.StartAttempt{Mode: "other", SelectedParts: []int{1}}},
		{"empty selection", requests.StartAttempt{Mode: enums.ExamModePractice}},
		{"duplicate selection", requests.StartAttempt{Mode: enums.ExamModePractice, SelectedParts: []int{1, 1}}},
		{"out of range", requests.StartAttempt{Mode: enums.ExamModePractice, SelectedParts: []int{8}}},
		{"incomplete full exam", requests.StartAttempt{Mode: enums.ExamModeExam, SelectedParts: []int{1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			tests := repomocks.NewMockExamTestRepository(t)
			tx := repomocks.NewMockTransactionManager(t)
			svc := services.NewAttemptService(attempts, tests, tx)
			_, err := svc.Start(context.Background(), "user", "test", tt.req)
			var appErr *common.AppError
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, http.StatusBadRequest, appErr.Status)
			tx.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
		})
	}
}

func TestStartMapsRepositoryAndContentFailures(t *testing.T) {
	raw, err := os.ReadFile("../../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	tests := []struct {
		name              string
		test              *models.ExamTest
		getErr, createErr error
		status            int
	}{
		{"missing test", nil, gorm.ErrRecordNotFound, nil, http.StatusNotFound},
		{"unsupported type", &models.ExamTest{Category: &models.ExamCategory{Type: "other"}}, nil, nil, http.StatusUnprocessableEntity},
		{"invalid content", &models.ExamTest{Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(`{}`)}, nil, nil, http.StatusInternalServerError},
		{"write failure", &models.ExamTest{ID: "test", Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(raw)}, nil, errors.New("database failed"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			repo := repomocks.NewMockExamTestRepository(t)
			transactions := repomocks.NewMockTransactionManager(t)
			transaction := &gorm.DB{}
			transactions.On("Run", mock.Anything, mock.Anything).Return(func(_ context.Context, fn func(*gorm.DB) error) error { return fn(transaction) })
			repo.On("WithTx", transaction).Return(repo)
			repo.On("GetForUpdate", mock.Anything, "test").Return(tt.test, tt.getErr)
			if tt.createErr != nil {
				attempts.On("WithTx", transaction).Return(attempts)
				attempts.On("Create", mock.Anything, mock.Anything).Return(tt.createErr)
			}
			svc := services.NewAttemptService(attempts, repo, transactions)
			_, err := svc.Start(context.Background(), "user", "test", requests.StartAttempt{Mode: enums.ExamModePractice, SelectedParts: []int{1}})
			var appErr *common.AppError
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, tt.status, appErr.Status)
		})
	}
}

func TestHistoryResultParsing(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		status int
	}{
		{"valid summary", `{"revision":2,"answers":[{"questionId":"q1","selectedKey":"A","flagged":false}],"correctCount":1,"accuracyPct":100,"scaledScore":null}`, http.StatusOK},
		{"invalid result", `{}`, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockAttemptRepository(t)
			tests := repomocks.NewMockExamTestRepository(t)
			tx := repomocks.NewMockTransactionManager(t)
			repo.On("ListExamAttempts", mock.Anything, "owner", 10, 0).Return([]*models.Attempt{{ID: "attempt", Result: datatypes.JSON(tt.raw)}}, nil)
			if tt.status == http.StatusOK {
				repo.On("CountExamAttempts", mock.Anything, "owner").Return(int64(1), nil)
			}
			got, _, err := services.NewAttemptService(repo, tests, tx).List(context.Background(), "owner", 1, 10)
			if tt.status != http.StatusOK {
				var app *common.AppError
				require.ErrorAs(t, err, &app)
				require.Equal(t, tt.status, app.Status)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 1, got[0].ParsedResult.QuestionCount)
			require.Equal(t, 1, *got[0].ParsedResult.CorrectCount)
			require.Equal(t, float64(100), *got[0].ParsedResult.AccuracyPct)
		})
	}
}

func TestCancelAttempt(t *testing.T) {
	raw, err := os.ReadFile("../../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	tests := []struct {
		name                string
		status              enums.AttemptStatus
		expired             bool
		getErr, errorOnSave error
		want                enums.AttemptStatus
		wantHTTP            int
	}{
		{"active practice", enums.AttemptStatusInProgress, false, nil, nil, enums.AttemptStatusAbandoned, http.StatusOK},
		{"expired exam", enums.AttemptStatusInProgress, true, nil, nil, enums.AttemptStatusCompleted, http.StatusOK},
		{"already abandoned", enums.AttemptStatusAbandoned, false, nil, nil, enums.AttemptStatusAbandoned, http.StatusOK},
		{"already completed", enums.AttemptStatusCompleted, false, nil, nil, enums.AttemptStatusCompleted, http.StatusOK},
		{"missing owner", enums.AttemptStatusInProgress, false, gorm.ErrRecordNotFound, nil, "", http.StatusNotFound},
		{"invalid id", enums.AttemptStatusInProgress, false, &common.InvalidIDError{Name: "attemptId"}, nil, "", http.StatusBadRequest},
		{"database failure", enums.AttemptStatusInProgress, false, errors.New("database unavailable"), nil, "", http.StatusInternalServerError},
		{"save failure", enums.AttemptStatusInProgress, false, nil, errors.New("write failed"), "", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockAttemptRepository(t)
			testRepo := repomocks.NewMockExamTestRepository(t)
			transactions := repomocks.NewMockTransactionManager(t)
			transaction := &gorm.DB{}
			transactions.On("Run", mock.Anything, mock.Anything).Return(func(_ context.Context, fn func(*gorm.DB) error) error { return fn(transaction) })
			repo.On("WithTx", transaction).Return(repo)
			started := time.Now().Add(-time.Minute)
			m := &models.Attempt{ID: "attempt", Status: tt.status, StartedAt: &started, SelectedParts: []int64{1},
				ExamTest: &models.ExamTest{Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(raw)},
				Result:   datatypes.JSON(`{"revision":1,"answers":[{"questionId":"q1","selectedKey":"A","flagged":true},{"questionId":"q2"},{"questionId":"q3"},{"questionId":"q4"},{"questionId":"q5"},{"questionId":"q6"}]}`)}
			if tt.expired {
				past := time.Now().Add(-time.Second)
				m.DeadlineAt = &past
			}
			repo.On("GetExamAttemptForUpdate", mock.Anything, "owner", "attempt").Return(m, tt.getErr)
			if tt.getErr == nil && tt.status == enums.AttemptStatusInProgress {
				repo.On("SaveExamAttempt", mock.Anything, m, 1).Return(tt.errorOnSave)
			}
			got, err := services.NewAttemptService(repo, testRepo, transactions).Cancel(context.Background(), "owner", "attempt")
			if tt.wantHTTP != http.StatusOK {
				var app *common.AppError
				require.ErrorAs(t, err, &app)
				require.Equal(t, tt.wantHTTP, app.Status)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got.Status)
			require.Equal(t, "A", *got.ParsedResult.Answers[0].SelectedKey)
			require.True(t, got.ParsedResult.Answers[0].Flagged)
			require.Equal(t, 6, got.ParsedResult.QuestionCount)
			if tt.status == enums.AttemptStatusInProgress {
				require.Equal(t, 2, got.ParsedResult.Revision)
				require.NotNil(t, got.DurationSec)
				if tt.expired {
					require.NotNil(t, got.SubmittedAt)
					require.NotNil(t, got.ParsedResult.CorrectCount)
				} else {
					require.Nil(t, got.SubmittedAt)
					require.Nil(t, got.ParsedResult.CorrectCount)
				}
			} else {
				repo.AssertNotCalled(t, "SaveExamAttempt", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStartIeltsAttempt(t *testing.T) {
	ieltsJSON := `{
		"schemaVersion": "1.0.0",
		"testId": "ielts-test-1",
		"title": "IELTS Academic Reading 1",
		"skill": "reading",
		"subType": "academic",
		"sections": [
			{"sectionNumber": 1, "title": "Passage 1", "passages": [], "questionGroups": [{"id": "g1", "type": "multiple_choice", "questions": [{"id": "q1", "questionNumber": 1, "correctKey": "A"}]}]},
			{"sectionNumber": 2, "title": "Passage 2", "passages": [], "questionGroups": [{"id": "g2", "type": "multiple_choice", "questions": [{"id": "q2", "questionNumber": 2, "correctKey": "B"}]}]},
			{"sectionNumber": 3, "title": "Passage 3", "passages": [], "questionGroups": [{"id": "g3", "type": "multiple_choice", "questions": [{"id": "q3", "questionNumber": 3, "correctKey": "C"}]}]}
		]
	}`

	skillReading := enums.ExamSkillReading
	subTypeAcademic := enums.ExamSubTypeAcademic
	ieltsTest := &models.ExamTest{
		ID:       "ielts-1",
		Skill:    &skillReading,
		SubType:  &subTypeAcademic,
		Content:  datatypes.JSON(ieltsJSON),
		Category: &models.ExamCategory{Type: enums.ExamTypeIELTS},
	}

	tests := []struct {
		name       string
		req        requests.StartAttempt
		wantStatus int
	}{
		{
			name:       "practice mode section 1",
			req:        requests.StartAttempt{Mode: enums.ExamModePractice, SelectedParts: []int{1}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "exam mode requires all 3 sections for reading",
			req:        requests.StartAttempt{Mode: enums.ExamModeExam, SelectedParts: []int{1, 2}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "exam mode all 3 sections succeeds",
			req:        requests.StartAttempt{Mode: enums.ExamModeExam, SelectedParts: []int{1, 2, 3}},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			repo := repomocks.NewMockExamTestRepository(t)
			transactions := repomocks.NewMockTransactionManager(t)
			transaction := &gorm.DB{}
			transactions.On("Run", mock.Anything, mock.Anything).Return(func(_ context.Context, fn func(*gorm.DB) error) error { return fn(transaction) })
			repo.On("WithTx", transaction).Return(repo)
			repo.On("GetForUpdate", mock.Anything, "ielts-1").Return(ieltsTest, nil)
			if tt.wantStatus == http.StatusOK {
				attempts.On("WithTx", transaction).Return(attempts)
				attempts.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			svc := services.NewAttemptService(attempts, repo, transactions)
			attempt, err := svc.Start(context.Background(), "user", "ielts-1", tt.req)
			if tt.wantStatus != http.StatusOK {
				var appErr *common.AppError
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantStatus, appErr.Status)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, attempt)
			require.NotNil(t, attempt.ParsedIeltsResult)
			require.NotNil(t, attempt.ExamTest.ParsedIeltsContent)
		})
	}
}

func TestIeltsHistoryResultParsing(t *testing.T) {
	ieltsJSON := `{"revision":2,"answers":[{"questionId":"q1","textValue":"kitchen","flagged":false}],"correctCount":1,"bandScore":6.5}`
	tests := []struct {
		name   string
		raw    string
		status int
	}{
		{"valid ielts summary", ieltsJSON, http.StatusOK},
		{"corrupt ielts summary", `{}`, http.StatusInternalServerError},
	}

	skill := enums.ExamSkillReading
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockAttemptRepository(t)
			tests := repomocks.NewMockExamTestRepository(t)
			tx := repomocks.NewMockTransactionManager(t)
			repo.On("ListExamAttempts", mock.Anything, "owner", 10, 0).Return([]*models.Attempt{
				{
					ID:     "attempt-ielts",
					Result: datatypes.JSON(tt.raw),
					ExamTest: &models.ExamTest{
						Skill: &skill,
						Category: &models.ExamCategory{
							Type: enums.ExamTypeIELTS,
						},
					},
				},
			}, nil)
			if tt.status == http.StatusOK {
				repo.On("CountExamAttempts", mock.Anything, "owner").Return(int64(1), nil)
			}
			got, _, err := services.NewAttemptService(repo, tests, tx).List(context.Background(), "owner", 1, 10)
			if tt.status != http.StatusOK {
				var app *common.AppError
				require.ErrorAs(t, err, &app)
				require.Equal(t, tt.status, app.Status)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got[0].ParsedIeltsResult)
			require.Equal(t, 1, *got[0].ParsedIeltsResult.CorrectCount)
			require.Equal(t, 6.5, got[0].ParsedIeltsResult.BandScore)
		})
	}
}
