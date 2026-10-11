package services

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"engflex-api/internal/clerkusers"
	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/profiles/dtos/responses"
	"engflex-api/internal/utils"
)

// AdminUserService is the read-only admin view over user accounts: identity
// lives in Clerk, onboarding state and premium status live in this DB.
// Account mutation (ban/lock) is deliberately not exposed here — that stays
// a manual Clerk Dashboard action.
type AdminUserService interface {
	List(ctx context.Context, limit, offset int) ([]*responses.AdminUserSummary, int64, error)
	GetByID(ctx context.Context, userID string) (*responses.AdminUserDetail, error)
}

type adminUserService struct {
	clerk         clerkusers.Client
	profiles      repositories.UserProfileRepository
	subscriptions SubscriptionService
}

// NewAdminUserService builds the service over the Clerk client, the profile
// repository, and the subscription service (reused for the premium check).
func NewAdminUserService(clerk clerkusers.Client, profiles repositories.UserProfileRepository, subscriptions SubscriptionService) AdminUserService {
	return &adminUserService{clerk: clerk, profiles: profiles, subscriptions: subscriptions}
}

func (s *adminUserService) List(ctx context.Context, limit, offset int) ([]*responses.AdminUserSummary, int64, error) {
	users, total, err := s.clerk.List(ctx, limit, offset)
	if err != nil {
		logger.Report(ctx, "list clerk users failed", common.Internal())
		return nil, 0, common.Internal()
	}
	out := make([]*responses.AdminUserSummary, 0, len(users))
	for _, u := range users {
		premium, err := s.subscriptions.IsPremium(ctx, u.ID)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, &responses.AdminUserSummary{
			ID:        u.ID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
			Premium:   premium,
			CreatedAt: u.CreatedAt,
		})
	}
	slog.InfoContext(ctx, "admin users listed", "count", len(out), "total", total)
	return out, total, nil
}

func (s *adminUserService) GetByID(ctx context.Context, userID string) (*responses.AdminUserDetail, error) {
	u, err := s.clerk.Get(ctx, userID)
	if err != nil {
		appErr := common.NotFound("user not found")
		logger.Report(ctx, "get clerk user failed", appErr, "userID", userID)
		return nil, appErr
	}

	premium, err := s.subscriptions.IsPremium(ctx, userID)
	if err != nil {
		return nil, err
	}

	detail := &responses.AdminUserDetail{
		AdminUserSummary: responses.AdminUserSummary{
			ID:        u.ID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
			Premium:   premium,
			CreatedAt: u.CreatedAt,
		},
		Subscriptions: []responses.Subscription{},
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err == nil {
		var dto responses.Profile
		_ = utils.Map(&dto, profile)
		detail.Profile = &dto
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(err, "user profile")
		logger.Report(ctx, "get user profile failed", appErr, "userID", userID)
		return nil, appErr
	}

	subs, err := s.subscriptions.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = utils.MapSlice(&detail.Subscriptions, subs)

	slog.InfoContext(ctx, "admin user detail fetched", "userID", userID)
	return detail, nil
}
