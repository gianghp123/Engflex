package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
)

func TestSubscriptionService_IsPremium(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockSubscriptionRepository)
		wantErr    bool
		wantStatus int
		wantValue  bool
	}{
		{
			name: "active subscription found: premium true",
			setupMock: func(m *mocks.MockSubscriptionRepository) {
				m.On("GetActiveByUserID", mock.Anything, "user-1").
					Return(&models.Subscription{ID: "sub-1", UserID: "user-1"}, nil)
			},
			wantValue: true,
		},
		{
			name: "no active row: premium false, not an error",
			setupMock: func(m *mocks.MockSubscriptionRepository) {
				m.On("GetActiveByUserID", mock.Anything, "user-1").
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantValue: false,
		},
		{
			name: "db error: 500",
			setupMock: func(m *mocks.MockSubscriptionRepository) {
				m.On("GetActiveByUserID", mock.Anything, "user-1").
					Return(nil, errors.New("db connection failure"))
			},
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setup(t)
			tt.setupMock(mockRepo)

			got, err := svc.IsPremium(context.Background(), "user-1")

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestSubscriptionService_Grant(t *testing.T) {
	expires := time.Now().Add(30 * 24 * time.Hour)
	tests := []struct {
		name       string
		expiresAt  *time.Time
		setupMock  func(m *mocks.MockSubscriptionRepository)
		wantErr    bool
		wantStatus int
	}{
		{
			name:      "success: grants premium",
			expiresAt: &expires,
			setupMock: func(m *mocks.MockSubscriptionRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						sub := args.Get(1).(*models.Subscription)
						sub.ID = "sub-1"
					}).
					Return(nil)
			},
		},
		{
			name:      "internal: db error returns 500",
			expiresAt: &expires,
			setupMock: func(m *mocks.MockSubscriptionRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("db connection failure"))
			},
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setup(t)
			tt.setupMock(mockRepo)

			got, err := svc.Grant(context.Background(), "user-1", "premium_monthly", tt.expiresAt)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "sub-1", got.ID)
				assert.Equal(t, "premium_monthly", got.Plan)
			}
		})
	}
}
