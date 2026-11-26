// Admin PUT binding table (spec §8.6): the same gin binding tags the
// controller enforces, driven through the real controller + service with
// mocked repos. Unknown counters, negative and absurd limits fail in
// binding (400) before any store call; unknown plans fail in the service.
package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/controllers"
	accessservices "engflex-api/internal/modules/access/services"
)

func TestAdminUpdatePlanLimitsBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		plan string
		body string
		want int
	}{
		{
			name: "valid batch upsert",
			plan: "free",
			body: `{"counters":[{"counterKey":"ai_credit","isEnabled":true,"limitValue":15,"period":"month"},{"counterKey":"voice_session","isEnabled":true,"limitValue":3,"period":"month"}]}`,
			want: http.StatusOK,
		},
		{
			name: "unknown counter key rejected",
			plan: "free",
			body: `{"counters":[{"counterKey":"tokens","limitValue":10,"period":"month"}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "negative limit rejected",
			plan: "free",
			body: `{"counters":[{"counterKey":"ai_credit","limitValue":-1,"period":"month"}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "absurd limit rejected",
			plan: "free",
			body: `{"counters":[{"counterKey":"ai_credit","limitValue":` + fmt.Sprint(1000000) + `,"period":"month"}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "unknown plan rejected",
			plan: "mega",
			body: `{"counters":[{"counterKey":"ai_credit","limitValue":10,"period":"month"}]}`,
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			policy := &stubRefresher{}
			if tt.want == http.StatusOK {
				limits.EXPECT().UpsertBatch(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := accessservices.NewAdminAccessService(limits, nil, nil, policy)
			ctl := controllers.NewAdminAccessController(svc, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/admin/premium/limits/"+tt.plan, bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "planKey", Value: tt.plan}}
			c.Set("userID", "admin_1")

			ctl.UpdatePlanLimits(c)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestAdminUserUsageEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	counters := []*models.UsageCounter{
		{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Used: 4},
	}
	events := []*models.UsageEvent{
		{ID: "00000000-0000-0000-0000-000000000001", UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Amount: 1},
	}
	counterRepo := repomocks.NewMockUsageCounterRepository(t)
	eventRepo := repomocks.NewMockUsageEventRepository(t)
	counterRepo.EXPECT().ListByUser(mock.Anything, "user_1").Return(counters, nil).Once()
	eventRepo.EXPECT().ListByUser(mock.Anything, "user_1", 50).Return(events, nil).Once()
	svc := accessservices.NewAdminAccessService(nil, counterRepo, eventRepo, &stubRefresher{})
	ctl := controllers.NewAdminAccessController(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/users/user_1/usage", nil)
	c.Params = gin.Params{{Key: "id", Value: "user_1"}}

	ctl.UserUsage(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "00000000-0000-0000-0000-000000000001")
}

// stubRefresher is the SnapshotRefresher seam for controller tests.
type stubRefresher struct{ calls int }

func (s *stubRefresher) SnapshotLimits(ctx context.Context) error { s.calls++; return nil }
