package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/services"
)

func TestExamPhase2HTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	raw, err := os.ReadFile("../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	category := &models.ExamCategory{Slug: uuid.NewString(), Title: "Synthetic", Type: enums.ExamTypeTOEIC}
	require.NoError(t, tx.Create(category).Error)
	test := &models.ExamTest{CategoryID: category.ID, SourceKey: uuid.NewString(), Slug: uuid.NewString(), Title: "Synthetic", Content: datatypes.JSON(raw)}
	require.NoError(t, tx.Create(test).Error)
	lesson := &models.Attempt{UserID: "user-a", Type: enums.AttemptTypeLesson, Status: enums.AttemptStatusInProgress, Result: datatypes.JSON(`{}`)}
	require.NoError(t, tx.Create(lesson).Error)

	categories := repositories.NewExamCategoryRepository(tx)
	var queryLog bytes.Buffer
	queryLogger := logger.New(log.New(&queryLog, "", 0), logger.Config{LogLevel: logger.Info})
	tests := repositories.NewExamTestRepository(tx.Session(&gorm.Session{Logger: queryLogger}))
	attempts := repositories.NewAttemptRepository(tx.Session(&gorm.Session{Logger: queryLogger}))
	ctl := controllers.NewExamController(services.NewCatalogService(categories, tests), services.NewAttemptService(attempts, tests, repositories.NewTransactionManager(tx)))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userID", c.GetHeader("X-Test-User")); c.Next() })
	r.GET("/exam-categories", ctl.ListCategories)
	r.GET("/exam-categories/:categoryId/tests", ctl.ListTests)
	r.GET("/exam-tests/:id", ctl.GetTest)
	r.POST("/exam-tests/:id/attempts", ctl.Start)
	r.GET("/exam-attempts", ctl.ListAttempts)
	r.GET("/exam-attempts/:id", ctl.GetAttempt)
	r.PATCH("/exam-attempts/:id", ctl.PatchAttempt)
	r.POST("/exam-attempts/:id/cancel", ctl.CancelAttempt)

	var attemptID string
	cases := []struct {
		name, method, path, body, user string
		status                         int
		check                          func(*testing.T, map[string]any, string)
	}{
		{"categories", "GET", "/exam-categories?type=toeic", "", "user-a", http.StatusOK, func(t *testing.T, v map[string]any, body string) { require.NotEmpty(t, v["data"]) }},
		{"bad category type", "GET", "/exam-categories?type=unknown", "", "user-a", http.StatusBadRequest, nil},
		{"invalid test id", "GET", "/exam-tests/bad", "", "user-a", http.StatusBadRequest, nil},
		{"unknown test", "GET", "/exam-tests/" + uuid.NewString(), "", "user-a", http.StatusNotFound, nil},
		{"test detail metadata", "GET", "/exam-tests/" + test.ID, "", "user-a", http.StatusOK, func(t *testing.T, v map[string]any, body string) {
			require.Contains(t, queryLog.String(), `LEFT JOIN "exam_categories" "Category" ON "exam_tests"."category_id" = "Category"."id"`)
			require.NotContains(t, queryLog.String(), "FOR UPDATE")
			data := v["data"].(map[string]any)
			require.Equal(t, test.ID, data["id"])
			require.Equal(t, category.ID, data["categoryId"])
			require.Equal(t, category.Title, data["category"].(map[string]any)["title"])
			parts := data["content"].(map[string]any)["parts"].([]any)
			require.Len(t, parts, 7)
			for i, count := range []int{6, 25, 39, 30, 30, 16, 54} {
				require.EqualValues(t, count, parts[i].(map[string]any)["questionCount"])
			}
			for _, forbidden := range []string{"correctKey", "PRIVATE", "questions", "transcript", "assets", "sourceKey"} {
				require.NotContains(t, body, forbidden)
			}
		}},
		{"invalid category id", "GET", "/exam-categories/bad/tests", "", "user-a", http.StatusBadRequest, nil},
		{"unknown category", "GET", "/exam-categories/" + uuid.NewString() + "/tests", "", "user-a", http.StatusNotFound, nil},
		{"test metadata", "GET", "/exam-categories/" + category.ID + "/tests", "", "user-a", http.StatusOK, func(t *testing.T, v map[string]any, body string) {
			items := v["data"].([]any)
			require.Len(t, items, 1)
			item := items[0].(map[string]any)
			parts := item["content"].(map[string]any)["parts"].([]any)
			require.Len(t, parts, 7)
			for i, count := range []int{6, 25, 39, 30, 30, 16, 54} {
				require.EqualValues(t, count, parts[i].(map[string]any)["questionCount"])
			}
			require.NotContains(t, body, "correctKey")
			require.NotContains(t, body, "PRIVATE")
		}},
		{"duplicate part", "POST", "/exam-tests/" + test.ID + "/attempts", `{"mode":"practice","selectedParts":[1,1]}`, "user-a", http.StatusBadRequest, nil},
		{"exam requires all parts", "POST", "/exam-tests/" + test.ID + "/attempts", `{"mode":"exam","selectedParts":[1,2]}`, "user-a", http.StatusBadRequest, nil},
		{"start practice", "POST", "/exam-tests/" + test.ID + "/attempts", `{"mode":"practice","selectedParts":[3,1]}`, "user-a", http.StatusCreated, func(t *testing.T, v map[string]any, body string) {
			data := v["data"].(map[string]any)
			attemptID = data["id"].(string)
			require.NotEqual(t, test.ID, attemptID)
			require.Equal(t, []any{float64(1), float64(3)}, data["selectedParts"])
			result := data["result"].(map[string]any)
			require.EqualValues(t, 1, result["revision"])
			require.Len(t, result["answers"], 45)
			require.NotContains(t, body, "correctKey")
			require.NotContains(t, body, "PRIVATE")
		}},
		{"history", "GET", "/exam-attempts?page=1&pageSize=1", "", "user-a", http.StatusOK, func(t *testing.T, v map[string]any, body string) {
			require.Equal(t, 2, strings.Count(queryLog.String(), "SELECT"), "history uses one list and one count query")
			require.NotContains(t, queryLog.String(), `"ExamTest"."content"`)
			require.Len(t, v["data"], 1)
			require.NotContains(t, body, "content")
			require.NotContains(t, body, "answers")
			require.Contains(t, body, "questionCount")
			require.NotContains(t, body, "correctKey")
			require.NotContains(t, body, "PRIVATE")
		}},
		{"other owner history", "GET", "/exam-attempts", "", "user-b", http.StatusOK, func(t *testing.T, v map[string]any, body string) { require.Empty(t, v["data"]) }},
		{"lesson attempt is hidden", "GET", "/exam-attempts/" + lesson.ID, "", "user-a", http.StatusNotFound, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("X-Test-User", tc.user)
			req.Header.Set("Content-Type", "application/json")
			if tc.name == "history" {
				queryLog.Reset()
			}
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			if tc.check != nil {
				var v map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
				tc.check(t, v, w.Body.String())
			}
		})
	}

	for _, tt := range []struct {
		name, body string
		status     int
	}{
		{"patch rejects server-owned fields", `{"revision":1,"score":99}`, http.StatusBadRequest},
		{"patch requires revision", `{"submit":true}`, http.StatusBadRequest},
		{"patch saves valid batch", `{"revision":1,"answers":[{"questionId":"q1","selectedKey":"A","flagged":true}]}`, http.StatusOK},
		{"patch rejects stale revision", `{"revision":1}`, http.StatusConflict},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/exam-attempts/"+attemptID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", "user-a")
			r.ServeHTTP(w, req)
			require.Equal(t, tt.status, w.Code, w.Body.String())
			require.NotContains(t, w.Body.String(), "correctKey")
		})
	}

	for _, tc := range []struct {
		name, user string
		status     int
	}{{"owner", "user-a", http.StatusOK}, {"other owner", "user-b", http.StatusNotFound}} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/exam-attempts/"+attemptID, nil)
			req.Header.Set("X-Test-User", tc.user)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			if tc.status == http.StatusOK {
				require.NotContains(t, w.Body.String(), "correctKey")
				require.NotContains(t, w.Body.String(), "PRIVATE")
				require.Contains(t, w.Body.String(), `"q32"`)
				require.NotContains(t, w.Body.String(), "q131")
			}
		})
	}
	t.Run("retake creates another attempt", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/exam-tests/"+test.ID+"/attempts", strings.NewReader(`{"mode":"practice","selectedParts":[1]}`))
		req.Header.Set("X-Test-User", "user-a")
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var v struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
		require.NotEqual(t, attemptID, v.Data.ID)
	})
	t.Run("history pagination counts only owner exams", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/exam-attempts?page=2&pageSize=1", nil)
		req.Header.Set("X-Test-User", "user-a")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Pagination struct {
				Total      int `json:"total"`
				TotalPages int `json:"totalPages"`
			} `json:"pagination"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.Len(t, result.Data, 1)
		require.Equal(t, attemptID, result.Data[0].ID)
		require.Equal(t, 2, result.Pagination.Total)
		require.Equal(t, 2, result.Pagination.TotalPages)
	})
	t.Run("full exam initializes all questions", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/exam-tests/"+test.ID+"/attempts", strings.NewReader(`{"mode":"exam","selectedParts":[7,6,5,4,3,2,1]}`))
		req.Header.Set("X-Test-User", "user-b")
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var result struct {
			Data struct {
				SelectedParts []int      `json:"selectedParts"`
				DeadlineAt    *time.Time `json:"deadlineAt"`
				Result        struct {
					Answers []json.RawMessage `json:"answers"`
				} `json:"result"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, result.Data.SelectedParts)
		require.Len(t, result.Data.Result.Answers, 200)
		require.NotNil(t, result.Data.DeadlineAt)
	})
	t.Run("completed review is owner only", func(t *testing.T) {
		require.NoError(t, tx.Model(&models.Attempt{}).Where("id = ?", attemptID).Updates(map[string]any{"status": enums.AttemptStatusCompleted, "submitted_at": time.Now().UTC()}).Error)
		for _, tc := range []struct {
			user   string
			status int
		}{{"user-a", http.StatusOK}, {"user-b", http.StatusNotFound}} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/exam-attempts/"+attemptID, nil)
			req.Header.Set("X-Test-User", tc.user)
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			if tc.status == http.StatusOK {
				require.Contains(t, w.Body.String(), "correctKey")
				require.Contains(t, w.Body.String(), "PRIVATE EXPLANATION")
				require.NotContains(t, w.Body.String(), "q131")
			}
		}
	})
	t.Run("routes require authentication", func(t *testing.T) {
		real := gin.New()
		exams.RegisterRoutes(real.Group("/api/v1"), tx)
		for _, endpoint := range []struct{ method, path string }{
			{"GET", "/api/v1/exam-categories"},
			{"POST", "/api/v1/exam-attempts/" + attemptID + "/cancel"},
			{"GET", "/api/v1/exam-tests/" + test.ID},
			{"GET", "/api/v1/exam-categories/" + category.ID + "/tests"},
			{"POST", "/api/v1/exam-tests/" + test.ID + "/attempts"},
			{"GET", "/api/v1/exam-attempts"},
			{"GET", "/api/v1/exam-attempts/statistics"},
			{"GET", "/api/v1/exam-attempts/" + attemptID},
			{"PATCH", "/api/v1/exam-attempts/" + attemptID},
		} {
			w := httptest.NewRecorder()
			real.ServeHTTP(w, httptest.NewRequest(endpoint.method, endpoint.path, nil))
			require.Equal(t, http.StatusUnauthorized, w.Code, endpoint.path)
		}
	})
}
