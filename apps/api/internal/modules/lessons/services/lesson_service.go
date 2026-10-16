package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/utils"
)

// LessonService owns section-first browsing: sections with units attached,
// in path order. The hub renders the whole path, unpaginated.
type LessonService interface {
	ListSections(ctx context.Context, req requests.ListLessons, isAdmin bool) ([]*models.LessonSection, error)
	GetDetail(ctx context.Context, id string, isAdmin bool) (*models.Lesson, error)
	GetDetailBySlug(ctx context.Context, slug string, isAdmin bool) (*models.Lesson, error)
	Create(ctx context.Context, req requests.CreateLesson) (*models.Lesson, error)
	Update(ctx context.Context, id string, req requests.UpdateLesson) (*models.Lesson, error)
	Delete(ctx context.Context, id string) error
	ListAllSections(ctx context.Context) ([]*models.LessonSection, error)
	CreateSection(ctx context.Context, req requests.CreateLessonSection) (*models.LessonSection, error)
	UpdateSection(ctx context.Context, id string, req requests.UpdateLessonSection) (*models.LessonSection, error)
	DeleteSection(ctx context.Context, id string) error
}

type lessonService struct {
	sections repositories.LessonSectionRepository
	lessons  repositories.LessonRepository
}

// NewLessonService builds the service over both repositories. Neither may
// be nil.
func NewLessonService(sections repositories.LessonSectionRepository, lessons repositories.LessonRepository) LessonService {
	return &lessonService{sections: sections, lessons: lessons}
}

func (s *lessonService) ListSections(ctx context.Context, req requests.ListLessons, isAdmin bool) ([]*models.LessonSection, error) {
	secs, err := s.sections.ListWithUnits(ctx, req.Level, isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "list lesson sections failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson sections listed", "count", len(secs))
	return secs, nil
}

// GetDetail: a draft unit is invisible to non-admins (404, existence not
// disclosed), same rule as every other admin-curated resource.
func (s *lessonService) GetDetail(ctx context.Context, id string, isAdmin bool) (*models.Lesson, error) {
	m, err := s.lessons.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "get lesson detail failed", appErr, "id", id)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("lesson not found")
	}
	slog.InfoContext(ctx, "lesson detail found", "id", m.ID)
	return m, nil
}

func (s *lessonService) GetDetailBySlug(ctx context.Context, slug string, isAdmin bool) (*models.Lesson, error) {
	m, err := s.lessons.GetDetailBySlug(ctx, slug)
	if err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "get lesson detail by slug failed", appErr, "slug", slug)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("lesson not found")
	}
	slog.InfoContext(ctx, "lesson detail found", "id", m.ID, "slug", m.Slug)
	return m, nil
}

// toLessonDetails hand-maps the wire input into the model's jsonb-backed
// detail struct. copier's DeepCopy does not reliably recurse into a nested
// field whose destination type implements sql.Scanner/driver.Valuer (as
// LessonDetails does) when that field sits inside a larger struct copy —
// verified empirically against the identical ScenarioDetail case, which
// silently zeroed instead of erroring. This one nested field always gets an
// explicit line-up instead of a blanket utils.Map.
func toLessonDetails(in requests.LessonDetailsInput) models.LessonDetails {
	return models.LessonDetails{
		EstimatedDurationMin: in.EstimatedDurationMin,
		AcousticTargetPct:    in.AcousticTargetPct,
		CoverImageURL:        in.CoverImageURL,
		Skill:                in.Skill,
	}
}

// Create adds an admin-curated unit, starting as draft until explicitly
// published.
func (s *lessonService) Create(ctx context.Context, req requests.CreateLesson) (*models.Lesson, error) {
	m := &models.Lesson{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create lesson failed", err)
		return nil, common.Internal()
	}
	m.Details = toLessonDetails(req.Details)
	m.Status = enums.PublishStatusDraft
	if err := s.lessons.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "create lesson failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson created", "id", m.ID)
	return m, nil
}

func (s *lessonService) Update(ctx context.Context, id string, req requests.UpdateLesson) (*models.Lesson, error) {
	m, err := s.lessons.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "get lesson for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update lesson failed", err, "id", id)
		return nil, common.Internal()
	}
	if req.Details != nil {
		m.Details = toLessonDetails(*req.Details)
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.lessons.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "update lesson failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson updated", "id", id)
	return m, nil
}

func (s *lessonService) Delete(ctx context.Context, id string) error {
	if err := s.lessons.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "lesson")
		logger.Report(ctx, "delete lesson failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "lesson deleted", "id", id)
	return nil
}

// ListAllSections returns every section unfiltered, position-ordered — the
// admin console's section picker and section-management table (unlike
// ListSections/ListWithUnits, which only ever return sections that have at
// least one matching unit for learners).
func (s *lessonService) ListAllSections(ctx context.Context) ([]*models.LessonSection, error) {
	secs, err := s.sections.List(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "list all lesson sections failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "all lesson sections listed", "count", len(secs))
	return secs, nil
}

func (s *lessonService) CreateSection(ctx context.Context, req requests.CreateLessonSection) (*models.LessonSection, error) {
	t := &models.LessonSection{}
	if err := utils.Map(t, req); err != nil {
		logger.Report(ctx, "map create lesson section failed", err)
		return nil, common.Internal()
	}
	if err := s.sections.Create(ctx, t); err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "create lesson section failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson section created", "id", t.ID)
	return t, nil
}

func (s *lessonService) UpdateSection(ctx context.Context, id string, req requests.UpdateLessonSection) (*models.LessonSection, error) {
	t, err := s.sections.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "get lesson section for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(t, req); err != nil {
		logger.Report(ctx, "map update lesson section failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.sections.Update(ctx, t); err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "update lesson section failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson section updated", "id", id)
	return t, nil
}

func (s *lessonService) DeleteSection(ctx context.Context, id string) error {
	if err := s.sections.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "lesson section")
		logger.Report(ctx, "delete lesson section failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "lesson section deleted", "id", id)
	return nil
}
