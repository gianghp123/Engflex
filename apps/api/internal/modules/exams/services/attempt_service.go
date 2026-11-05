package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sort"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/utils"
)

type AttemptService struct {
	attempts     repositories.AttemptRepository
	tests        repositories.ExamTestRepository
	transactions repositories.TransactionManager
}

func NewAttemptService(attempts repositories.AttemptRepository, tests repositories.ExamTestRepository, transactions repositories.TransactionManager) *AttemptService {
	return &AttemptService{attempts: attempts, tests: tests, transactions: transactions}
}

func (s *AttemptService) Start(ctx context.Context, userID, testID string, req requests.StartAttempt) (*models.Attempt, error) {
	if req.Mode != enums.ExamModeExam && req.Mode != enums.ExamModePractice {
		return nil, common.BadRequest("invalid exam mode")
	}
	if len(req.SelectedParts) == 0 || len(req.SelectedParts) > 7 {
		return nil, common.BadRequest("invalid selected parts")
	}
	parts := append([]int(nil), req.SelectedParts...)
	sort.Ints(parts)
	for i, part := range parts {
		if part < 1 || part > 7 || (i > 0 && part == parts[i-1]) {
			return nil, common.BadRequest("invalid selected parts")
		}
	}
	if req.Mode == enums.ExamModeExam && len(parts) != 7 {
		return nil, common.BadRequest("exam mode requires all parts")
	}
	var attempt *models.Attempt
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		test, err := s.tests.WithTx(tx).GetForUpdate(ctx, testID)
		if err != nil {
			return common.FromDBError(err, "exam test")
		}
		if test.Category == nil || test.Category.Type != enums.ExamTypeTOEIC {
			return common.UnprocessableEntity("unsupported exam type")
		}
		content, err := toeic.ParseContent(test.Content)
		if err != nil {
			logger.Report(ctx, "parse exam content failed", common.Internal(), "testID", testID)
			return common.Internal()
		}
		test.ParsedContent = content
		result := toeic.Result{Revision: 1, Answers: []toeic.Answer{}}
		selected := make(map[int]bool, len(parts))
		for _, part := range parts {
			selected[part] = true
		}
		for _, part := range content.Parts {
			if !selected[part.PartNumber] {
				continue
			}
			for _, item := range part.Items {
				for _, question := range item.Questions {
					result.Answers = append(result.Answers, toeic.Answer{QuestionID: question.ID})
				}
			}
		}
		raw, err := json.Marshal(result)
		if err != nil {
			logger.Report(ctx, "encode exam result failed", common.Internal(), "testID", testID)
			return common.Internal()
		}
		attempt = &models.Attempt{}
		if err := utils.Map(attempt, req); err != nil {
			logger.Report(ctx, "map exam attempt failed", common.Internal(), "testID", testID)
			return common.Internal()
		}
		mode := req.Mode
		now := time.Now().UTC()
		attempt.UserID, attempt.ExamTestID, attempt.Mode = userID, &test.ID, &mode
		if req.Mode == enums.ExamModeExam {
			deadline := now.Add(toeic.DurationMinutes * time.Minute)
			attempt.DeadlineAt = &deadline
		}
		attempt.Type, attempt.Status, attempt.StartedAt = enums.AttemptTypeExam, enums.AttemptStatusInProgress, &now
		attempt.SelectedParts = make(pq.Int64Array, len(parts))
		for i, part := range parts {
			attempt.SelectedParts[i] = int64(part)
		}
		filterSelectedParts(content, attempt.SelectedParts)
		attempt.Result = raw
		attempt.ParsedResult = &result
		if err := s.attempts.WithTx(tx).Create(ctx, attempt); err != nil {
			return common.FromDBError(err, "attempt")
		}
		attempt.ExamTest = test
		return nil
	})
	if err != nil {
		var appErr *common.AppError
		if !errors.As(err, &appErr) {
			appErr = common.FromDBError(err, "attempt")
		}
		logger.Report(ctx, "start exam attempt failed", appErr, "testID", testID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "exam attempt started", "attemptID", attempt.ID, "testID", testID)
	return attempt, nil
}

func (s *AttemptService) List(ctx context.Context, userID string, page, pageSize int) ([]*models.Attempt, int64, error) {
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	attempts, err := s.attempts.ListExamAttempts(ctx, userID, pageSize, utils.Offset(page, pageSize))
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "list exam attempts failed", appErr)
		return nil, 0, appErr
	}
	for _, attempt := range attempts {
		attempt.ParsedResult, err = toeic.ParseResult(attempt.Result)
		if err != nil {
			appErr := common.Internal()
			logger.Report(ctx, "parse exam history result failed", appErr, "attemptID", attempt.ID, "error", err)
			return nil, 0, appErr
		}
	}
	total, err := s.attempts.CountExamAttempts(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "count exam attempts failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "exam attempts listed", "count", len(attempts), "total", total)
	return attempts, total, nil
}

func (s *AttemptService) Get(ctx context.Context, userID, attemptID string) (*models.Attempt, error) {
	return s.change(ctx, userID, attemptID, nil, false)
}

func filterSelectedParts(content *toeic.Content, selectedParts pq.Int64Array) {
	selected := make(map[int]bool, len(selectedParts))
	for _, part := range selectedParts {
		selected[int(part)] = true
	}
	parts := make([]toeic.Part, 0, len(selectedParts))
	for _, part := range content.Parts {
		if selected[part.PartNumber] {
			parts = append(parts, part)
		}
	}
	content.Parts = parts
}
