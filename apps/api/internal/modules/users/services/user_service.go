package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/users/dtos/requests"
	"engflex-api/internal/modules/users/dtos/responses"
)

// UserService lists and manages roles for Clerk accounts. Clerk is the
// only store — there is no local users table — so every method is a thin
// wrapper over the Clerk Backend API.
type UserService interface {
	List(ctx context.Context, req requests.ListUsers) ([]*responses.User, int64, error)
	UpdateRole(ctx context.Context, callerUserID, id, role string) (*responses.User, error)
}

// clerkUserClient is the slice of *user.Client this service actually calls
// — narrowed to an interface so tests can fake the Clerk Backend API
// instead of hitting the network.
type clerkUserClient interface {
	List(ctx context.Context, params *user.ListParams) (*clerk.UserList, error)
	Update(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error)
}

type userService struct {
	client clerkUserClient
}

// NewUserService builds the service over the Clerk Go SDK's default
// backend (configured once in main.go via clerk.SetKey).
func NewUserService() UserService {
	return &userService{client: &user.Client{Backend: clerk.GetBackend()}}
}

// NewUserServiceWithClient is the test seam: builds the service over a
// fake clerkUserClient instead of the real Clerk Backend API.
func NewUserServiceWithClient(c clerkUserClient) UserService {
	return &userService{client: c}
}

// fromClerkError maps a Clerk API error to the house AppError shape by its
// HTTP status, so a bad user id 404s the same way a bad DB id does.
func fromClerkError(err error, resource string) *common.AppError {
	var apiErr *clerk.APIErrorResponse
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case http.StatusNotFound:
			return common.NotFound(resource + " not found")
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			return common.BadRequest(apiErr.Error())
		}
	}
	return common.ServiceUnavailable("user directory unavailable")
}

type publicMetadata struct {
	Role string `json:"role,omitempty"`
}

func toUserResponse(u *clerk.User) *responses.User {
	email := ""
	for _, addr := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && addr.ID == *u.PrimaryEmailAddressID {
			email = addr.EmailAddress
			break
		}
	}
	if email == "" && len(u.EmailAddresses) > 0 {
		email = u.EmailAddresses[0].EmailAddress
	}

	var meta publicMetadata
	_ = json.Unmarshal(u.PublicMetadata, &meta)
	role := meta.Role
	if role == "" {
		role = string(enums.UserRoleUser)
	}

	firstName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	lastName := ""
	if u.LastName != nil {
		lastName = *u.LastName
	}

	var lastActive *time.Time
	if u.LastActiveAt != nil {
		t := time.UnixMilli(*u.LastActiveAt)
		lastActive = &t
	}

	return &responses.User{
		ID:           u.ID,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		Banned:       u.Banned,
		CreatedAt:    time.UnixMilli(u.CreatedAt),
		LastActiveAt: lastActive,
	}
}

func (s *userService) List(ctx context.Context, req requests.ListUsers) ([]*responses.User, int64, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
	}
	limit := int64(pageSize)
	offset := int64((page - 1) * pageSize)
	params := &user.ListParams{
		ListParams: clerk.ListParams{Limit: &limit, Offset: &offset},
	}
	if req.Search != "" {
		params.Query = &req.Search
	}

	list, err := s.client.List(ctx, params)
	if err != nil {
		appErr := fromClerkError(err, "user")
		logger.Report(ctx, "list clerk users failed", appErr)
		return nil, 0, appErr
	}
	out := make([]*responses.User, 0, len(list.Users))
	for _, u := range list.Users {
		out = append(out, toUserResponse(u))
	}
	slog.InfoContext(ctx, "users listed", "count", len(out), "total", list.TotalCount)
	return out, list.TotalCount, nil
}

// UpdateRole sets the target user's public_metadata.role. An admin cannot
// demote themselves — the console would have no way back in without a
// second admin or direct Clerk Dashboard access.
func (s *userService) UpdateRole(ctx context.Context, callerUserID, id, role string) (*responses.User, error) {
	if callerUserID == id && role != string(enums.UserRoleAdmin) {
		return nil, common.BadRequest("you cannot change your own admin role")
	}
	meta, err := json.Marshal(publicMetadata{Role: role})
	if err != nil {
		return nil, common.Internal()
	}
	raw := json.RawMessage(meta)
	updated, err := s.client.Update(ctx, id, &user.UpdateParams{PublicMetadata: &raw})
	if err != nil {
		appErr := fromClerkError(err, "user")
		logger.Report(ctx, "update user role failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user role updated", "id", id, "role", role)
	return toUserResponse(updated), nil
}
