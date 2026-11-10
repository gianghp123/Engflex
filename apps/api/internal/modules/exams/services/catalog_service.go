package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/ielts"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/logger"
	"engflex-api/internal/utils"
)

type CatalogService struct {
	categories repositories.ExamCategoryRepository
	tests      repositories.ExamTestRepository
}

func NewCatalogService(categories repositories.ExamCategoryRepository, tests repositories.ExamTestRepository) *CatalogService {
	return &CatalogService{categories: categories, tests: tests}
}

func (s *CatalogService) GetTest(ctx context.Context, testID string) (*models.ExamTestCatalog, error) {
	test, err := s.tests.GetByID(ctx, testID)
	if err != nil {
		appErr := common.FromDBError(err, "exam test")
		logger.Report(ctx, "get exam test failed", appErr, "testID", testID)
		return nil, appErr
	}
	if test.Category == nil {
		return nil, common.UnprocessableEntity("missing test category")
	}

	switch test.Category.Type {
	case enums.ExamTypeTOEIC:
		test.ParsedContent, err = toeic.ParseContent(test.Content)
		if err != nil {
			appErr := common.Internal()
			logger.Report(ctx, "parse catalog exam failed", appErr, "testID", testID, "error", err)
			return nil, appErr
		}
		test.DurationMinutes = toeic.DurationMinutes
	case enums.ExamTypeIELTS:
		test.ParsedIeltsContent, err = ielts.ParseContent(test.Content)
		if err != nil {
			appErr := common.Internal()
			logger.Report(ctx, "parse catalog ielts exam failed", appErr, "testID", testID, "error", err)
			return nil, appErr
		}
		skill := enums.ExamSkillReading
		if test.Skill != nil {
			skill = *test.Skill
		}
		switch skill {
		case enums.ExamSkillListening:
			test.DurationMinutes = ielts.ListeningDurationMinutes
		case enums.ExamSkillWriting:
			test.DurationMinutes = ielts.WritingDurationMinutes
		default:
			test.DurationMinutes = ielts.ReadingDurationMinutes
		}
	default:
		return nil, common.UnprocessableEntity("unsupported exam type")
	}

	slog.InfoContext(ctx, "exam test retrieved", "testID", testID)
	return test, nil
}

func (s *CatalogService) ListCategories(ctx context.Context, examType enums.ExamType) ([]*models.ExamCategoryCatalog, error) {
	if examType != "" && examType != enums.ExamTypeTOEIC && examType != enums.ExamTypeIELTS {
		return nil, common.BadRequest("invalid exam type")
	}
	categories, err := s.categories.List(ctx, examType)
	if err != nil {
		appErr := common.FromDBError(err, "exam category")
		logger.Report(ctx, "list exam categories failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "exam categories listed", "count", len(categories))
	return categories, nil
}

func (s *CatalogService) ListTests(ctx context.Context, categoryID string, page, pageSize int) ([]*models.ExamTestCatalog, int64, error) {
	category, err := s.categories.GetByID(ctx, categoryID)
	if err != nil {
		appErr := common.FromDBError(err, "exam category")
		logger.Report(ctx, "find exam category failed", appErr, "categoryID", categoryID)
		return nil, 0, appErr
	}
	if category.Type != enums.ExamTypeTOEIC && category.Type != enums.ExamTypeIELTS {
		return nil, 0, common.UnprocessableEntity("unsupported exam type")
	}
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	tests, err := s.tests.ListByCategory(ctx, categoryID, pageSize, utils.Offset(page, pageSize))
	if err != nil {
		appErr := common.FromDBError(err, "exam test")
		logger.Report(ctx, "list exam tests failed", appErr, "categoryID", categoryID)
		return nil, 0, appErr
	}
	for _, test := range tests {
		switch category.Type {
		case enums.ExamTypeTOEIC:
			test.DurationMinutes = toeic.DurationMinutes
			test.ParsedContent, err = toeic.ParseContent(test.Content)
			if err != nil {
				logger.Report(ctx, "parse catalog exam failed", common.Internal(), "testID", test.ID)
				return nil, 0, common.Internal()
			}
		case enums.ExamTypeIELTS:
			skill := enums.ExamSkillReading
			if test.Skill != nil {
				skill = *test.Skill
			}
			switch skill {
			case enums.ExamSkillListening:
				test.DurationMinutes = ielts.ListeningDurationMinutes
			case enums.ExamSkillWriting:
				test.DurationMinutes = ielts.WritingDurationMinutes
			default:
				test.DurationMinutes = ielts.ReadingDurationMinutes
			}
			test.ParsedIeltsContent, err = ielts.ParseContent(test.Content)
			if err != nil {
				logger.Report(ctx, "parse catalog ielts exam failed", common.Internal(), "testID", test.ID)
				return nil, 0, common.Internal()
			}
		}
	}
	total, err := s.tests.CountByCategory(ctx, categoryID)
	if err != nil {
		appErr := common.FromDBError(err, "exam test")
		logger.Report(ctx, "count exam tests failed", appErr, "categoryID", categoryID)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "exam tests listed", "categoryID", categoryID, "count", len(tests), "total", total)
	return tests, total, nil
}
