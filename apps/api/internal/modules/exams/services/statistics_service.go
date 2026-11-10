package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/logger"
)

func (s *AttemptService) Statistics(ctx context.Context, userID string) (*models.ExamStatistics, error) {
	stats, err := s.attempts.GetExamStatistics(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "exam statistics")
		logger.Report(ctx, "get exam statistics failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "exam statistics retrieved", "userID", userID)
	return stats, nil
}
