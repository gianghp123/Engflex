package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/services"
)

func TestCreateCategoryHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		status     int
		create     bool
		dbErr      error
		role       enums.UserRole
	}{
		{name: "toeic", body: `{"slug":"ets-2026","title":"ETS 2026","type":"toeic"}`, status: http.StatusCreated, create: true, role: enums.UserRoleAdmin},
		{name: "ielts", body: `{"slug":"cambridge-20","title":"Cambridge 20","type":"ielts","subType":"academic"}`, status: http.StatusCreated, create: true, role: enums.UserRoleAdmin},
		{name: "ielts general", body: `{"slug":"general","title":"General","type":"ielts","subType":"general_training"}`, status: http.StatusCreated, create: true, role: enums.UserRoleAdmin},
		{name: "ielts null subtype", body: `{"slug":"ielts","title":"IELTS","type":"ielts","subType":null}`, status: http.StatusCreated, create: true, role: enums.UserRoleAdmin},
		{name: "duplicate", body: `{"slug":"ets-2026","title":"ETS 2026","type":"toeic"}`, status: http.StatusConflict, create: true, dbErr: &pgconn.PgError{Code: "23505"}, role: enums.UserRoleAdmin},
		{name: "database failure", body: `{"slug":"ets-2026","title":"ETS 2026","type":"toeic"}`, status: http.StatusInternalServerError, create: true, dbErr: errors.New("private database failure"), role: enums.UserRoleAdmin},
		{name: "non-admin", body: `{"slug":"ets-2026","title":"ETS 2026","type":"toeic"}`, status: http.StatusCreated, create: true, role: enums.UserRoleUser},
		{name: "missing fields", body: `{}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "invalid type", body: `{"slug":"test","title":"Test","type":"other"}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "invalid subtype", body: `{"slug":"test","title":"Test","type":"ielts","subType":"other"}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "toeic subtype", body: `{"slug":"test","title":"Test","type":"toeic","subType":"academic"}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "blank title", body: `{"slug":"test","title":"   ","type":"toeic"}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "blank slug", body: `{"slug":"   ","title":"Test","type":"toeic"}`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
		{name: "malformed JSON", body: `{`, status: http.StatusBadRequest, role: enums.UserRoleAdmin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			categories := repomocks.NewMockExamCategoryRepository(t)
			tests := repomocks.NewMockExamTestRepository(t)
			attempts := repomocks.NewMockAttemptRepository(t)
			txm := repomocks.NewMockTransactionManager(t)
			if tc.create {
				categories.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					category := args.Get(1).(*models.ExamCategory)
					var expected models.ExamCategory
					require.NoError(t, json.Unmarshal([]byte(tc.body), &expected))
					require.Equal(t, expected, *category)
					category.ID = uuid.NewString()
				}).Return(tc.dbErr).Once()
			}
			ctl := controllers.NewExamController(services.NewCatalogService(categories, tests), services.NewAttemptService(attempts, tests, txm))
			r := gin.New()
			r.POST("/exam-categories", func(c *gin.Context) { c.Set("role", tc.role) }, ctl.CreateCategory)
			req := httptest.NewRequest(http.MethodPost, "/exam-categories", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			require.NotContains(t, rec.Body.String(), "private database failure")
			if tc.status == http.StatusCreated {
				var response struct {
					Data models.ExamCategory `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				require.NoError(t, uuid.Validate(response.Data.ID))
				var expected models.ExamCategory
				require.NoError(t, json.Unmarshal([]byte(tc.body), &expected))
				expected.ID = response.Data.ID
				require.Equal(t, expected, response.Data)
			}
		})
	}
}
