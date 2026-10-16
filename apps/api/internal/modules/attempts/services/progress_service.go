// Package services holds the attempts-domain business logic. Progress is the
// first inhabitant; attempt lifecycle still lives in the lessons module.
package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/attempts/dtos/responses"
)

// defaultDailyGoalMinutes is the fallback before onboarding stores a value,
// matching the onboarding default commitment (15 min).
const defaultDailyGoalMinutes = 15

// streakWindowDays caps the streak scan. A streak longer than a year is not a
// product case; the window also bounds the per-request work.
const streakWindowDays = 366

// ProgressService composes the measurable-only dashboard rollup.
type ProgressService struct {
	attempts      repositories.AttemptRepository
	conversations repositories.ConversationRepository
	profiles      repositories.UserProfileRepository
}

// NewProgressService takes interfaces, never concretes.
func NewProgressService(
	attempts repositories.AttemptRepository,
	conversations repositories.ConversationRepository,
	profiles repositories.UserProfileRepository,
) *ProgressService {
	return &ProgressService{attempts: attempts, conversations: conversations, profiles: profiles}
}

// GetProgress returns the fields that have a real data source. Days are
// server UTC — a documented limitation (learner-local midnight can differ);
// per-timezone streaks are a later concern.
func (s *ProgressService) GetProgress(ctx context.Context, userID string) (*responses.Progress, error) {
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	since := dayStart.AddDate(0, 0, -streakWindowDays)

	attemptDays, err := s.attempts.ListCompletedDays(ctx, userID, since)
	if err != nil {
		appErr := common.FromDBError(err, "progress")
		logger.Report(ctx, "progress attempt days failed", appErr, "userID", userID)
		return nil, appErr
	}
	endedDays, err := s.conversations.ListEndedDays(ctx, userID, since)
	if err != nil {
		appErr := common.FromDBError(err, "progress")
		logger.Report(ctx, "progress ended days failed", appErr, "userID", userID)
		return nil, appErr
	}
	lessons, err := s.attempts.CountCompleted(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "progress")
		logger.Report(ctx, "progress count failed", appErr, "userID", userID)
		return nil, appErr
	}
	seconds, err := s.conversations.SumDurationSince(ctx, userID, dayStart)
	if err != nil {
		appErr := common.FromDBError(err, "progress")
		logger.Report(ctx, "progress minutes failed", appErr, "userID", userID)
		return nil, appErr
	}

	active := make(map[string]bool, len(attemptDays)+len(endedDays))
	for _, d := range attemptDays {
		active[d] = true
	}
	for _, d := range endedDays {
		active[d] = true
	}

	goal := defaultDailyGoalMinutes
	if p, perr := s.profiles.GetByUserID(ctx, userID); perr == nil {
		if p.Preferences.DailyCommitmentMin > 0 {
			goal = p.Preferences.DailyCommitmentMin
		}
	} else if !errors.Is(perr, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(perr, "profile")
		logger.Report(ctx, "progress profile failed", appErr, "userID", userID)
		return nil, appErr
	}

	out := &responses.Progress{
		StreakDays:        streak(active, dayStart),
		VoiceMinutesToday: int(seconds / 60),
		DailyGoalMinutes:  goal,
		LessonsCompleted:  int(lessons),
	}
	slog.InfoContext(ctx, "progress computed",
		"userID", userID, "streak", out.StreakDays, "minutes", out.VoiceMinutesToday)
	return out, nil
}

// streak counts back from today — or yesterday, so a streak stays "alive"
// before today's first session — while consecutive UTC days are active.
func streak(active map[string]bool, today time.Time) int {
	mostRecent := today
	if !active[mostRecent.Format("2006-01-02")] {
		mostRecent = today.AddDate(0, 0, -1)
		if !active[mostRecent.Format("2006-01-02")] {
			return 0
		}
	}
	n := 0
	for d := mostRecent; active[d.Format("2006-01-02")]; d = d.AddDate(0, 0, -1) {
		n++
	}
	return n
}
