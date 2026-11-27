package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"engflex-api/config"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/services"
)

func TestCreateCategoryPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	repo := repositories.NewExamCategoryRepository(tx)
	svc := services.NewCatalogService(repo, repositories.NewExamTestRepository(tx))
	ctx := context.Background()
	subType := enums.ExamSubTypeAcademic
	for _, req := range []requests.CreateCategory{
		{Slug: uuid.NewString(), Title: "ETS fixture", Type: enums.ExamTypeTOEIC},
		{Slug: uuid.NewString(), Title: "Cambridge fixture", Type: enums.ExamTypeIELTS, SubType: &subType},
	} {
		category, err := svc.CreateCategory(ctx, req)
		require.NoError(t, err)
		require.NoError(t, uuid.Validate(category.ID))
		require.False(t, category.CreatedAt.IsZero())
		stored, err := repo.GetByID(ctx, category.ID)
		require.NoError(t, err)
		require.Equal(t, req.Slug, stored.Slug)
		require.Equal(t, req.Title, stored.Title)
		require.Equal(t, req.Type, stored.Type)
		require.Equal(t, req.SubType, stored.SubType)
		categories, err := repo.List(ctx, req.Type)
		require.NoError(t, err)
		found := false
		for _, item := range categories {
			if item.ID == category.ID {
				found = true
				require.Zero(t, item.TestsCount)
				require.Zero(t, item.AttemptsCount)
			}
		}
		require.True(t, found)
		require.NoError(t, tx.SavePoint("before_duplicate").Error)
		_, err = svc.CreateCategory(ctx, req)
		var appErr *common.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, http.StatusConflict, appErr.Status)
		require.NoError(t, tx.RollbackTo("before_duplicate").Error)
	}

	// Check the production route actually applies authentication before writes.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	exams.RegisterRoutes(r.Group("/api/v1"), tx, config.LLMConfig{}, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/exam-categories", nil))
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
