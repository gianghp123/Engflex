package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/payments/services"
)

func setupSubscription(t *testing.T) (services.SubscriptionService, *mocks.MockSubscriptionRepository, *mocks.MockPlanRepository) {
	t.Helper()
	mockRepo := mocks.NewMockSubscriptionRepository(t)
	mockPlans := mocks.NewMockPlanRepository(t)
	svc := services.NewSubscriptionService(mockRepo, mockPlans)
	return svc, mockRepo, mockPlans
}

func setupPlan(t *testing.T) (services.PlanService, *mocks.MockPlanRepository) {
	t.Helper()
	mockRepo := mocks.NewMockPlanRepository(t)
	svc := services.NewPlanService(mockRepo)
	return svc, mockRepo
}

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}
