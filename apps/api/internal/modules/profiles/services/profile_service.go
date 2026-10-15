// Package services holds the profiles-domain business logic.
package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/profiles/dtos/requests"
)

// ProfileService owns the user profile and onboarding completion.
type ProfileService struct {
	profiles repositories.UserProfileRepository
}

// NewProfileService takes interfaces, never concretes.
func NewProfileService(profiles repositories.UserProfileRepository) *ProfileService {
	return &ProfileService{profiles: profiles}
}

// GetProfile returns the caller's profile, or a 404 AppError when absent.
func (s *ProfileService) GetProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "profile")
		logger.Report(ctx, "profile read failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "profile read", "userID", userID)
	return p, nil
}

// CompleteOnboarding upserts the profile and marks onboarding done. The
// completion timestamp is set once and never reset, so re-submits and later
// profile edits are idempotent.
func (s *ProfileService) CompleteOnboarding(ctx context.Context, userID string, req requests.UpdateProfile) (*models.UserProfile, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		p = &models.UserProfile{UserID: userID}
	case err != nil:
		appErr := common.FromDBError(err, "profile")
		logger.Report(ctx, "profile load failed", appErr, "userID", userID)
		return nil, appErr
	}

	// Partial update: only provided fields overwrite (zero value = absent).
	if req.Level != "" {
		p.Preferences.Level = req.Level
	}
	if len(req.Goals) > 0 {
		p.Preferences.Goals = req.Goals
	}
	if req.DailyCommitmentMin != 0 {
		p.Preferences.DailyCommitmentMin = req.DailyCommitmentMin
	}
	if len(req.Topics) > 0 {
		p.Preferences.Topics = req.Topics
	}

	now := time.Now().UTC()
	if p.OnboardingCompletedAt == nil {
		p.OnboardingCompletedAt = &now
	}
	p.UpdatedAt = now

	if err := s.profiles.Upsert(ctx, p); err != nil {
		appErr := common.FromDBError(err, "profile")
		logger.Report(ctx, "profile upsert failed", appErr, "userID", userID)
		return nil, appErr
	}

	slog.InfoContext(ctx, "profile onboarding completed", "userID", userID)
	return p, nil
}
