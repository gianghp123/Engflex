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
)

// TrialService starts and reads the once-per-user trial (spec §8.4).
type TrialService struct {
	trials repositories.UserTrialRepository
}

// NewTrialService builds the trial service.
func NewTrialService(trials repositories.UserTrialRepository) *TrialService {
	return &TrialService{trials: trials}
}

// Start records the trial once: idempotent for an active trial, false
// without error for a used (even expired) one. One row per Clerk user —
// the primary key is the guard, and re-registration is Clerk's problem.
func (s *TrialService) Start(ctx context.Context, userID string) (bool, error) {
	existing, err := s.trials.Get(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(err, "trial")
		logger.Report(ctx, "trial lookup failed", appErr, "userID", userID)
		return false, appErr
	}
	if existing != nil {
		if existing.EndsAt.After(time.Now().UTC()) {
			slog.InfoContext(ctx, "trial start ignored: already active", "userID", userID)
			return true, nil
		}
		// A used trial never restarts — silently idempotent, no error: the
		// onboarding flow must not fail for a returning user.
		slog.InfoContext(ctx, "trial start ignored: already used", "userID", userID)
		return false, nil
	}
	now := time.Now().UTC()
	m := &models.UserTrial{
		UserID:         userID,
		StartedAt:      now,
		EndsAt:         now.AddDate(0, 0, common.TrialDays),
		CreditsGranted: common.TrialCredits,
	}
	if err := s.trials.Create(ctx, m); err != nil {
		// A unique violation means a concurrent start won the race.
		appErr := common.FromDBError(err, "trial")
		if appErr.Status == 409 {
			slog.InfoContext(ctx, "trial start raced: row already exists", "userID", userID)
			return true, nil
		}
		logger.Report(ctx, "trial create failed", appErr, "userID", userID)
		return false, appErr
	}
	slog.InfoContext(ctx, "trial started", "userID", userID, "endsAt", m.EndsAt, "credits", m.CreditsGranted)
	return true, nil
}
