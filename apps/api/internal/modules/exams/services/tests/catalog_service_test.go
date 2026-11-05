package tests

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/exams/services"
)

func TestGetCatalogTest(t *testing.T) {
	raw, err := os.ReadFile("../../toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	tests := []struct {
		name   string
		test   *models.ExamTest
		err    error
		status int
	}{
		{"metadata counts", &models.ExamTest{Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(raw)}, nil, http.StatusOK},
		{"invalid id", nil, &common.InvalidIDError{Name: "testId"}, http.StatusBadRequest},
		{"missing test", nil, gorm.ErrRecordNotFound, http.StatusNotFound},
		{"database failure", nil, errors.New("database unavailable"), http.StatusInternalServerError},
		{"missing category", &models.ExamTest{}, nil, http.StatusUnprocessableEntity},
		{"unsupported type", &models.ExamTest{Category: &models.ExamCategory{Type: "other"}}, nil, http.StatusUnprocessableEntity},
		{"invalid content", &models.ExamTest{Category: &models.ExamCategory{Type: enums.ExamTypeTOEIC}, Content: datatypes.JSON(`{}`)}, nil, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories := repomocks.NewMockExamCategoryRepository(t)
			repo := repomocks.NewMockExamTestRepository(t)
			var catalog *models.ExamTestCatalog
			if tt.test != nil {
				catalog = &models.ExamTestCatalog{ExamTest: *tt.test}
			}
			repo.On("GetByID", mock.Anything, "test").Return(catalog, tt.err)
			svc := services.NewCatalogService(categories, repo)
			got, err := svc.GetTest(context.Background(), "test")
			if tt.status != http.StatusOK {
				var appErr *common.AppError
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.status, appErr.Status)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got.ParsedContent)
			require.Len(t, got.ParsedContent.Parts, 7)
			for i, count := range []int{6, 25, 39, 30, 30, 16, 54} {
				require.Equal(t, count, got.ParsedContent.Parts[i].QuestionCount)
			}
		})
	}
}
