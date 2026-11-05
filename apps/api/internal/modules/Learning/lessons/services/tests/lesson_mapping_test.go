package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/utils"
)

func TestLessonDetailResponse_MapsNestedGraph(t *testing.T) {
	m := &models.Lesson{
		ID: "11111111-1111-1111-1111-111111111111", Slug: "unit-1", Title: "Unit 1",
		CEFRLevel: enums.CEFR("B1"),
		Activities: []*models.LessonActivity{
			{ID: "33333333-3333-3333-3333-333333333333", LessonID: "11111111-1111-1111-1111-111111111111", PartNumber: 1, Title: "Part 1"},
		},
	}

	var dto responses.LessonDetail
	require.NoError(t, utils.Map(&dto, m))
	require.Len(t, dto.Activities, 1)
	assert.Equal(t, 1, dto.Activities[0].PartNumber)
}

func TestLessonDetailResponse_HasNoCategory(t *testing.T) {
	m := &models.Lesson{ID: "11111111-1111-1111-1111-111111111111", Slug: "u1", Title: "U1", CEFRLevel: enums.CEFR("B1")}
	var dto responses.LessonDetail
	require.NoError(t, utils.Map(&dto, m))
	assert.Equal(t, "U1", dto.Title)
	raw, err := json.Marshal(dto)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "category")
}

func TestLessonDetailResponse_CarriesSection(t *testing.T) {
	m := &models.Lesson{
		ID: "11111111-1111-1111-1111-111111111111", Slug: "u1", Title: "U1",
		SectionID: "22222222-2222-2222-2222-222222222222", CEFRLevel: enums.CEFR("B1"),
		Section: &models.LessonSection{ID: "22222222-2222-2222-2222-222222222222", Slug: "b1", Title: "B1 Threshold", CEFRBand: enums.CEFR("B1"), Position: 3},
	}
	var dto responses.LessonDetail
	require.NoError(t, utils.Map(&dto, m))
	require.NotNil(t, dto.Section)
	assert.Equal(t, "B1 Threshold", dto.Section.Title)
	assert.Equal(t, 3, dto.Section.Position)
}

func TestListLessons_RejectsUnknownLevel(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/lessons?level=Z9", nil)
	var req requests.ListLessons
	err := c.ShouldBindQuery(&req)
	require.Error(t, err)
}

func TestListLessons_HasNoCategorySlugField(t *testing.T) {
	rt := reflect.TypeOf(requests.ListLessons{})
	_, hasField := rt.FieldByName("CategorySlug")
	assert.False(t, hasField, "CategorySlug must be gone")
}
