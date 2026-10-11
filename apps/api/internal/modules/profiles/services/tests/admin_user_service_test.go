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

	"engflex-api/internal/clerkusers"
	clerkmocks "engflex-api/internal/clerkusers/mocks"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/profiles/services"
)

func setupAdminUser(t *testing.T) (services.AdminUserService, *clerkmocks.MockClient, *mocks.MockUserProfileRepository, *mocks.MockSubscriptionRepository) {
	clerkMock := clerkmocks.NewMockClient(t)
	profileMock := mocks.NewMockUserProfileRepository(t)
	subRepoMock := mocks.NewMockSubscriptionRepository(t)
	subSvc := services.NewSubscriptionService(subRepoMock)
	svc := services.NewAdminUserService(clerkMock, profileMock, subSvc)
	return svc, clerkMock, profileMock, subRepoMock
}

func TestAdminUserService_List(t *testing.T) {
	createdAt := time.Now()
	tests := []struct {
		name       string
		setupMock  func(c *clerkmocks.MockClient, s *mocks.MockSubscriptionRepository)
		wantErr    bool
		wantStatus int
		wantCount  int
		wantTotal  int64
	}{
		{
			name: "success: lists users with premium flag",
			setupMock: func(c *clerkmocks.MockClient, s *mocks.MockSubscriptionRepository) {
				c.On("List", mock.Anything, 20, 0).
					Return([]*clerkusers.User{
						{ID: "user-1", Email: "a@test.com", Role: "user", CreatedAt: createdAt},
					}, int64(1), nil)
				s.On("GetActiveByUserID", mock.Anything, "user-1").
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "clerk list failure: 500",
			setupMock: func(c *clerkmocks.MockClient, s *mocks.MockSubscriptionRepository) {
				c.On("List", mock.Anything, 20, 0).
					Return(nil, int64(0), errors.New("clerk api down"))
			},
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, clerkMock, _, subRepoMock := setupAdminUser(t)
			tt.setupMock(clerkMock, subRepoMock)

			got, total, err := svc.List(context.Background(), 20, 0)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestAdminUserService_GetByID(t *testing.T) {
	createdAt := time.Now()
	tests := []struct {
		name       string
		setupMock  func(c *clerkmocks.MockClient, p *mocks.MockUserProfileRepository, s *mocks.MockSubscriptionRepository)
		wantErr    bool
		wantStatus int
		wantEmail  string
	}{
		{
			name: "success: detail with no onboarding profile yet",
			setupMock: func(c *clerkmocks.MockClient, p *mocks.MockUserProfileRepository, s *mocks.MockSubscriptionRepository) {
				c.On("Get", mock.Anything, "user-1").
					Return(&clerkusers.User{ID: "user-1", Email: "a@test.com", CreatedAt: createdAt}, nil)
				s.On("GetActiveByUserID", mock.Anything, "user-1").
					Return(nil, gorm.ErrRecordNotFound)
				p.On("GetByUserID", mock.Anything, "user-1").
					Return(nil, gorm.ErrRecordNotFound)
				s.On("ListByUserID", mock.Anything, "user-1").
					Return([]*models.Subscription{}, nil)
			},
			wantEmail: "a@test.com",
		},
		{
			name: "clerk user not found: 404",
			setupMock: func(c *clerkmocks.MockClient, p *mocks.MockUserProfileRepository, s *mocks.MockSubscriptionRepository) {
				c.On("Get", mock.Anything, "user-missing").
					Return(nil, errors.New("not found"))
			},
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, clerkMock, profileMock, subRepoMock := setupAdminUser(t)
			tt.setupMock(clerkMock, profileMock, subRepoMock)

			id := "user-1"
			if tt.wantErr {
				id = "user-missing"
			}
			got, err := svc.GetByID(context.Background(), id)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantEmail, got.Email)
				assert.Nil(t, got.Profile)
				assert.False(t, got.Premium)
			}
		})
	}
}
