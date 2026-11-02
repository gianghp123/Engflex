package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/users/dtos/requests"
	"engflex-api/internal/modules/users/services"
)

// fakeClerkClient implements the service's clerkUserClient seam without
// hitting the network.
type fakeClerkClient struct {
	listFn   func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error)
	updateFn func(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error)
}

func (f fakeClerkClient) List(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
	return f.listFn(ctx, params)
}

func (f fakeClerkClient) Update(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error) {
	return f.updateFn(ctx, id, params)
}

// fakeSubscriptionLookup implements the service's subscriptionLookup seam.
// Defaults to "no active subscription" — these tests exercise the
// Clerk-facing behavior, not premium enrichment.
type fakeSubscriptionLookup struct{}

func (fakeSubscriptionLookup) Enrich(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error) {
	return users, nil
}

type fakeSubscriptionLookupFn func(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error)

func (f fakeSubscriptionLookupFn) Enrich(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error) {
	return f(ctx, users)
}

func newUserService(t *testing.T, c fakeClerkClient) services.UserService {
	return services.NewUserServiceWithClient(c, fakeSubscriptionLookup{}, mocks.NewMockAttemptRepository(t))
}

func strPtr(s string) *string { return &s }

func TestUserService_List_MapsPrimaryEmailAndRole(t *testing.T) {
	meta, _ := json.Marshal(map[string]string{"role": "admin"})
	primaryID := "email_1"
	fake := fakeClerkClient{
		listFn: func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
			return &clerk.UserList{
				Users: []*clerk.User{
					{
						ID:                    "user_1",
						FirstName:             strPtr("Ada"),
						LastName:              strPtr("Lovelace"),
						PrimaryEmailAddressID: &primaryID,
						EmailAddresses: []*clerk.EmailAddress{
							{ID: "email_0", EmailAddress: "old@example.com"},
							{ID: primaryID, EmailAddress: "ada@example.com"},
						},
						PublicMetadata: meta,
						CreatedAt:      1700000000000,
					},
				},
				TotalCount: 1,
			}, nil
		},
	}
	svc := newUserService(t, fake)

	items, total, err := svc.List(context.Background(), requests.ListUsers{ListParams: common.ListParams{Page: 1, PageSize: 20}})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "ada@example.com", items[0].Email)
	assert.Equal(t, "admin", items[0].Role)
	assert.Equal(t, "Ada", items[0].FirstName)
}

func TestUserService_List_DefaultsRoleToUser(t *testing.T) {
	fake := fakeClerkClient{
		listFn: func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
			return &clerk.UserList{
				Users: []*clerk.User{
					{ID: "user_2", PublicMetadata: json.RawMessage(`{}`)},
				},
				TotalCount: 1,
			}, nil
		},
	}
	svc := newUserService(t, fake)

	items, _, err := svc.List(context.Background(), requests.ListUsers{ListParams: common.ListParams{Page: 1, PageSize: 20}})

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "user", items[0].Role)
}

func TestUserService_List_EnrichesPremiumStatus(t *testing.T) {
	fake := fakeClerkClient{
		listFn: func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
			return &clerk.UserList{
				Users:      []*clerk.User{{ID: "user_3", PublicMetadata: json.RawMessage(`{}`)}},
				TotalCount: 1,
			}, nil
		},
	}
	expires := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	subs := fakeSubscriptionLookupFn(func(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error) {
		users[0].Subscription = &models.Subscription{ExpiresAt: &expires}
		return users, nil
	})
	svc := services.NewUserServiceWithClient(fake, subs, mocks.NewMockAttemptRepository(t))

	items, _, err := svc.List(context.Background(), requests.ListUsers{ListParams: common.ListParams{Page: 1, PageSize: 20}})

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].Subscription)
	assert.Equal(t, &expires, items[0].Subscription.ExpiresAt)
}

func TestUserService_List_ClerkErrorMapsTo503(t *testing.T) {
	fake := fakeClerkClient{
		listFn: func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
			return nil, assert.AnError
		},
	}
	svc := newUserService(t, fake)

	items, total, err := svc.List(context.Background(), requests.ListUsers{ListParams: common.ListParams{Page: 1, PageSize: 20}})

	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusServiceUnavailable, appErr.Status)
	assert.Nil(t, items)
	assert.Zero(t, total)
}

// An admin cannot demote themselves — the console would have no way back
// in without a second admin or direct Clerk Dashboard access.
func TestUserService_UpdateRole_RejectsSelfDemotion(t *testing.T) {
	fake := fakeClerkClient{
		updateFn: func(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error) {
			t.Fatal("Update should not be called for a self-demotion")
			return nil, nil
		},
	}
	svc := newUserService(t, fake)

	_, err := svc.UpdateRole(context.Background(), "user_1", "user_1", "user")

	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Status)
}

// Promoting yourself to the role you already hold (no-op) and changing
// another user's role both go through to Clerk.
func TestUserService_UpdateRole_AllowsOtherUsers(t *testing.T) {
	var gotRole string
	fake := fakeClerkClient{
		updateFn: func(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error) {
			var meta map[string]string
			_ = json.Unmarshal(*params.PublicMetadata, &meta)
			gotRole = meta["role"]
			return &clerk.User{ID: id, PublicMetadata: *params.PublicMetadata}, nil
		},
	}
	svc := newUserService(t, fake)

	updated, err := svc.UpdateRole(context.Background(), "admin_1", "user_2", "admin")

	require.NoError(t, err)
	assert.Equal(t, "admin", gotRole)
	assert.Equal(t, "admin", updated.Role)
}

func TestUserService_UpdateRole_NotFoundMapsTo404(t *testing.T) {
	fake := fakeClerkClient{
		updateFn: func(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error) {
			return nil, &clerk.APIErrorResponse{HTTPStatusCode: http.StatusNotFound}
		},
	}
	svc := newUserService(t, fake)

	_, err := svc.UpdateRole(context.Background(), "admin_1", "missing", "admin")

	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Status)
}

func TestUserService_ListAttempts(t *testing.T) {
	lessonType := enums.AttemptTypeLesson
	score := 85.0
	dur := 120
	now := time.Now().UTC()

	tests := []struct {
		name          string
		userID        string
		req           requests.ListUserAttempts
		setupMock     func(m *mocks.MockAttemptRepository)
		expectedCount int
		expectedTotal int64
		expectErr     bool
	}{
		{
			name:   "success without type filter",
			userID: "user_123",
			req: requests.ListUserAttempts{
				ListParams: common.ListParams{Page: 1, PageSize: 10},
			},
			setupMock: func(m *mocks.MockAttemptRepository) {
				m.EXPECT().CountByUser(mock.Anything, "user_123", (*enums.AttemptType)(nil)).Return(int64(1), nil).Once()
				m.EXPECT().ListByUser(mock.Anything, "user_123", (*enums.AttemptType)(nil), 10, 0).Return([]*models.Attempt{
					{
						ID:          "att_1",
						UserID:      "user_123",
						Type:        enums.AttemptTypeLesson,
						Status:      enums.AttemptStatusCompleted,
						Score:       &score,
						DurationSec: &dur,
						CreatedAt:   now,
						Lesson: &models.Lesson{
							ID:    "less_1",
							Title: "Greetings Lesson",
						},
					},
				}, nil).Once()
			},
			expectedCount: 1,
			expectedTotal: 1,
			expectErr:     false,
		},
		{
			name:   "success with type filter",
			userID: "user_123",
			req: requests.ListUserAttempts{
				ListParams: common.ListParams{Page: 1, PageSize: 10},
				Type:       &lessonType,
			},
			setupMock: func(m *mocks.MockAttemptRepository) {
				m.EXPECT().CountByUser(mock.Anything, "user_123", &lessonType).Return(int64(2), nil).Once()
				m.EXPECT().ListByUser(mock.Anything, "user_123", &lessonType, 10, 0).Return([]*models.Attempt{
					{
						ID:        "att_2",
						UserID:    "user_123",
						Type:      enums.AttemptTypeLesson,
						Status:    enums.AttemptStatusCompleted,
						CreatedAt: now,
					},
				}, nil).Once()
			},
			expectedCount: 1,
			expectedTotal: 2,
			expectErr:     false,
		},
		{
			name:   "count failure returns error",
			userID: "user_123",
			req: requests.ListUserAttempts{
				ListParams: common.ListParams{Page: 1, PageSize: 10},
			},
			setupMock: func(m *mocks.MockAttemptRepository) {
				m.EXPECT().CountByUser(mock.Anything, "user_123", (*enums.AttemptType)(nil)).Return(int64(0), errors.New("db error")).Once()
			},
			expectErr: true,
		},
		{
			name:   "list failure returns error",
			userID: "user_123",
			req: requests.ListUserAttempts{
				ListParams: common.ListParams{Page: 1, PageSize: 10},
			},
			setupMock: func(m *mocks.MockAttemptRepository) {
				m.EXPECT().CountByUser(mock.Anything, "user_123", (*enums.AttemptType)(nil)).Return(int64(1), nil).Once()
				m.EXPECT().ListByUser(mock.Anything, "user_123", (*enums.AttemptType)(nil), 10, 0).Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAttempts := mocks.NewMockAttemptRepository(t)
			tt.setupMock(mockAttempts)
			svc := services.NewUserServiceWithClient(fakeClerkClient{}, fakeSubscriptionLookup{}, mockAttempts)

			items, total, err := svc.ListAttempts(context.Background(), tt.userID, tt.req)
			if tt.expectErr {
				require.Error(t, err)
				var appErr *common.AppError
				require.ErrorAs(t, err, &appErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, items, tt.expectedCount)
			if tt.expectedCount > 0 && items[0].Lesson != nil {
				assert.Equal(t, "Greetings Lesson", items[0].Lesson.Title)
			}
		})
	}
}
