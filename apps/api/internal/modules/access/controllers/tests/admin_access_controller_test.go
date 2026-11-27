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
	"gorm.io/gorm"

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
			svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, nil, nil, policy)
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
	svc := accessservices.NewAdminAccessService(nil, counterRepo, eventRepo, nil, nil, nil, &stubRefresher{})
	ctl := controllers.NewAdminAccessController(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/users/user_1/usage", nil)
	c.Params = gin.Params{{Key: "id", Value: "user_1"}}

	ctl.UserUsage(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "00000000-0000-0000-0000-000000000001")
}

// stubRefresher is the SnapshotCatalog seam for controller tests.
type stubRefresher struct{ calls int }

func (s *stubRefresher) SnapshotCatalog(ctx context.Context) error { s.calls++; return nil }

func TestAdminUpdateActionCostsBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
		want int
	}{
		{
			name: "valid batch upsert",
			body: `{"costs":[{"action":"voice_session_start","credits":1},{"action":"exam_ai_review","credits":1},{"action":"study_plan","credits":2}]}`,
			want: http.StatusOK,
		},
		{
			name: "unknown action rejected",
			body: `{"costs":[{"action":"nope","credits":1},{"action":"exam_ai_review","credits":1},{"action":"study_plan","credits":2}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "absurd credits rejected",
			body: `{"costs":[{"action":"voice_session_start","credits":` + fmt.Sprint(1000000) + `},{"action":"exam_ai_review","credits":1},{"action":"study_plan","credits":2}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "short batch passes binding, fails service",
			body: `{"costs":[{"action":"voice_session_start","credits":1}]}`,
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			costs := repomocks.NewMockActionCreditCostRepository(t)
			policy := &stubRefresher{}
			if tt.want == http.StatusOK {
				costs.EXPECT().UpsertBatch(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := accessservices.NewAdminAccessService(nil, nil, nil, costs, nil, nil, policy)
			ctl := controllers.NewAdminAccessController(svc, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/admin/premium/action-costs", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("userID", "admin_1")

			ctl.UpdateActionCosts(c)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestAdminUpdateTrialTermsBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
		want int
	}{
		{
			name: "valid terms",
			body: `{"days":10,"credits":40,"voiceSessions":10}`,
			want: http.StatusOK,
		},
		{
			name: "zero days rejected",
			body: `{"days":0,"credits":40,"voiceSessions":10}`,
			want: http.StatusBadRequest,
		},
		{
			name: "absurd credits rejected",
			body: `{"days":10,"credits":` + fmt.Sprint(1000000) + `,"voiceSessions":10}`,
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			settings := repomocks.NewMockTrialSettingRepository(t)
			tx := repomocks.NewMockTransactionManager(t)
			policy := &stubRefresher{}
			if tt.want == http.StatusOK {
				tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, fn func(*gorm.DB) error) error { return fn(nil) }).Once()
				limits.EXPECT().WithTx(mock.Anything).Return(limits).Once()
				settings.EXPECT().WithTx(mock.Anything).Return(settings).Once()
				settings.EXPECT().Upsert(mock.Anything, mock.Anything).Return(nil).Once()
				limits.EXPECT().UpsertBatch(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, settings, tx, policy)
			ctl := controllers.NewAdminAccessController(svc, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/admin/premium/trial-terms", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("userID", "admin_1")

			ctl.UpdateTrialTerms(c)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}
