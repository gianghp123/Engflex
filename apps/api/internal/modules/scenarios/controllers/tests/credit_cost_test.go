package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/scenarios/controllers"
	"engflex-api/internal/modules/scenarios/services"
)

// freePremiumChecker defaults to "not premium" — the fixtures below are
// non-premium content, so the lookup never matters.
type freePremiumChecker struct{}

func (freePremiumChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

func publishedScenario(id string) *models.Scenario {
	return &models.Scenario{
		ID:         id,
		Title:      "Card",
		Status:     enums.PublishStatusPublished,
		Objectives: datatypes.JSON(`[]`),
	}
}

func newScenarioRouter(scenarios *repomocks.MockScenarioRepository, topics *repomocks.MockScenarioTopicRepository) *gin.Engine {
	svc := services.NewScenarioService(scenarios, topics, freePremiumChecker{})
	ctl := controllers.NewScenarioController(svc)
	r := gin.New()
	// The string "userID" must match middleware.userIDKey (unexported), which
	// is the gin context key production's RequireAuth stamps.
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	r.GET("/scenarios", ctl.List)
	r.GET("/scenarios/:id", ctl.GetByID)
	r.GET("/scenario-topics", ctl.TopicsWithPreview)
	return r
}

func decodeData(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	data, ok := envelope["data"]
	require.True(t, ok, "envelope has no data: %s", w.Body.String())
	out, ok := data.(map[string]any)
	// List endpoints return an array; callers that need it assert the slice
	// form separately.
	if !ok {
		require.IsType(t, []any{}, data)
		return map[string]any{"__list__": data}
	}
	return out
}

// TestScenarioListStampsCreditCost pins the card badge at the HTTP layer:
// every scenario on the list carries aiCreditCost == 1 (voice start).
func TestScenarioListStampsCreditCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scenarios := repomocks.NewMockScenarioRepository(t)
	topics := repomocks.NewMockScenarioTopicRepository(t)
	scenarios.On("List", mock.Anything, (*string)(nil), true, 20, 0).
		Return([]*models.Scenario{publishedScenario("s1")}, nil)
	scenarios.On("Count", mock.Anything, (*string)(nil), true).Return(int64(1), nil)

	w := httptest.NewRecorder()
	newScenarioRouter(scenarios, topics).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/scenarios", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var envelope struct {
		Data []struct {
			ID           string `json:"id"`
			AICreditCost int    `json:"aiCreditCost"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data, 1)
	assert.Equal(t, 1, envelope.Data[0].AICreditCost)
}

// TestScenarioDetailStampsCreditCost pins the badge on the detail page.
func TestScenarioDetailStampsCreditCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scenarios := repomocks.NewMockScenarioRepository(t)
	topics := repomocks.NewMockScenarioTopicRepository(t)
	scenarios.On("GetDetail", mock.Anything, "s1").Return(publishedScenario("s1"), nil)

	w := httptest.NewRecorder()
	newScenarioRouter(scenarios, topics).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/scenarios/s1", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, float64(1), decodeData(t, w)["aiCreditCost"])
}

// TestScenarioTopicPreviewStampsCreditCost pins the badge on every nested
// preview card inside the topics hub.
func TestScenarioTopicPreviewStampsCreditCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scenarios := repomocks.NewMockScenarioRepository(t)
	topics := repomocks.NewMockScenarioTopicRepository(t)
	topics.On("ListWithPreview", mock.Anything, 3, (*string)(nil), "").
		Return([]*models.ScenarioTopic{{
			ID: "t1", Slug: "job", Name: "Job", ShortName: "Job", Position: 1,
			Scenarios: []*models.Scenario{publishedScenario("s1"), publishedScenario("s2")},
		}}, nil)

	w := httptest.NewRecorder()
	newScenarioRouter(scenarios, topics).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/scenario-topics", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var envelope struct {
		Data []struct {
			Scenarios []struct {
				AICreditCost int `json:"aiCreditCost"`
			} `json:"scenarios"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data, 1)
	require.Len(t, envelope.Data[0].Scenarios, 2)
	for _, s := range envelope.Data[0].Scenarios {
		assert.Equal(t, 1, s.AICreditCost)
	}
}
