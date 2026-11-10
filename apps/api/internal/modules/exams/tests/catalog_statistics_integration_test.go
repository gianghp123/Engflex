package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/services"
)

// Uses the real migrated schema. All fixtures and requests live in a rolled-back
// transaction, so catalog counts can also be verified on an existing dev copy.
func TestExamCatalogStatisticsPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	raw, err := os.ReadFile("../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	zero, eighty, fifty, ninety := 0.0, 80.0, 50.0, 90.0
	type run struct {
		mode      enums.ExamMode
		status    enums.AttemptStatus
		score     *float64
		otherUser bool
	}
	cases := []struct {
		name                     string
		runs                     []run
		completed                int64
		highest, average, latest *float64
	}{
		{name: "empty catalog counters and unattempted user"},
		{name: "unscaled score stays null", runs: []run{{enums.ExamModeExam, enums.AttemptStatusCompleted, nil, false}}, completed: 1},
		{name: "scope mode status and zero score", runs: []run{
			{enums.ExamModeExam, enums.AttemptStatusCompleted, &zero, false},
			{enums.ExamModePractice, enums.AttemptStatusCompleted, &ninety, false},
			{enums.ExamModeExam, enums.AttemptStatusInProgress, &ninety, false},
			{enums.ExamModeExam, enums.AttemptStatusAbandoned, &ninety, false},
			{enums.ExamModeExam, enums.AttemptStatusCompleted, &ninety, true},
		}, completed: 1, highest: &zero, average: &zero, latest: &zero},
		{name: "latest follows submission and preserves null", runs: []run{
			{enums.ExamModeExam, enums.AttemptStatusCompleted, &fifty, false},
			{enums.ExamModeExam, enums.AttemptStatusCompleted, &eighty, false},
			{enums.ExamModeExam, enums.AttemptStatusCompleted, nil, false},
		}, completed: 3, highest: &eighty, average: func() *float64 { v := 65.0; return &v }()},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
			userID := uuid.NewString()
			category := &models.ExamCategory{Slug: uuid.NewString(), Title: "Catalog fixture", Type: enums.ExamTypeTOEIC}
			emptyCategory := &models.ExamCategory{Slug: uuid.NewString(), Title: "Empty fixture", Type: enums.ExamTypeTOEIC}
			require.NoError(t, tx.Create(category).Error)
			require.NoError(t, tx.Create(emptyCategory).Error)
			exam := &models.ExamTest{CategoryID: category.ID, SourceKey: uuid.NewString(), Slug: uuid.NewString(), Title: "A", Content: datatypes.JSON(raw)}
			emptyTest := &models.ExamTest{CategoryID: category.ID, SourceKey: uuid.NewString(), Slug: uuid.NewString(), Title: "B", Content: datatypes.JSON(raw)}
			require.NoError(t, tx.Create(exam).Error)
			require.NoError(t, tx.Create(emptyTest).Error)
			now := time.Now().UTC().Add(-time.Hour)
			for i, run := range tt.runs {
				owner := userID
				if run.otherUser {
					owner = uuid.NewString()
				}
				submitted := now.Add(time.Duration(i+1) * time.Minute)
				attempt := &models.Attempt{UserID: owner, Type: enums.AttemptTypeExam, ExamTestID: &exam.ID, Mode: &run.mode,
					Status: run.status, Score: run.score, StartedAt: &now, SubmittedAt: &submitted, CreatedAt: now.Add(time.Duration(len(tt.runs)-i) * time.Minute),
					SelectedParts: pq.Int64Array{1, 2, 3, 4, 5, 6, 7}, Result: datatypes.JSON(`{"revision":1,"answers":[]}`)}
				require.NoError(t, tx.Create(attempt).Error)
			}
			categories := repositories.NewExamCategoryRepository(tx)
			tests := repositories.NewExamTestRepository(tx)
			attempts := repositories.NewAttemptRepository(tx)
			ctl := controllers.NewExamController(services.NewCatalogService(categories, tests), services.NewAttemptService(attempts, tests, repositories.NewTransactionManager(tx)))
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("userID", userID) })
			router.GET("/categories", ctl.ListCategories)
			router.GET("/categories/:categoryId/tests", ctl.ListTests)
			router.GET("/tests/:id", ctl.GetTest)
			router.GET("/statistics", ctl.Statistics)
			get := func(path string) map[string]any {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				if path == "/statistics" {
					require.Equal(t, "private, no-store", w.Header().Get("Cache-Control"))
				}
				var body map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				return body
			}
			foundCategory, foundEmpty := false, false
			for _, item := range get("/categories?type=toeic")["data"].([]any) {
				cat := item.(map[string]any)
				if cat["id"] == category.ID {
					foundCategory = true
					require.EqualValues(t, 2, cat["testsCount"])
					require.EqualValues(t, len(tt.runs), cat["attemptsCount"])
				}
				if cat["id"] == emptyCategory.ID {
					foundEmpty = true
					require.EqualValues(t, 0, cat["testsCount"])
					require.EqualValues(t, 0, cat["attemptsCount"])
				}
			}
			require.True(t, foundCategory)
			require.True(t, foundEmpty)
			listing := get("/categories/" + category.ID + "/tests?page=1&pageSize=1")
			require.EqualValues(t, 2, listing["pagination"].(map[string]any)["total"])
			items := listing["data"].([]any)
			require.Len(t, items, 1)
			require.Equal(t, exam.ID, items[0].(map[string]any)["id"])
			require.EqualValues(t, len(tt.runs), items[0].(map[string]any)["attemptsCount"])
			detail := get("/tests/" + exam.ID)["data"].(map[string]any)
			require.EqualValues(t, len(tt.runs), detail["attemptsCount"])
			require.EqualValues(t, 120, detail["durationMinutes"])
			require.Equal(t, category.ID, detail["category"].(map[string]any)["id"])
			require.Len(t, detail["content"].(map[string]any)["parts"], 7)
			page2 := get("/categories/" + category.ID + "/tests?page=2&pageSize=1")["data"].([]any)
			require.Len(t, page2, 1)
			require.Equal(t, emptyTest.ID, page2[0].(map[string]any)["id"])
			require.EqualValues(t, 0, page2[0].(map[string]any)["attemptsCount"])
			stats := get("/statistics")["data"].(map[string]any)
			require.EqualValues(t, tt.completed, stats["completedTests"])
			for key, want := range map[string]*float64{"highestScore": tt.highest, "averageScore": tt.average, "latestScore": tt.latest} {
				if want == nil {
					require.Nil(t, stats[key], key)
				} else {
					require.EqualValues(t, *want, stats[key], key)
				}
			}
			// Locking/join reads still use the physical models without aggregate columns.
			locked, err := tests.GetForUpdate(t.Context(), exam.ID)
			require.NoError(t, err)
			require.Equal(t, exam.ID, locked.ID)
			started, err := services.NewAttemptService(attempts, tests, repositories.NewTransactionManager(tx)).Start(t.Context(), userID, exam.ID, requests.StartAttempt{Mode: enums.ExamModeExam, SelectedParts: []int{1, 2, 3, 4, 5, 6, 7}})
			require.NoError(t, err)
			require.Equal(t, time.Duration(detail["durationMinutes"].(float64))*time.Minute, started.DeadlineAt.Sub(*started.StartedAt))
		})
	}
}
