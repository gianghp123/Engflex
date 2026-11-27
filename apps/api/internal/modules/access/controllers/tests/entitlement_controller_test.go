// GET /premium/usage controller (spec §10): the 200 body carries plan,
// counters, costs, and history; unauthenticated requests stop at the
// RequireAuth gate with 401 before the service is touched.
package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/controllers"
	accessservices "engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
)

// fakeSubs is the SubscriptionPlanSource seam: nil sub resolves free.
type fakeSubs struct {
	sub *models.Subscription
	err error
}

func (f fakeSubs) GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	return f.sub, f.err
}

func usageControllerSvc(t *testing.T) *accessservices.EntitlementsService {
	t.Helper()
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
	}, nil).Once()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}, nil).Once()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Once()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1",
		[]enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved},
		20, 0).Return([]*models.UsageEvent{
		{ID: "evt-1", UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11",
			Action: enums.AccessActionExamReview, Amount: 1, Status: enums.UsageEventStatusCommitted,
			RefType: enums.UsageRefTypeExamAttempt, RefID: "attempt-1"},
	}, nil).Once()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(1), nil).Once()
	policy := accessservices.NewPolicyService(fakeSubs{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	return accessservices.NewEntitlementsService(policy, events)
}

func TestGetUsageShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := controllers.NewEntitlementController(usageControllerSvc(t))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/premium/usage", nil)
	c.Set("userID", "user_1")

	ctl.GetUsage(c)
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"plan":"free"`)
	assert.Contains(t, body, `"counters"`)
	assert.Contains(t, body, `"costs"`)
	assert.Contains(t, body, `"history"`)
	assert.Contains(t, body, "exam_ai_review")
	assert.Contains(t, body, `"total":1`)
}

func TestGetUsagePaginationPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
	}, nil).Once()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}, nil).Once()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Once()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, 5, 5).
		Return([]*models.UsageEvent{}, nil).Once()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()
	policy := accessservices.NewPolicyService(fakeSubs{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	ctl := controllers.NewEntitlementController(accessservices.NewEntitlementsService(policy, events))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/premium/usage?page=2&pageSize=5", nil)
	c.Set("userID", "user_1")

	ctl.GetUsage(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"page":2`)
}

func TestGetUsageRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// No expectations: the auth gate stops the request before the service
	// is touched, and bare mocks assert exactly that (zero calls).
	policy := accessservices.NewPolicyService(fakeSubs{}, repomocks.NewMockUserTrialRepository(t),
		repomocks.NewMockUsageCounterRepository(t), repomocks.NewMockPlanFeatureLimitRepository(t),
		repomocks.NewMockActionCreditCostRepository(t), repomocks.NewMockTrialSettingRepository(t),
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	ctl := controllers.NewEntitlementController(accessservices.NewEntitlementsService(policy,
		repomocks.NewMockUsageEventRepository(t)))
	r := gin.New()
	r.GET("/premium/usage", middleware.RequireAuth(), ctl.GetUsage)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/premium/usage", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
