package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/services"
)

func newExamRouter(t *testing.T, attempts *repomocks.MockAttemptRepository, tests *repomocks.MockExamTestRepository, tx *repomocks.MockTransactionManager) *gin.Engine {
	catalog := services.NewCatalogService(repomocks.NewMockExamCategoryRepository(t), tests)
	ctl := controllers.NewExamController(catalog, services.NewAttemptService(attempts, tests, tx))
	r := gin.New()
	// The string "userID" must match middleware.userIDKey (unexported), which
	// is the gin context key production's RequireAuth stamps.
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	r.POST("/exam-tests/:id/attempts", ctl.Start)
	r.GET("/exam-attempts/:id", ctl.GetAttempt)
	r.PATCH("/exam-attempts/:id", ctl.PatchAttempt)
	return r
}

func syntheticContent(t *testing.T) datatypes.JSON {
	t.Helper()
	raw, err := os.ReadFile("../../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	return datatypes.JSON(raw)
}

func invokeTx(tx *repomocks.MockTransactionManager) *gorm.DB {
	transaction := &gorm.DB{}
	tx.On("Run", mock.Anything, mock.Anything).
		Return(func(_ context.Context, fn func(*gorm.DB) error) error { return fn(transaction) })
	return transaction
}

// TestStartStampsCreditCosts pins the paid-AI badges on a freshly started
// attempt: review costs 1 credit, the study plan 2.
func TestStartStampsCreditCosts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	attempts := repomocks.NewMockAttemptRepository(t)
	tests := repomocks.NewMockExamTestRepository(t)
	tx := repomocks.NewMockTransactionManager(t)
	transaction := invokeTx(tx)
	tests.On("WithTx", transaction).Return(tests)
	tests.On("GetForUpdate", mock.Anything, "test-1").Return(&models.ExamTest{
		ID:       "test-1",
		Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC},
		Content:  syntheticContent(t),
	}, nil)
	attempts.On("WithTx", transaction).Return(attempts)
	attempts.On("Create", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/exam-tests/test-1/attempts", strings.NewReader(`{"mode":"practice","selectedParts":[1]}`))
	req.Header.Set("Content-Type", "application/json")
	newExamRouter(t, attempts, tests, tx).ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var envelope struct {
		Data struct {
			AIReviewCreditCost  int `json:"aiReviewCreditCost"`
			StudyPlanCreditCost int `json:"studyPlanCreditCost"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	assert.Equal(t, 1, envelope.Data.AIReviewCreditCost)
	assert.Equal(t, 2, envelope.Data.StudyPlanCreditCost)
}

func completedAttempt() *models.Attempt {
	now := time.Now().UTC()
	return &models.Attempt{
		ID:            "attempt-1",
		UserID:        "user-1",
		Type:          enums.AttemptTypeExam,
		Status:        enums.AttemptStatusCompleted,
		StartedAt:     &now,
		SelectedParts: pq.Int64Array{1},
		Result:        datatypes.JSON(`{"revision":1,"answers":[]}`),
		ExamTest: &models.ExamTest{
			ID:       "test-1",
			Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC},
		},
	}
}

// TestGetCompletedStampsCreditCosts pins both badges on the completed
// (review) view of an attempt.
func TestGetCompletedStampsCreditCosts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	attempts := repomocks.NewMockAttemptRepository(t)
	tests := repomocks.NewMockExamTestRepository(t)
	tx := repomocks.NewMockTransactionManager(t)
	transaction := invokeTx(tx)
	m := completedAttempt()
	m.ExamTest.Content = syntheticContent(t)
	attempts.On("WithTx", transaction).Return(attempts)
	attempts.On("GetExamAttemptForUpdate", mock.Anything, "user-1", "attempt-1").Return(m, nil)

	w := httptest.NewRecorder()
	newExamRouter(t, attempts, tests, tx).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/exam-attempts/attempt-1", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var envelope struct {
		Data struct {
			AIReviewCreditCost  int `json:"aiReviewCreditCost"`
			StudyPlanCreditCost int `json:"studyPlanCreditCost"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	assert.Equal(t, 1, envelope.Data.AIReviewCreditCost)
	assert.Equal(t, 2, envelope.Data.StudyPlanCreditCost)
}

// TestPatchCompletedStampsCreditCosts pins both badges on the submit-to-complete
// path, which maps a second ReviewAttempt DTO.
func TestPatchCompletedStampsCreditCosts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	attempts := repomocks.NewMockAttemptRepository(t)
	tests := repomocks.NewMockExamTestRepository(t)
	tx := repomocks.NewMockTransactionManager(t)
	transaction := invokeTx(tx)
	m := completedAttempt()
	m.ExamTest.Content = syntheticContent(t)
	attempts.On("WithTx", transaction).Return(attempts)
	attempts.On("GetExamAttemptForUpdate", mock.Anything, "user-1", "attempt-1").Return(m, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/exam-attempts/attempt-1", strings.NewReader(`{"revision":1,"submit":true}`))
	req.Header.Set("Content-Type", "application/json")
	newExamRouter(t, attempts, tests, tx).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var envelope struct {
		Data struct {
			Status              string `json:"status"`
			AIReviewCreditCost  int    `json:"aiReviewCreditCost"`
			StudyPlanCreditCost int    `json:"studyPlanCreditCost"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	assert.Equal(t, "completed", envelope.Data.Status)
	assert.Equal(t, 1, envelope.Data.AIReviewCreditCost)
	assert.Equal(t, 2, envelope.Data.StudyPlanCreditCost)
}
