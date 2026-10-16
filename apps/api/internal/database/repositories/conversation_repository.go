package repositories

import (
	"context"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

// ConversationRepository is the DB access layer for the conversations table.
// Turns live in ConversationTurnRepository.
type ConversationRepository interface {
	Create(ctx context.Context, m *models.Conversation) error
	// CreateClosingActive atomically closes the caller's non-terminal
	// conversations and inserts the new one. Deliberate exception to the
	// "one method = one query" rule: the close + insert must be atomic, and
	// pushing the transaction into the service would require handing it a
	// concrete *gorm.DB (violating "constructors take interfaces").
	CreateClosingActive(ctx context.Context, m *models.Conversation) error
	GetByID(ctx context.Context, id string) (*models.Conversation, error)
	// ListByUser returns the caller's finished conversations, newest first,
	// with scenario + persona attached via one LEFT JOIN round trip. Free
	// talk and deleted scenarios come back with a nil Scenario.
	ListByUser(ctx context.Context, userID string, scenarioID *string, limit, offset int) ([]*models.Conversation, error)
	// CountByUser counts the same set ListByUser returns.
	CountByUser(ctx context.Context, userID string, scenarioID *string) (int64, error)
	SetSpeechStart(ctx context.Context, id, speechSessionID string, startResponse []byte) (int64, error)
	SetStatus(ctx context.Context, id string, from, to enums.ConversationStatus) (int64, error)
	SetEnded(ctx context.Context, id string, durationSec *int) (int64, error)
	// DeleteIfEmpty deletes the conversation only when it has no turns,
	// atomically (one conditional DELETE). Returns true when the row was
	// deleted; a conversation holding any turn is left untouched.
	DeleteIfEmpty(ctx context.Context, id string) (bool, error)
	// ListEndedDays returns the distinct UTC dates (YYYY-MM-DD) on which the
	// caller ended a conversation, at or after since. Progress input.
	ListEndedDays(ctx context.Context, userID string, since time.Time) ([]string, error)
	// SumDurationSince returns the caller's total ended-conversation seconds
	// at or after since. Voice-minutes input.
	SumDurationSince(ctx context.Context, userID string, since time.Time) (int64, error)
	WithTx(tx *gorm.DB) ConversationRepository
}

type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository builds the repository over the shared DB handle.
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
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

// historyQuery is the shared read for the history list: the caller's
// finished conversations with the scenario (and its persona) attached.
// LEFT joins, so free talk and deleted scenarios never disappear.
// The EXISTS filter excludes empty conversations from list **and** count
// (both go through `historyQuery`); an EXISTS subquery does not fan out,
// so the counting constraint above still holds. If a turn is ever deleted
// from a conversation this could hide it — acceptable, turns are never
// deleted individually.
// CountByUser counts through this query, which is correct only while every
// join here is fan-out-free (belongs-to); a future has-many join must not
// count through it — it would duplicate parent rows and inflate the count,
// so count that shape with a separate distinct/count query instead.
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
		// Lock the caller's non-terminal rows so two concurrent creates
		// serialize instead of both slipping through.
		var ids []string
		if err := tx.Model(&models.Conversation{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status IN ?", m.UserID,
				[]enums.ConversationStatus{enums.ConversationStatusPending, enums.ConversationStatusLive}).
			Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) > 0 {
			// Never-started, turn-less rows are artifacts of an abandoned
			// click: no engine session exists for them, so nothing can ever
			// arrive afterwards — delete them instead of closing empty
			// history. Rows that reached the engine (speech_session_id set)
			// or already hold turns are closed as before; a gone-late
			// finalize for a never-started row cannot exist.
			if err := tx.Where("id IN ?", ids).
				Where("COALESCE(speech_session_id, '') = ''").
				Where("NOT EXISTS (SELECT 1 FROM conversation_turns WHERE conversation_id = conversations.id)").
				Delete(&models.Conversation{}).Error; err != nil {
				return err
			}
			var remaining []string
			if err := tx.Model(&models.Conversation{}).Where("id IN ?", ids).Pluck("id", &remaining).Error; err != nil {
				return err
			}
			if len(remaining) > 0 {
				if err := tx.Model(&models.Conversation{}).
					Where("id IN ?", remaining).
					Updates(map[string]any{
						"status":     enums.ConversationStatusEnded,
						"ended_at":   gorm.Expr("COALESCE(ended_at, now())"),
						"updated_at": gorm.Expr("now()"),
					}).Error; err != nil {
					return err
				}
			}
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
		Where("id = ? AND status = ? AND COALESCE(speech_session_id, '') = ''", id, enums.ConversationStatusPending).
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

// ListEndedDays returns distinct UTC calendar dates as text, forcing UTC so
// the result never depends on the server's TimeZone setting.
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
