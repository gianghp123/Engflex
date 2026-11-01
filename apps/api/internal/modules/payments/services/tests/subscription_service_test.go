package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/payments/dtos/requests"
)

func TestSubscriptionService_IsPremium(t *testing.T) {
	t.Run("active unexpired subscription reports premium", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		expires := time.Now().Add(24 * time.Hour)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").
			Return(&models.Subscription{ID: "sub_1", UserID: "user_1", Status: enums.SubscriptionStatusActive, ExpiresAt: &expires}, nil)

		isPremium, err := svc.IsPremium(context.Background(), "user_1")

		require.NoError(t, err)
		assert.True(t, isPremium)
	})

	t.Run("no active subscription reports non-premium, not an error", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound)

		isPremium, err := svc.IsPremium(context.Background(), "user_1")

		require.NoError(t, err)
		assert.False(t, isPremium)
	})
}

func TestSubscriptionService_GetActive(t *testing.T) {
	t.Run("returns the active subscription", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		expires := time.Now().Add(24 * time.Hour)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").
			Return(&models.Subscription{ID: "sub_1", UserID: "user_1", Status: enums.SubscriptionStatusActive, ExpiresAt: &expires}, nil)

		sub, err := svc.GetActive(context.Background(), "user_1")

		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, "sub_1", sub.ID)
	})

	t.Run("no active subscription returns nil, not an error", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound)

		sub, err := svc.GetActive(context.Background(), "user_1")

		require.NoError(t, err)
		assert.Nil(t, sub)
	})
}

func TestSubscriptionService_Grant(t *testing.T) {
	svc, mockRepo, mockPlans := setupSubscription(t)
	mockPlans.On("GetByPlan", mock.Anything, enums.SubscriptionPlanOneMonth).
		Return(&models.Plan{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, Currency: "VND", IsActive: true}, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Subscription) bool {
		return m.UserID == "user_1" &&
			m.Plan == enums.SubscriptionPlanOneMonth &&
			m.Status == enums.SubscriptionStatusActive &&
			m.StartedAt != nil &&
			m.ExpiresAt != nil &&
			m.ExpiresAt.After(*m.StartedAt) &&
			m.Amount > 0
	})).Return(nil)

	res, err := svc.Grant(context.Background(), "user_1", requests.GrantSubscription{Plan: enums.SubscriptionPlanOneMonth})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, enums.SubscriptionStatusActive, res.Status)
}

func TestSubscriptionService_Grant_UnknownPlan(t *testing.T) {
	svc, mockRepo, mockPlans := setupSubscription(t)
	mockPlans.On("GetByPlan", mock.Anything, enums.SubscriptionPlan("bogus")).
		Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Grant(context.Background(), "user_1", requests.GrantSubscription{Plan: enums.SubscriptionPlan("bogus")})

	requireAppError(t, err, http.StatusBadRequest)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSubscriptionService_Revoke(t *testing.T) {
	t.Run("cancels the active subscription", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		expires := time.Now().Add(24 * time.Hour)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").
			Return(&models.Subscription{ID: "sub_1", UserID: "user_1", Status: enums.SubscriptionStatusActive, ExpiresAt: &expires}, nil)
		mockRepo.On("UpdateStatus", mock.Anything, "sub_1", enums.SubscriptionStatusCancelled, (*time.Time)(nil), (*time.Time)(nil)).
			Return(nil)

		err := svc.Revoke(context.Background(), "user_1")

		require.NoError(t, err)
	})

	t.Run("no active subscription is a no-op, not an error", func(t *testing.T) {
		svc, mockRepo, _ := setupSubscription(t)
		mockRepo.On("GetActiveByUserID", mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound)

		err := svc.Revoke(context.Background(), "user_1")

		require.NoError(t, err)
		mockRepo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestSubscriptionService_ListAll(t *testing.T) {
	svc, mockRepo, _ := setupSubscription(t)
	mockRepo.On("ListAll", mock.Anything, enums.SubscriptionStatus(""), enums.SubscriptionPlan(""), 20, 0).
		Return([]*models.Subscription{{ID: "sub_1", UserID: "user_1"}}, nil)
	mockRepo.On("CountAll", mock.Anything, enums.SubscriptionStatus(""), enums.SubscriptionPlan("")).
		Return(int64(1), nil)

	items, total, err := svc.ListAll(context.Background(), "", "", 20, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
}
