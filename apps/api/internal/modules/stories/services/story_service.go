package services

import (
	"context"
	"log/slog"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/stories/dtos/requests"
	"engflex-api/internal/utils"
)

type StoryService interface {
	ListStories(ctx context.Context, req requests.ListStories, limit, offset int, isAdmin bool) ([]*models.Story, int64, error)
	GetByID(ctx context.Context, id, userID string, isAdmin bool) (*models.Story, error)
	GetBySlug(ctx context.Context, slug, userID string, isAdmin bool) (*models.Story, error)
	ListTopics(ctx context.Context) ([]string, error)
	Create(ctx context.Context, req requests.CreateStory) (*models.Story, error)
	Update(ctx context.Context, id string, req requests.UpdateStory) (*models.Story, error)
	Delete(ctx context.Context, id string) error
}

type storyService struct {
	repo         repositories.StoryRepository
	premium      common.PremiumChecker
	transactions repositories.TransactionManager
}

func NewStoryService(repo repositories.StoryRepository, premium common.PremiumChecker, transactions repositories.TransactionManager) StoryService {
	return &storyService{repo: repo, premium: premium, transactions: transactions}
}

// toSentenceModels converts the admin form's flat JSON input into rows
// ReplaceSentences can insert. StoryID is filled in by the repository, not
// here, since the caller doesn't know the story's ID yet on create.
func toSentenceModels(in []requests.SentenceInput) []*models.StorySentence {
	out := make([]*models.StorySentence, len(in))
	for i, sentence := range in {
		out[i] = &models.StorySentence{
			Sequence:       sentence.Sequence,
			EnglishText:    sentence.EnglishText,
			VietnameseText: sentence.VietnameseText,
			ImageURL:       sentence.ImageURL,
			StartTimestamp: sentence.StartTimestamp,
			EndTimestamp:   sentence.EndTimestamp,
		}
	}
	return out
}

func (s *storyService) ListStories(ctx context.Context, req requests.ListStories, limit, offset int, isAdmin bool) ([]*models.Story, int64, error) {
	f := repositories.StoryFilter{
		CEFR:          string(req.CEFR),
		Topic:         req.Topic,
		Search:        req.Search,
		OnlyPublished: !isAdmin,
		Limit:         limit,
		Offset:        offset,
	}
	items, err := s.repo.List(ctx, f)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "list stories failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx, f)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "count stories failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "stories listed", "count", len(items), "total", total)
	return items, total, nil
}

// GetByID: a draft story is invisible to non-admins (404, existence not
// disclosed), same rule as video exercises.
func (s *storyService) GetByID(ctx context.Context, id, userID string, isAdmin bool) (*models.Story, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "get story by id failed", appErr, "id", id)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("story not found")
	}
	if err := common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, m.IsPremium); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "story retrieved", "id", id)
	return m, nil
}

func (s *storyService) GetBySlug(ctx context.Context, slug, userID string, isAdmin bool) (*models.Story, error) {
	m, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "get story by slug failed", appErr, "slug", slug)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("story not found")
	}
	if err := common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, m.IsPremium); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "story retrieved by slug", "slug", slug)
	return m, nil
}

// Create is admin-only (route-gated) and always starts as draft — only an
// explicit publish makes it visible to learners. Sentences, if given, are
// written in the same transaction as the story row.
func (s *storyService) Create(ctx context.Context, req requests.CreateStory) (*models.Story, error) {
	m := &models.Story{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create story failed", err)
		return nil, common.Internal()
	}
	m.Status = enums.PublishStatusDraft
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Create(ctx, m); err != nil {
			return err
		}
		if req.Sentences != nil {
			return repo.ReplaceSentences(ctx, m.ID, toSentenceModels(req.Sentences))
		}
		return nil
	})
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "failed to create story", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "story created", "story_id", m.ID)
	return m, nil
}

// Update: Sentences is nil-vs-empty sensitive — omitted (nil) leaves the
// existing timeline untouched, `[]` clears it. A plain metadata edit from
// the admin form always resends its own just-loaded sentences, so nil only
// happens from a caller that never touches content at all.
func (s *storyService) Update(ctx context.Context, id string, req requests.UpdateStory) (*models.Story, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "failed to get story for update", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update story failed", err, "id", id)
		return nil, common.Internal()
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if req.IsPremium != nil {
		m.IsPremium = *req.IsPremium
	}
	err = s.transactions.Run(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Update(ctx, m); err != nil {
			return err
		}
		if req.Sentences != nil {
			return repo.ReplaceSentences(ctx, m.ID, toSentenceModels(req.Sentences))
		}
		return nil
	})
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "failed to update story", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "story updated", "story_id", m.ID)
	return m, nil
}

func (s *storyService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "failed to delete story", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "story deleted", "story_id", id)
	return nil
}

func (s *storyService) ListTopics(ctx context.Context) ([]string, error) {
	topics, err := s.repo.ListTopics(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "story")
		logger.Report(ctx, "list story topics failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "story topics listed", "count", len(topics))
	return topics, nil
}
