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

func TestGetCatalogIeltsTest(t *testing.T) {
	ieltsJSON := `{
		"schemaVersion": "1.0.0",
		"testId": "ielts-reading-test",
		"title": "IELTS Academic Reading Test 1",
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
		ID:       "ielts-test-id",
		Skill:    &skillReading,
		SubType:  &subTypeAcademic,
		Content:  datatypes.JSON(ieltsJSON),
		Category: &models.ExamCategory{Type: enums.ExamTypeIELTS},
	}

	categories := repomocks.NewMockExamCategoryRepository(t)
	repo := repomocks.NewMockExamTestRepository(t)
	catalog := &models.ExamTestCatalog{ExamTest: *ieltsTest}
	repo.On("GetByID", mock.Anything, "ielts-test-id").Return(catalog, nil)

	svc := services.NewCatalogService(categories, repo)
	got, err := svc.GetTest(context.Background(), "ielts-test-id")
	require.NoError(t, err)
	require.NotNil(t, got.ParsedIeltsContent)
	require.Len(t, got.ParsedIeltsContent.Sections, 3)
	require.Equal(t, enums.ExamSkillReading, got.ParsedIeltsContent.Skill)
	require.Equal(t, enums.ExamSubTypeAcademic, got.ParsedIeltsContent.SubType)
}
