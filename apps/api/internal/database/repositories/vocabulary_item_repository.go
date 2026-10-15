package repositories

import (
	"context"
	"fmt"
	"strings"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// VocabularyItemFilter scopes the catalog/hub reads. Search matches term or
// definition (case-insensitive); CEFR is an exact level match. Source and
// Sort apply to ListSaved/CountSaved only — List/Count ignore them.
type VocabularyItemFilter struct {
	Search string
	CEFR   string
	Source string
	Sort   string
	Limit  int
	Offset int
}

// VocabularyItemRepository is the DB access layer for the vocabulary_items table.
//
// Reads are visibility-scoped: curated rows (created_by_user_id IS NULL) are
// readable by everyone, private rows only by their owner. The hub reads join
// the caller's user_vocabulary row as UserState.
type VocabularyItemRepository interface {
	List(ctx context.Context, userID string, f VocabularyItemFilter) ([]*models.VocabularyItem, error)
	ListSaved(ctx context.Context, userID string, f VocabularyItemFilter) ([]*models.VocabularyItem, error)
	Count(ctx context.Context, userID string, f VocabularyItemFilter) (int64, error)
	CountSaved(ctx context.Context, userID string, f VocabularyItemFilter) (int64, error)
	CountPrivate(ctx context.Context, userID string) (int64, error)
	GetByID(ctx context.Context, userID, id string) (*models.VocabularyItem, error)
	GetPrivateByTermKey(ctx context.Context, userID, termKey string) (*models.VocabularyItem, error)
	Create(ctx context.Context, i *models.VocabularyItem) error
	WithTx(tx *gorm.DB) VocabularyItemRepository
}

type vocabularyItemRepository struct {
	db *gorm.DB
}

// NewVocabularyItemRepository builds the repository over the shared DB handle.
func NewVocabularyItemRepository(db *gorm.DB) VocabularyItemRepository {
	return &vocabularyItemRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *vocabularyItemRepository) WithTx(tx *gorm.DB) VocabularyItemRepository {
	return &vocabularyItemRepository{db: tx}
}

// visibleItem builds the "curated or mine" predicate over the given table
// qualifier. Curated rows (NULL owner) are readable by everyone; private rows
// only by their owner. qualifier is a compile-time constant.
func visibleItem(qualifier, userID string) (string, any) {
	return fmt.Sprintf("(%[1]s.created_by_user_id IS NULL OR %[1]s.created_by_user_id = ?)", qualifier), userID
}

// applyItemFilters adds the search/cefr filters. itemQualifier is the alias
// the query knows the items table by ("vocabulary_items" or the join alias).
func applyItemFilters(q *gorm.DB, itemQualifier string, f VocabularyItemFilter) *gorm.DB {
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		q = q.Where(fmt.Sprintf("lower(%s.term) LIKE ? OR lower(%s.definition) LIKE ?", itemQualifier, itemQualifier), like, like)
	}
	if f.CEFR != "" {
		q = q.Where(fmt.Sprintf("%s.cefr = ?", itemQualifier), f.CEFR)
	}
	return q
}

func (r *vocabularyItemRepository) List(ctx context.Context, userID string, f VocabularyItemFilter) ([]*models.VocabularyItem, error) {
	where, arg := visibleItem("vocabulary_items", userID)
	q := r.db.WithContext(ctx).
		// LEFT JOIN the caller's state; NULL-tolerant so unsaved items survive.
		Joins("UserState").
		Where(`"UserState".user_id IS NULL OR "UserState".user_id = ?`, userID).
		Where(where, arg)
	q = applyItemFilters(q, "vocabulary_items", f)
	out := make([]*models.VocabularyItem, 0)
	if err := q.Order("vocabulary_items.term asc").Limit(f.Limit).Offset(f.Offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *vocabularyItemRepository) ListSaved(ctx context.Context, userID string, f VocabularyItemFilter) ([]*models.VocabularyItem, error) {
	where, arg := visibleItem("vocabulary_items", userID)
	q := r.db.WithContext(ctx).
		Joins("UserState").
		Where(`"UserState".user_id = ?`, userID). // strict: hub = saved rows only
		Where(where, arg)
	q = applyItemFilters(q, "vocabulary_items", f)
	if f.Source != "" {
		q = q.Where(`"UserState".source_type = ?`, f.Source)
	}
	out := make([]*models.VocabularyItem, 0)
	if err := q.Order(itemOrder(f.Sort)).Limit(f.Limit).Offset(f.Offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// itemOrder maps the sort key to a column on the quoted join alias. Defaults to
// newest-saved.
func itemOrder(sort string) string {
	switch sort {
	case "alphabetical":
		return "vocabulary_items.term asc"
	case "mastery":
		return `"UserState".mastered desc, "UserState".srs_due_at asc nulls last`
	case "interval":
		return `"UserState".srs_due_at asc nulls last`
	default:
		return `"UserState".created_at desc`
	}
}

// Count is the catalog total: visible items under the same search/cefr
// filters, without the user-state join.
func (r *vocabularyItemRepository) Count(ctx context.Context, userID string, f VocabularyItemFilter) (int64, error) {
	where, arg := visibleItem("vocabulary_items", userID)
	q := r.db.WithContext(ctx).Model(&models.VocabularyItem{}).Where(where, arg)
	q = applyItemFilters(q, "vocabulary_items", f)
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *vocabularyItemRepository) CountSaved(ctx context.Context, userID string, f VocabularyItemFilter) (int64, error) {
	where, arg := visibleItem("vocabulary_items", userID)
	q := r.db.WithContext(ctx).Model(&models.VocabularyItem{}).
		Joins("UserState").
		Where(`"UserState".user_id = ?`, userID).
		Where(where, arg)
	q = applyItemFilters(q, "vocabulary_items", f)
	if f.Source != "" {
		q = q.Where(`"UserState".source_type = ?`, f.Source)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *vocabularyItemRepository) CountPrivate(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.VocabularyItem{}).
		Where("created_by_user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *vocabularyItemRepository) GetByID(ctx context.Context, userID, id string) (*models.VocabularyItem, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	where, arg := visibleItem("vocabulary_items", userID)
	var m models.VocabularyItem
	if err := r.db.WithContext(ctx).
		Joins("UserState").
		Where(`"UserState".user_id IS NULL OR "UserState".user_id = ?`, userID).
		Where("vocabulary_items.id = ?", id).
		Where(where, arg).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vocabularyItemRepository) Create(ctx context.Context, i *models.VocabularyItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

// GetPrivateByTermKey finds the caller's own entry by normalized term key. It
// never returns a curated or another user's row. The caller normalizes with
// utils.TermKey — the repository does not own that rule.
func (r *vocabularyItemRepository) GetPrivateByTermKey(ctx context.Context, userID, termKey string) (*models.VocabularyItem, error) {
	var m models.VocabularyItem
	if err := r.db.WithContext(ctx).
		Where("created_by_user_id = ? AND term_key = ?", userID, termKey).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
