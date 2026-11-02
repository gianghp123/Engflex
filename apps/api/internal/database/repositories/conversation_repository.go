package repositories

import (
	"context"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

type ConversationRepository interface {
	ApplySnapshot(ctx context.Context, id, sessionID string, revision int64, fingerprint string, turns []*models.ConversationTurn, duration *int) (*models.Conversation, error)
	StoreAnalysis(ctx context.Context, id string, revision int64, feedback *models.Feedback) error
	Create(ctx context.Context, m *models.Conversation) error

	CreateClosingActive(ctx context.Context, m *models.Conversation) error
	GetByID(ctx context.Context, id string) (*models.Conversation, error)

	ListByUser(ctx context.Context, userID string, scenarioID *string, limit, offset int) ([]*models.Conversation, error)

	CountByUser(ctx context.Context, userID string, scenarioID *string) (int64, error)
	SetSpeechStart(ctx context.Context, id, speechSessionID string, startResponse []byte) (int64, error)
	SetStatus(ctx context.Context, id string, from, to enums.ConversationStatus) (int64, error)
	SetEnded(ctx context.Context, id string, durationSec *int) (int64, error)

	DeleteIfEmpty(ctx context.Context, id string) (bool, error)

	ListEndedDays(ctx context.Context, userID string, since time.Time) ([]string, error)

	SumDurationSince(ctx context.Context, userID string, since time.Time) (int64, error)
	WithTx(tx *gorm.DB) ConversationRepository
}

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) WithTx(tx *gorm.DB) ConversationRepository {
	return &conversationRepository{db: tx}
}

func (r *conversationRepository) Create(ctx context.Context, m *models.Conversation) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *conversationRepository) GetByID(ctx context.Context, id string) (*models.Conversation, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return nil, err
	}
	var m models.Conversation
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *conversationRepository) historyQuery(ctx context.Context, userID string, scenarioID *string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Conversation{}).
		Joins("Scenario").
		Joins("Scenario.Persona").
		Where("conversations.user_id = ?", userID).
		Where("conversations.status IN ?", []enums.ConversationStatus{
			enums.ConversationStatusEnded, enums.ConversationStatusFailed,
		}).
		Where("EXISTS (SELECT 1 FROM conversation_turns t WHERE t.conversation_id = conversations.id)")
	if scenarioID != nil && *scenarioID != "" {
		q = q.Where("conversations.scenario_id = ?", *scenarioID)
	}
	return q
}

func (r *conversationRepository) ListByUser(ctx context.Context, userID string, scenarioID *string, limit, offset int) ([]*models.Conversation, error) {
	var out []*models.Conversation
	if err := r.historyQuery(ctx, userID, scenarioID).
		Order("conversations.started_at DESC, conversations.id DESC").
		Limit(limit).Offset(offset).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *conversationRepository) CountByUser(ctx context.Context, userID string, scenarioID *string) (int64, error) {
	var n int64
	if err := r.historyQuery(ctx, userID, scenarioID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *conversationRepository) CreateClosingActive(ctx context.Context, m *models.Conversation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", m.UserID).Error; err != nil {
			return err
		}
		// Do not infer an empty final transcript from a missing callback. Retain
		// terminal rows for final snapshot delivery and idempotent replay.
		if err := tx.Model(&models.Conversation{}).Where("user_id = ? AND status IN ?", m.UserID, []enums.ConversationStatus{enums.ConversationStatusPending, enums.ConversationStatusLive}).Updates(map[string]any{"status": enums.ConversationStatusEnded, "ended_at": gorm.Expr("COALESCE(ended_at, now())")}).Error; err != nil {
			return err
		}
		return tx.Create(m).Error
	})
}

func (r *conversationRepository) SetSpeechStart(ctx context.Context, id, speechSessionID string, startResponse []byte) (int64, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return 0, err
	}
	res := r.db.WithContext(ctx).
		Model(&models.Conversation{}).
		Where("id = ? AND status = ?", id, enums.ConversationStatusPending).
		Updates(map[string]any{
			"speech_session_id":     speechSessionID,
			"speech_start_response": datatypes.JSON(startResponse),
			"updated_at":            gorm.Expr("now()"),
		})
	return res.RowsAffected, res.Error
}

func (r *conversationRepository) SetStatus(ctx context.Context, id string, from, to enums.ConversationStatus) (int64, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return 0, err
	}
	res := r.db.WithContext(ctx).
		Model(&models.Conversation{}).
		Where("id = ? AND status = ?", id, from).
		Updates(map[string]any{"status": to, "updated_at": gorm.Expr("now()")})
	return res.RowsAffected, res.Error
}

func (r *conversationRepository) SetEnded(ctx context.Context, id string, durationSec *int) (int64, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return 0, err
	}
	updates := map[string]any{
		"status":     gorm.Expr("CASE WHEN status = ? THEN status ELSE ? END", enums.ConversationStatusFailed, enums.ConversationStatusEnded),
		"ended_at":   gorm.Expr("COALESCE(ended_at, now())"),
		"updated_at": gorm.Expr("now()"),
	}
	if durationSec != nil {
		updates["duration_sec"] = gorm.Expr("COALESCE(duration_sec, ?)", *durationSec)
	}
	res := r.db.WithContext(ctx).
		Model(&models.Conversation{}).
		Where("id = ?", id).
		Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *conversationRepository) DeleteIfEmpty(ctx context.Context, id string) (bool, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return false, err
	}
	res := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("NOT EXISTS (SELECT 1 FROM conversation_turns WHERE conversation_id = conversations.id)").
		Delete(&models.Conversation{})
	return res.RowsAffected > 0, res.Error
}

func (r *conversationRepository) ListEndedDays(ctx context.Context, userID string, since time.Time) ([]string, error) {
	var days []string
	err := r.db.WithContext(ctx).Model(&models.Conversation{}).
		Select("DISTINCT to_char(ended_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Where("user_id = ? AND status = ? AND ended_at >= ?", userID, enums.ConversationStatusEnded, since).
		Scan(&days).Error
	return days, err
}

func (r *conversationRepository) SumDurationSince(ctx context.Context, userID string, since time.Time) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Conversation{}).
		Select("COALESCE(SUM(duration_sec), 0)").
		Where("user_id = ? AND status = ? AND ended_at >= ?", userID, enums.ConversationStatusEnded, since).
		Scan(&total).Error
	return total, err
}
