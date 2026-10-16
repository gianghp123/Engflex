package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
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
	svc := services.NewUserServiceWithClient(fake)

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
	svc := services.NewUserServiceWithClient(fake)

	items, _, err := svc.List(context.Background(), requests.ListUsers{ListParams: common.ListParams{Page: 1, PageSize: 20}})

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "user", items[0].Role)
}

func TestUserService_List_ClerkErrorMapsTo503(t *testing.T) {
	fake := fakeClerkClient{
		listFn: func(ctx context.Context, params *user.ListParams) (*clerk.UserList, error) {
			return nil, assert.AnError
		},
	}
	svc := services.NewUserServiceWithClient(fake)

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
	svc := services.NewUserServiceWithClient(fake)

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
	svc := services.NewUserServiceWithClient(fake)

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
	svc := services.NewUserServiceWithClient(fake)

	_, err := svc.UpdateRole(context.Background(), "admin_1", "missing", "admin")

	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Status)
}
