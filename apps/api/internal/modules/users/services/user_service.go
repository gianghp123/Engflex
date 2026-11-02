package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/identity"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/users/dtos/requests"
)

// UserService lists and manages roles for Clerk accounts. Clerk is the
// only store — there is no local users table — so every method is a thin
// wrapper over the Clerk Backend API. Premium status is the one field that
// does come from a local table, filled in via subscriptionLookup.
type UserService interface {
	List(ctx context.Context, req requests.ListUsers) ([]*models.DirectoryUser, int64, error)
	UpdateRole(ctx context.Context, callerUserID, id, role string) (*models.DirectoryUser, error)
	ListAttempts(ctx context.Context, userID string, req requests.ListUserAttempts) ([]*models.Attempt, int64, error)
}

// clerkUserClient is the slice of *user.Client this service actually calls
// — narrowed to an interface so tests can fake the Clerk Backend API
// instead of hitting the network.
type clerkUserClient interface {
	List(ctx context.Context, params *user.ListParams) (*clerk.UserList, error)
	Update(ctx context.Context, id string, params *user.UpdateParams) (*clerk.User, error)
}

// subscriptionLookup is the minimal surface List needs to enrich a user
// row with premium status — narrowed here (not imported as the full
// services.SubscriptionService) so this package only depends on the shape
// it actually calls.
type subscriptionLookup interface {
	Enrich(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error)
}

// Learning read interface keeps the directory independent of attempt mutations.
type UserAttemptsReader interface {
	CountByUser(ctx context.Context, userID string, kind *enums.AttemptType) (int64, error)
	ListByUser(ctx context.Context, userID string, kind *enums.AttemptType, limit, offset int) ([]*models.Attempt, error)
}

type userService struct {
	client   clerkUserClient
	subs     subscriptionLookup
	attempts UserAttemptsReader
}

// NewUserService builds the service over the Clerk Go SDK's default
// backend (configured once in main.go via clerk.SetKey).
func NewUserService(subs subscriptionLookup, attempts UserAttemptsReader) UserService {
	return &userService{client: &user.Client{Backend: clerk.GetBackend()}, subs: subs, attempts: attempts}
}

// NewUserServiceWithClient is the test seam: builds the service over a
// fake clerkUserClient instead of the real Clerk Backend API.
func NewUserServiceWithClient(c clerkUserClient, subs subscriptionLookup, attempts UserAttemptsReader) UserService {
	return &userService{client: c, subs: subs, attempts: attempts}
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

func (s *userService) List(ctx context.Context, req requests.ListUsers) ([]*models.DirectoryUser, int64, error) {
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
	out, err := s.subs.Enrich(ctx, identity.DirectoryUsers(list.Users))
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "enrich directory failed", appErr)
		return nil, 0, appErr
	}

	slog.InfoContext(ctx, "users listed", "count", len(out), "total", list.TotalCount)
	return out, list.TotalCount, nil
}

// UpdateRole sets the target user's public_metadata.role. An admin cannot
// demote themselves — the console would have no way back in without a
// second admin or direct Clerk Dashboard access.
func (s *userService) UpdateRole(ctx context.Context, callerUserID, id, role string) (*models.DirectoryUser, error) {
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
	out, err := s.subs.Enrich(ctx, []*models.DirectoryUser{identity.DirectoryUser(updated)})
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "enrich updated user failed", appErr)
		return nil, appErr
	}
	return out[0], nil
}

func (s *userService) ListAttempts(ctx context.Context, userID string, req requests.ListUserAttempts) ([]*models.Attempt, int64, error) {
	if s.attempts == nil {
		appErr := common.Internal()
		logger.Report(ctx, "attempts repository missing in user service", appErr, "userId", userID)
		return nil, 0, appErr
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
	}
	offset := (page - 1) * pageSize

	total, err := s.attempts.CountByUser(ctx, userID, req.Type)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "count user attempts failed", appErr, "userId", userID)
		return nil, 0, appErr
	}

	rows, err := s.attempts.ListByUser(ctx, userID, req.Type, pageSize, offset)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "list user attempts failed", appErr, "userId", userID)
		return nil, 0, appErr
	}

	slog.InfoContext(ctx, "user attempts listed", "userId", userID, "count", len(rows), "total", total)
	return rows, total, nil
}
