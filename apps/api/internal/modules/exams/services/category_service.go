package services

import (
	"context"
	"log/slog"
	"strings"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/utils"
)

func (s *CatalogService) CreateCategory(ctx context.Context, req requests.CreateCategory) (*models.ExamCategory, error) {
	if strings.TrimSpace(req.Slug) == "" || strings.TrimSpace(req.Title) == "" {
		return nil, common.BadRequest("slug and title are required")
	}
	if req.Type != enums.ExamTypeTOEIC && req.Type != enums.ExamTypeIELTS {
		return nil, common.BadRequest("invalid exam type")
	}
	if req.SubType != nil {
		if *req.SubType != enums.ExamSubTypeAcademic && *req.SubType != enums.ExamSubTypeGeneralTraining {
			return nil, common.BadRequest("invalid exam subtype")
		}
		if req.Type != enums.ExamTypeIELTS {
			return nil, common.BadRequest("subType is only supported for IELTS")
		}
	}
	category := &models.ExamCategory{}
	if err := utils.Map(category, req); err != nil {
		logger.Report(ctx, "map create exam category failed", err)
		return nil, common.Internal()
	}
	if err := s.categories.Create(ctx, category); err != nil {
		appErr := common.FromDBError(err, "exam category")
		logger.Report(ctx, "create exam category failed", appErr, "slug", req.Slug)
		return nil, appErr
	}
	slog.InfoContext(ctx, "exam category created", "categoryID", category.ID)
	return category, nil
}
