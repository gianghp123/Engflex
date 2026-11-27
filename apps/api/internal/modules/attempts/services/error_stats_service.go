// Package services holds the attempts-domain business logic: progress plus
// the display-only mistake aggregation (no drill generation).
package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	lessonsservices "engflex-api/internal/modules/lessons/services"
)

// recentFeedbackCap bounds the aggregation scan: 200 rows, newest first.
const recentFeedbackCap = 200

// recentAttemptCap bounds the writing scan: 50 newest attempts.
const recentAttemptCap = 50

// ErrorStat is one taxonomy bucket with its count.
type ErrorStat struct {
	ErrorType string `json:"errorType"`
	Count     int    `json:"count"`
}

// Mistake is one flattened learner error for the mistakes display: the span
// plus where it came from and when.
type Mistake struct {
	ErrorType  string    `json:"errorType"`
	Text       string    `json:"text"`
	Correction string    `json:"correction"`
	Reason     string    `json:"reason"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ErrorStatsService aggregates error-type counts and mistake rows over
// recent feedbacks plus recent writing attempts. No new table (query-light).
type ErrorStatsService struct {
	feedbacks repositories.FeedbackRepository
	attempts  repositories.AttemptRepository
}

// NewErrorStatsService wires the service over repository interfaces.
func NewErrorStatsService(feedbacks repositories.FeedbackRepository, attempts repositories.AttemptRepository) *ErrorStatsService {
	return &ErrorStatsService{feedbacks: feedbacks, attempts: attempts}
}

// TopErrors returns taxonomy buckets sorted by count desc, truncated to
// limit. Unknown/missing errorType reads as "other" (old payloads predate
// the taxonomy; no backfill migration — payload is jsonb schemaless).
func (s *ErrorStatsService) TopErrors(ctx context.Context, userID string, limit int) ([]ErrorStat, error) {
	mistakes, err := s.listMistakes(ctx, userID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, m := range mistakes {
		counts[m.ErrorType]++
	}
	ranked := make([]ErrorStat, 0, len(counts))
	for t, c := range counts {
		ranked = append(ranked, ErrorStat{ErrorType: t, Count: c})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Count == ranked[j].Count {
			return ranked[i].ErrorType < ranked[j].ErrorType
		}
		return ranked[i].Count > ranked[j].Count
	})
	if limit > 0 && len(ranked) > limit {
		ranked = ranked[:limit]
	}
	slog.InfoContext(ctx, "top errors", "userID", userID, "buckets", len(ranked))
	return ranked, nil
}

// Mistakes returns flattened spans newest first, optionally filtered by
// errorType. Old payloads without errorType read as "other".
func (s *ErrorStatsService) Mistakes(ctx context.Context, userID, errorType string, limit int) ([]Mistake, error) {
	mistakes, err := s.listMistakes(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Mistake, 0, len(mistakes))
	for _, m := range mistakes {
		if errorType != "" && m.ErrorType != errorType {
			continue
		}
		out = append(out, m)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *ErrorStatsService) listMistakes(ctx context.Context, userID string) ([]Mistake, error) {
	fbs, err := s.feedbacks.ListRecentByUser(ctx, userID, recentFeedbackCap)
	if err != nil {
		appErr := common.FromDBError(err, "feedback")
		logger.Report(ctx, "top errors load failed", appErr, "userID", userID)
		return nil, appErr
	}
	out := make([]Mistake, 0, len(fbs)*2)
	for _, fb := range fbs {
		out = appendMistakesFromFeedback(out, fb)
	}
	rows, err := s.attempts.ListRecentByUser(ctx, userID, recentAttemptCap)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "top errors attempts load failed", appErr, "userID", userID)
		return nil, appErr
	}
	for _, m := range rows {
		out = appendMistakesFromAttempt(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func appendMistakesFromFeedback(out []Mistake, fb *models.Feedback) []Mistake {
	for _, sp := range fb.Payload.Spans {
		t := string(sp.ErrorType)
		if t == "" {
			t = "other"
		}
		out = append(out, Mistake{
			ErrorType:  t,
			Text:       sp.Text,
			Correction: sp.Correction,
			Reason:     sp.Reason,
			Source:     "feedback",
			CreatedAt:  fb.CreatedAt,
		})
	}
	return out
}

func appendMistakesFromAttempt(out []Mistake, m *models.Attempt) []Mistake {
	parts := map[string]lessonsservices.PartEntry{}
	if len(m.Result) == 0 {
		return out
	}
	if err := json.Unmarshal(m.Result, &parts); err != nil {
		return out
	}
	for _, entry := range parts {
		if string(entry.Type) != "writing" {
			continue
		}
		raw, err := json.Marshal(entry.Detail)
		if err != nil {
			continue
		}
		var detail lessonsservices.WritingPartDetail
		if err := json.Unmarshal(raw, &detail); err != nil {
			continue
		}
		for _, sp := range detail.Feedback.Grammar {
			t := string(sp.ErrorType)
			if t == "" {
				t = "other"
			}
			out = append(out, Mistake{
				ErrorType:  t,
				Text:       sp.Text,
				Correction: sp.Correction,
				Reason:     sp.Reason,
				Source:     "writing",
				CreatedAt:  m.CreatedAt,
			})
		}
		for _, sp := range detail.Feedback.Phrasing {
			t := string(sp.ErrorType)
			if t == "" {
				t = "other"
			}
			out = append(out, Mistake{
				ErrorType:  t,
				Text:       sp.Text,
				Correction: sp.Correction,
				Reason:     sp.Reason,
				Source:     "writing",
				CreatedAt:  m.CreatedAt,
			})
		}
	}
	return out
}
