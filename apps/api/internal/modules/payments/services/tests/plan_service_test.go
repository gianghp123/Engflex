package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/payments/dtos/requests"
)

func TestPlanService_List(t *testing.T) {
	svc, mockRepo := setupPlan(t)
	mockRepo.On("List", mock.Anything).Return([]*models.Plan{
		{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, IsActive: true},
		{Plan: enums.SubscriptionPlanThreeMonth, Months: 3, Amount: 249_000, IsActive: false},
	}, nil)

	items, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestPlanService_ListActive_FiltersInactive(t *testing.T) {
	svc, mockRepo := setupPlan(t)
	mockRepo.On("List", mock.Anything).Return([]*models.Plan{
		{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, IsActive: true},
		{Plan: enums.SubscriptionPlanThreeMonth, Months: 3, Amount: 249_000, IsActive: false},
	}, nil)

	items, err := svc.ListActive(context.Background())

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, enums.SubscriptionPlanOneMonth, items[0].Plan)
}

func TestPlanService_Update(t *testing.T) {
	svc, mockRepo := setupPlan(t)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Plan) bool {
		return m.Plan == enums.SubscriptionPlanOneMonth && m.Amount == 149_000
	})).Return(nil)
	mockRepo.On("GetByPlan", mock.Anything, enums.SubscriptionPlanOneMonth).
		Return(&models.Plan{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 149_000, Currency: "VND", IsActive: true}, nil)

	updated, err := svc.Update(context.Background(), enums.SubscriptionPlanOneMonth, requests.UpdatePlan{
		Months: 1, Amount: 149_000, Currency: "VND", IsActive: true,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(149_000), updated.Amount)
}
