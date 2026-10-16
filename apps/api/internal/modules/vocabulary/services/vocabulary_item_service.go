package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/utils"
)

// VocabularyStats aggregates the hub's stat cards. Plain ints; the
// controller maps this to responses.VocabularyStats with utils.Map — the
// service never touches a response DTO.
type VocabularyStats struct {
	TotalSaved      int
	Mastered        int
	NeedsReview     int
	CustomAdditions int
	DueToday        int
	// SavedToday counts words saved in the caller's local day: the controller
	// passes that day's start, so a UTC+7 reader's "today" is their own.
	SavedToday int
}

// VocabularyItemService owns the shared dictionary reads (catalog + hub),
// scoped single-item reads, private creates, and hub stats.
type VocabularyItemService interface {
	Catalog(ctx context.Context, userID string, f repositories.VocabularyItemFilter) ([]*models.VocabularyItem, int64, error)
	Hub(ctx context.Context, userID string, f repositories.VocabularyItemFilter) ([]*models.VocabularyItem, int64, error)
	GetByID(ctx context.Context, userID, id string) (*models.VocabularyItem, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyItem) (*models.VocabularyItem, error)
	AddWord(ctx context.Context, userID string, req requests.AddUserVocabularyWord) (*models.VocabularyItem, error)
	Stats(ctx context.Context, userID string, now, dayStart time.Time) (VocabularyStats, error)
}

type vocabularyItemService struct {
	items     repositories.VocabularyItemRepository
	userVocab repositories.UserVocabularyRepository
	decks     repositories.VocabularyDeckRepository
	deckItems repositories.VocabularyDeckItemRepository
}

// NewVocabularyItemService builds the service over all four repositories.
func NewVocabularyItemService(items repositories.VocabularyItemRepository, userVocab repositories.UserVocabularyRepository, decks repositories.VocabularyDeckRepository, deckItems repositories.VocabularyDeckItemRepository) VocabularyItemService {
	return &vocabularyItemService{items: items, userVocab: userVocab, decks: decks, deckItems: deckItems}
}

// Catalog lists visible items (curated or the caller's own) under the
// search/cefr filters.
func (s *vocabularyItemService) Catalog(ctx context.Context, userID string, f repositories.VocabularyItemFilter) ([]*models.VocabularyItem, int64, error) {
	list, err := s.items.List(ctx, userID, f)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "list vocabulary catalog failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	total, err := s.items.Count(ctx, userID, f)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "count vocabulary catalog failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "vocabulary catalog listed", "count", len(list), "total", total)
	return list, total, nil
}

// Hub lists the caller's saved items with their user state attached.
func (s *vocabularyItemService) Hub(ctx context.Context, userID string, f repositories.VocabularyItemFilter) ([]*models.VocabularyItem, int64, error) {
	list, err := s.items.ListSaved(ctx, userID, f)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "list saved vocabulary failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	total, err := s.items.CountSaved(ctx, userID, f)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "count saved vocabulary failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "vocabulary hub listed", "count", len(list), "total", total)
	return list, total, nil
}

func (s *vocabularyItemService) GetByID(ctx context.Context, userID, id string) (*models.VocabularyItem, error) {
	m, err := s.items.GetByID(ctx, userID, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "get vocabulary item failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary item found", "id", m.ID)
	return m, nil
}

// Create always writes a private row: no API path creates a curated
// (NULL-owner) entry.
func (s *vocabularyItemService) Create(ctx context.Context, userID string, req requests.CreateVocabularyItem) (*models.VocabularyItem, error) {
	// The stored term is display text; identity lives in the generated term_key.
	// A term that is only whitespace collapses to "" — reject it before any write.
	term := utils.TidyTerm(req.Term)
	if term == "" {
		return nil, common.BadRequest("term is required")
	}
	i := &models.VocabularyItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create vocabulary item failed", err)
		return nil, common.Internal()
	}
	i.Term = term
	i.CEFR = enums.CEFR(req.CEFR)
	// copier does not convert *string -> *enums.VocabularyPartOfSpeech.
	if req.PartOfSpeech != nil {
		pos := enums.VocabularyPartOfSpeech(*req.PartOfSpeech)
		i.PartOfSpeech = &pos
	}
	// Details is *VocabularyDetails on the wire but a value on the model;
	// copier silently drops pointer-to-value struct copies, so it maps alone.
	if req.Details != nil {
		var details models.VocabularyDetails
		if err := utils.Map(&details, req.Details); err != nil {
			logger.Report(ctx, "map create vocabulary details failed", err)
			return nil, common.Internal()
		}
		i.Details = details
	}
	uid := userID
	i.CreatedByUserID = &uid
	if err := s.items.Create(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "create vocabulary item failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary item created", "id", i.ID, "private", true)
	return i, nil
}

// AddWord find-or-creates the caller's private entry for a term — keyed by the
// normalized term, so whitespace and case variants are one word — and enrols it
// in their library. A concurrent double-add can still lose the unique index and
// surface 409; the normal path is idempotent.
func (s *vocabularyItemService) AddWord(ctx context.Context, userID string, req requests.AddUserVocabularyWord) (*models.VocabularyItem, error) {
	term := utils.TidyTerm(req.Term)
	// Validate the term and the optional deck target BEFORE anything is written:
	// a rejected request must leave no side effect (no item, no enrolment, no
	// membership).
	if term == "" {
		return nil, common.BadRequest("term is required")
	}
	if req.DeckID != nil {
		if _, dErr := requireWritableDeck(ctx, s.decks, userID, *req.DeckID); dErr != nil {
			return nil, dErr
		}
	}
	item, err := s.items.GetPrivateByTermKey(ctx, userID, utils.TermKey(term))
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		item = &models.VocabularyItem{
			Term:       term,
			Definition: req.Definition,
			CEFR:       enums.CEFR(req.CEFR),
		}
		// copier does not convert *string -> *enums.VocabularyPartOfSpeech.
		if req.PartOfSpeech != nil {
			pos := enums.VocabularyPartOfSpeech(*req.PartOfSpeech)
			item.PartOfSpeech = &pos
		}
		uid := userID
		item.CreatedByUserID = &uid
		if req.Details != nil {
			// Pointer -> the value's address: a single Map here, unlike Create's
			// two-call workaround, because the destination is addressable.
			if mErr := utils.Map(&item.Details, req.Details); mErr != nil {
				logger.Report(ctx, "map added word details failed", mErr)
				return nil, common.Internal()
			}
		}
		if cErr := s.items.Create(ctx, item); cErr != nil {
			appErr := common.FromDBError(cErr, "vocabulary item")
			logger.Report(ctx, "create added word failed", appErr, "term", term)
			return nil, appErr
		}
	case err != nil:
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "lookup added word failed", appErr, "term", term)
		return nil, appErr
	}

	now := time.Now().UTC()
	source := enums.VocabularySourceManual
	if req.SourceType != "" {
		source = enums.VocabularySource(req.SourceType)
	}
	u := &models.UserVocabulary{
		UserID:     userID,
		ItemID:     item.ID,
		SourceType: source,
		SourceID:   req.SourceID,
		SRSDueAt:   &now, // insert-only: Upsert excludes the schedule on conflict
	}
	// Never modify an existing enrolment: re-adding a saved word must leave its
	// note, source, and schedule untouched (Upsert would overwrite note and
	// source_type on conflict). Only enrol when the word is not yet saved —
	// and the source is a fact about the first save, never rewritten by a
	// later re-add.
	if _, gErr := s.userVocab.Get(ctx, userID, item.ID); gErr != nil {
		if !errors.Is(gErr, gorm.ErrRecordNotFound) {
			appErr := common.FromDBError(gErr, "user vocabulary")
			logger.Report(ctx, "lookup enrolment for added word failed", appErr, "itemID", item.ID)
			return nil, appErr
		}
		// Product cap: only a new enrolment is capped; re-adding a saved word
		// (the found branch) never counts against the list limit.
		owned, cErr := s.userVocab.CountByUser(ctx, userID)
		if cErr != nil {
			appErr := common.FromDBError(cErr, "user vocabulary")
			logger.Report(ctx, "count saved words failed", appErr)
			return nil, appErr
		}
		if owned >= common.MaxUserWords {
			return nil, common.UnprocessableEntity("word list limit reached")
		}
		if uErr := s.userVocab.Upsert(ctx, u, false); uErr != nil {
			appErr := common.FromDBError(uErr, "user vocabulary")
			logger.Report(ctx, "enrol added word failed", appErr, "itemID", item.ID)
			return nil, appErr
		}
	}

	// Optional deck target: the deck was validated above, so this only writes the
	// membership. Both writes are idempotent, so a retry converges without a
	// transaction.
	if req.DeckID != nil {
		row := &models.VocabularyDeckItem{DeckID: *req.DeckID, ItemID: item.ID}
		if _, mErr := ensureDeckItem(ctx, s.deckItems, row); mErr != nil {
			return nil, mErr
		}
	}

	got, gErr := s.items.GetByID(ctx, userID, item.ID)
	if gErr != nil {
		appErr := common.FromDBError(gErr, "vocabulary item")
		logger.Report(ctx, "read added word failed", appErr, "itemID", item.ID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary word added", "itemID", item.ID, "term", term)
	return got, nil
}

// Stats builds the hub's stat cards. now is a parameter (not time.Now())
// so tests pin the clock.
func (s *vocabularyItemService) Stats(ctx context.Context, userID string, now, dayStart time.Time) (VocabularyStats, error) {
	total, err := s.userVocab.CountByUser(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary stats")
		logger.Report(ctx, "count saved failed", appErr)
		return VocabularyStats{}, appErr
	}
	mastered, err := s.userVocab.CountMastered(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary stats")
		logger.Report(ctx, "count mastered failed", appErr)
		return VocabularyStats{}, appErr
	}
	due, err := s.userVocab.CountDue(ctx, userID, now)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary stats")
		logger.Report(ctx, "count due failed", appErr)
		return VocabularyStats{}, appErr
	}
	custom, err := s.items.CountPrivate(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary stats")
		logger.Report(ctx, "count custom failed", appErr)
		return VocabularyStats{}, appErr
	}
	savedToday, err := s.userVocab.CountSavedBetween(ctx, userID, dayStart, dayStart.AddDate(0, 0, 1))
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary stats")
		logger.Report(ctx, "count saved today failed", appErr)
		return VocabularyStats{}, appErr
	}
	slog.InfoContext(ctx, "vocabulary stats built", "total", total)
	return VocabularyStats{
		TotalSaved:      int(total),
		Mastered:        int(mastered),
		NeedsReview:     int(due),
		CustomAdditions: int(custom),
		DueToday:        int(due),
		SavedToday:      int(savedToday),
	}, nil
}
