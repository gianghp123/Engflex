package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/services"
)

func setupPremiumFeature(t *testing.T) (services.PremiumFeatureService, *mocks.MockPremiumFeatureRepository) {
	t.Helper()
	mockRepo := mocks.NewMockPremiumFeatureRepository(t)
	svc := services.NewPremiumFeatureService(mockRepo)
	return svc, mockRepo
}

func TestPremiumFeatureService_List(t *testing.T) {
	svc, mockRepo := setupPremiumFeature(t)
	mockRepo.On("List", mock.Anything, 20, 0).Return([]*models.PremiumFeature{
		{ID: "f1", Key: "voice_conversations", Label: "Voice conversations", IsActive: true},
	}, nil)
	mockRepo.On("Count", mock.Anything).Return(int64(1), nil)

	items, total, err := svc.List(context.Background(), 20, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "voice_conversations", items[0].Key)
}

func TestPremiumFeatureService_Create(t *testing.T) {
	svc, mockRepo := setupPremiumFeature(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *models.PremiumFeature) bool {
		return m.Key == "voice_conversations" && m.Label == "Voice conversations"
	})).Return(nil)

	created, err := svc.Create(context.Background(), requests.CreatePremiumFeature{
		Key: "voice_conversations", Label: "Voice conversations", IsActive: true,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "voice_conversations", created.Key)
}

func TestPremiumFeatureService_Update(t *testing.T) {
	svc, mockRepo := setupPremiumFeature(t)
	mockRepo.On("GetByID", mock.Anything, "f1").
		Return(&models.PremiumFeature{ID: "f1", Key: "voice_conversations", Label: "old", IsActive: true}, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.PremiumFeature) bool {
		return m.ID == "f1" && m.Label == "new label" && !m.IsActive
	})).Return(nil)

	updated, err := svc.Update(context.Background(), "f1", requests.UpdatePremiumFeature{
		Label: "new label", IsActive: false,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "new label", updated.Label)
	assert.False(t, updated.IsActive)
}

func TestPremiumFeatureService_Delete(t *testing.T) {
	svc, mockRepo := setupPremiumFeature(t)
	mockRepo.On("Delete", mock.Anything, "f1").Return(nil)

	err := svc.Delete(context.Background(), "f1")

	require.NoError(t, err)
}
