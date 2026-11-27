// Public pricing endpoint (spec §10): reachable without auth and cacheable
// for 30s at the edge.
package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/config"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/controllers"
	accessservices "engflex-api/internal/modules/access/services"
)

func pricingControllerSvc(t *testing.T) *accessservices.PricingService {
	t.Helper()
	plans := repomocks.NewMockPlanRepository(t)
	plans.EXPECT().ListActive(mock.Anything).Return([]*models.Plan{
		{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99000, Currency: "VND", IsActive: true},
	}, nil).Maybe()
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 30, Period: enums.UsagePeriodTotal},
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 7, Period: enums.UsagePeriodTotal},
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 100, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 60, Period: enums.UsagePeriodMonth},
	}, nil).Maybe()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}, nil).Maybe()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Maybe()
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	policy := accessservices.NewPolicyService(fakeSubs{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	return accessservices.NewPricingService(plans, policy)
}

func TestPricingEndpointIsPublicAndCacheable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := controllers.NewPricingController(pricingControllerSvc(t))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pricing", nil)
	ctl.Get(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "public, max-age=30", w.Header().Get("Cache-Control"))
	assert.Contains(t, w.Body.String(), `"free"`)
}

func TestPricingEndpointRequiresNoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := controllers.NewPricingController(pricingControllerSvc(t))
	r := gin.New()
	r.GET("/pricing", ctl.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pricing", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "no auth middleware: the route is public")
}
