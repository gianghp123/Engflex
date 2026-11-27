package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/ielts"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/modules/exams/services"
)

// Opt-in verification of real imported packages. Every generated attempt rolls back.
func TestIeltsPublishedPackagesHTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" || os.Getenv("TEST_PUBLISHED_IELTS") != "1" {
		t.Skip("TEST_DATABASE_URL and TEST_PUBLISHED_IELTS=1 required; import the two IELTS packages first")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	tests := repositories.NewExamTestRepository(tx)
	ctl := controllers.NewExamController(services.NewCatalogService(repositories.NewExamCategoryRepository(tx), tests), services.NewAttemptService(repositories.NewAttemptRepository(tx), tests, repositories.NewTransactionManager(tx)))
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("userID", c.GetHeader("X-Test-User")) })
	router.GET("/tests/:id", ctl.GetTest)
	router.POST("/tests/:id/attempts", ctl.Start)
	router.GET("/attempts/:id", ctl.GetAttempt)
	router.PATCH("/attempts/:id", ctl.PatchAttempt)
	router.POST("/attempts/:id/cancel", ctl.CancelAttempt)
	call := func(method, path string, payload any, user string, status int) []byte {
		t.Helper()
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		request := httptest.NewRequest(method, path, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-User", user)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		require.Equal(t, status, recorder.Code, recorder.Body.String())
		var response struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		return response.Data
	}
	for _, source := range []string{"study4_6842", "study4_6832"} {
		t.Run(source, func(t *testing.T) {
			var exam models.ExamTest
			require.NoError(t, tx.Where("source_key = ?", source).First(&exam).Error)
			content, err := ielts.ParseContent(exam.Content)
			require.NoError(t, err)
			metadata := call(http.MethodGet, "/tests/"+exam.ID, nil, "ielts-smoke", http.StatusOK)
			var detail responses.TestDetail
			require.NoError(t, json.Unmarshal(metadata, &detail))
			total := 0
			for _, section := range detail.ParsedIeltsContent.Sections {
				total += section.QuestionCount
			}
			require.Equal(t, 40, total)
			for _, forbidden := range []string{"sourceAnswer", "transcript", "correctKey", "context"} {
				require.NotContains(t, string(metadata), `"`+forbidden+`":`)
			}
			parts := make([]int, len(content.Sections))
			answers := []requests.Answer{}
			for i, section := range content.Sections {
				parts[i] = section.SectionNumber
				for _, group := range section.QuestionGroups {
					for _, question := range group.Questions {
						answer := requests.Answer{QuestionID: question.ID, SelectedKey: question.CorrectKey, SelectedKeys: question.CorrectKeys}
						if len(question.CorrectAnswers) > 0 {
							value := question.CorrectAnswers[0] + "."
							answer.TextValue = &value
						}
						answers = append(answers, answer)
					}
				}
			}
			started := call(http.MethodPost, "/tests/"+exam.ID+"/attempts", map[string]any{"mode": "exam", "selectedParts": parts}, "ielts-smoke", http.StatusCreated)
			for _, forbidden := range []string{"sourceAnswer", "transcript", "correctKey", "correctAnswers", "explanation"} {
				require.NotContains(t, string(started), `"`+forbidden+`":`)
			}
			var taking responses.ExamAttempt
			require.NoError(t, json.Unmarshal(started, &taking))
			require.NotNil(t, taking.DeadlineAt)
			require.Len(t, taking.ParsedIeltsResult.Answers, 40)
			if source == "study4_6842" {
				require.NotNil(t, taking.ExamTest.ParsedIeltsContent.Sections[0].AudioAssetID)
				require.Contains(t, string(started), `"tag":"table"`)
				require.Contains(t, string(started), `"tag":"img"`)
			}
			path := "/attempts/" + taking.ID
			call(http.MethodPatch, path, requests.PatchAttempt{Revision: 1, Answers: answers[:1]}, "ielts-smoke", http.StatusOK)
			call(http.MethodPatch, path, requests.PatchAttempt{Revision: 1}, "ielts-smoke", http.StatusConflict)
			resumed := call(http.MethodGet, path, nil, "ielts-smoke", http.StatusOK)
			require.NoError(t, json.Unmarshal(resumed, &taking))
			require.Equal(t, 2, taking.ParsedIeltsResult.Revision)
			call(http.MethodPatch, path, requests.PatchAttempt{Revision: 2, Answers: []requests.Answer{{QuestionID: answers[0].QuestionID}}}, "ielts-smoke", http.StatusOK)
			graded := call(http.MethodPatch, path, requests.PatchAttempt{Revision: 3, Submit: true, Answers: answers}, "ielts-smoke", http.StatusOK)
			var review responses.ReviewAttempt
			require.NoError(t, json.Unmarshal(graded, &review))
			require.EqualValues(t, 40, *review.ParsedIeltsResult.CorrectCount)
			require.EqualValues(t, 9, *review.ParsedIeltsResult.BandScore)
			if source == "study4_6842" {
				require.NotEmpty(t, review.ExamTest.ParsedIeltsContent.Sections[0].Transcript)
			}
			require.Contains(t, string(graded), "sourceAnswer")
			call(http.MethodGet, path, nil, "another-user", http.StatusNotFound)
			repeated := call(http.MethodPatch, path, requests.PatchAttempt{Revision: 3, Submit: true}, "ielts-smoke", http.StatusOK)
			var repeatedReview responses.ReviewAttempt
			require.NoError(t, json.Unmarshal(repeated, &repeatedReview))
			require.Equal(t, review.ParsedIeltsResult, repeatedReview.ParsedIeltsResult)
			require.Equal(t, review.Score, repeatedReview.Score)
			require.Equal(t, review.Status, repeatedReview.Status)
			practice := call(http.MethodPost, "/tests/"+exam.ID+"/attempts", map[string]any{"mode": "practice", "selectedParts": parts[:1]}, "ielts-smoke", http.StatusCreated)
			taking = responses.ExamAttempt{}
			require.NoError(t, json.Unmarshal(practice, &taking))
			require.Nil(t, taking.DeadlineAt)
			require.Len(t, taking.ExamTest.ParsedIeltsContent.Sections, 1)
			call(http.MethodPost, "/attempts/"+taking.ID+"/cancel", nil, "ielts-smoke", http.StatusOK)
		})
	}
}
