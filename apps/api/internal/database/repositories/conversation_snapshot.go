package repositories

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The conversation row serializes snapshots, finalization and analysis commits.
func (r *conversationRepository) ApplySnapshot(ctx context.Context, id, sessionID string, revision int64, fingerprint string, turns []*models.ConversationTurn, duration *int) (*models.Conversation, error) {
	if _, err := parseID(id, "conversationId"); err != nil {
		return nil, err
	}
	var conv models.Conversation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&conv, "id = ?", id).Error; err != nil {
			return err
		}
		if sessionID == "" || conv.SpeechSessionID != sessionID {
			return common.Conflict("engine session does not match conversation")
		}
		if revision < conv.TranscriptRevision {
			return nil
		}
		if revision == conv.TranscriptRevision && fingerprint != conv.TranscriptHash {
			return common.Conflict("snapshot revision payload differs")
		}
		if conv.TranscriptFinalized {
			if revision == conv.TranscriptRevision && (duration == nil || conv.DurationSec != nil && *duration == *conv.DurationSec) {
				return nil
			}
			return common.Conflict("transcript is finalized")
		}
		if revision > conv.TranscriptRevision {
			// Keep coaching for an unchanged prefix. A correction invalidates
			// its own turn and every later turn whose analysis context changed.
			if fingerprint != conv.TranscriptHash {
				var previous []*models.ConversationTurn
				if err := tx.Where("conversation_id = ?", id).Order("position ASC").Find(&previous).Error; err != nil {
					return err
				}
				firstChanged := min(len(previous), len(turns)) + 1
				for i := 0; i < min(len(previous), len(turns)); i++ {
					old, next := previous[i], turns[i]
					if old.Position != next.Position || old.Role != next.Role || old.Text != next.Text || old.WasInterrupted != next.WasInterrupted {
						firstChanged = i + 1
						break
					}
				}
				if err := tx.Where("subject_type = ? AND subject_id IN (?)", enums.FeedbackSubjectConversationTurn, tx.Model(&models.ConversationTurn{}).Select("id").Where("conversation_id = ? AND position >= ?", id, firstChanged)).Delete(&models.Feedback{}).Error; err != nil {
					return err
				}
			}
			if len(turns) > 0 {
				if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "conversation_id"}, {Name: "position"}}, DoUpdates: clause.AssignmentColumns([]string{"role", "text", "was_interrupted", "updated_at"})}).Create(&turns).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("conversation_id = ? AND position > ?", id, len(turns)).Delete(&models.ConversationTurn{}).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{"transcript_revision": revision, "transcript_hash": fingerprint}
		if duration != nil {
			updates["transcript_finalized"] = true
			updates["duration_sec"] = *duration
			updates["ended_at"] = gorm.Expr("COALESCE(ended_at, now())")
			updates["status"] = gorm.Expr("CASE WHEN status = ? THEN status ELSE ? END", enums.ConversationStatusFailed, enums.ConversationStatusEnded)
		}
		if err := tx.Model(&conv).Updates(updates).Error; err != nil {
			return err
		}
		// Retain empty finalized rows for replay safety. History already excludes
		// zero-turn conversations, and cleanup never infers failed delivery.
		return nil
	})
	return &conv, err
}

func (r *conversationRepository) StoreAnalysis(ctx context.Context, id string, revision int64, feedback *models.Feedback) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv models.Conversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&conv, "id = ?", id).Error; err != nil {
			return err
		}
		if conv.TranscriptRevision != revision {
			return common.Conflict("transcript changed during analysis")
		}
		return NewFeedbackRepository(tx).Upsert(ctx, feedback)
	})
}
