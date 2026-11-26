package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/logger"
)

// AIReviewService orchestrates paid exam AI (spec §8.3): cache read free;
// generate = decide → reserve → LLM → store output + commit. Output and
// charge commit together with a two-step boundary: the ai_outputs row is
// inserted first and the usage Commit follows best-effort (a Commit failure
// is logged loudly and left to the sweeper, which releases the uncommitted
// hold on expiry). Never charge without an output; never return an output
// without the insert.
type AIReviewService struct {
	attempts repositories.AttemptRepository
	tests    repositories.ExamTestRepository
	outputs  repositories.AIOutputRepository
	llm      llm.LLMClient
	policy   PolicyDecider
	usage    UsageMeter
	enforce  bool
}

// NewAIReviewService builds the paid-AI orchestrator.
func NewAIReviewService(attempts repositories.AttemptRepository, tests repositories.ExamTestRepository, outputs repositories.AIOutputRepository, llm llm.LLMClient, policy PolicyDecider, usage UsageMeter, enforce bool) *AIReviewService {
	return &AIReviewService{attempts: attempts, tests: tests, outputs: outputs, llm: llm, policy: policy, usage: usage, enforce: enforce}
}

// GetExamReview is the free cached read (owner-only; 404 when absent).
func (s *AIReviewService) GetExamReview(ctx context.Context, userID, attemptID string) (*models.AIOutput, error) {
	return s.getCached(ctx, userID, attemptID, enums.AIOutputKindExamReview)
}

// GetStudyPlan is the free cached read for the plan.
func (s *AIReviewService) GetStudyPlan(ctx context.Context, userID, attemptID string) (*models.AIOutput, error) {
	return s.getCached(ctx, userID, attemptID, enums.AIOutputKindStudyPlan)
}

func (s *AIReviewService) getCached(ctx context.Context, userID, attemptID string, kind enums.AIOutputKind) (*models.AIOutput, error) {
	out, err := s.outputs.GetBySubject(ctx, enums.UsageRefTypeExamAttempt, attemptID, kind)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("review not generated")
		}
		appErr := common.FromDBError(err, "ai output")
		logger.Report(ctx, "ai output read failed", appErr, "attemptID", attemptID, "kind", string(kind))
		return nil, appErr
	}
	if out.UserID != userID {
		// Owner-only: another user's cached output is a 404, never a leak.
		return nil, common.NotFound("review not generated")
	}
	slog.InfoContext(ctx, "ai output read", "attemptID", attemptID, "kind", string(kind))
	return out, nil
}

// GenerateExamReview charges 1 credit and builds the mistake review for a
// completed attempt (spec §8.3).
func (s *AIReviewService) GenerateExamReview(ctx context.Context, userID, attemptID string) (*models.AIOutput, error) {
	return s.generate(ctx, userID, attemptID, ActionExamReview,
		enums.AIOutputKindExamReview, 1, "exam_review:"+attemptID, buildReviewPrompt)
}

// GenerateStudyPlan charges 2 credits and links free lessons/videos.
func (s *AIReviewService) GenerateStudyPlan(ctx context.Context, userID, attemptID string) (*models.AIOutput, error) {
	return s.generate(ctx, userID, attemptID, ActionStudyPlan,
		enums.AIOutputKindStudyPlan, 2, "study_plan:"+attemptID, buildPlanPrompt)
}

type promptBuilder func(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) (string, any, string, map[string]any, error)

func (s *AIReviewService) generate(ctx context.Context, userID, attemptID string, action Action, kind enums.AIOutputKind, credits int, idempotencyBase string, build promptBuilder) (*models.AIOutput, error) {
	m, err := s.attempts.GetByID(ctx, attemptID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "ai review load attempt failed", appErr, "attemptID", attemptID)
		return nil, appErr
	}
	if m.UserID != userID {
		return nil, common.NotFound("attempt not found")
	}
	if m.Status != enums.AttemptStatusCompleted {
		return nil, common.UnprocessableEntity("attempt is not completed")
	}
	if existing, err := s.getCached(ctx, userID, attemptID, kind); err == nil {
		return existing, nil
	} else if appErr, ok := err.(*common.AppError); !ok || appErr.Status != http.StatusNotFound {
		// getCached only surfaces AppError: a non-404 is a store failure
		// and must fail the generate, while a 404 means "not yet
		// generated" and falls through. The owner-mismatch 404 cannot
		// happen here (attempt ownership was just verified), so any 404
		// is the cache miss.
		return nil, err
	}

	p, _, err := s.policy.Principal(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	d, err := s.policy.Decide(ctx, userID, false, action)
	if err != nil {
		return nil, err
	}
	if !d.Allow && s.enforce {
		return nil, AppError(d)
	}
	var reservation *Reservation
	if s.enforce {
		reservation, err = s.usage.Reserve(ctx, userID, p.Plan, d.Charges, idempotencyBase,
			enums.UsageRefTypeExamAttempt, attemptID, common.ReviewLeaseTTL, time.Now().UTC(), s.policy.Location())
		if err != nil {
			appErr, ok := err.(*common.AppError)
			if !ok {
				return nil, err
			}
			if appErr.Code == "in_progress" {
				// Committed by a concurrent/past request: replay the cache.
				if existing, gerr := s.getCached(ctx, userID, attemptID, kind); gerr == nil {
					return existing, nil
				}
			}
			return nil, appErr
		}
		if reservation != nil && reservation.Status == ReservationStatusCommitted {
			// A committed replay: another request already settled this
			// item's hold — re-read and replay the cache without calling
			// the LLM.
			if existing, gerr := s.getCached(ctx, userID, attemptID, kind); gerr == nil {
				return existing, nil
			}
			return nil, common.Conflict("a request for this item is already in progress").WithCode("in_progress")
		}
	}

	prompt, output, schemaName, schema, err := build(ctx, s.tests, m)
	if err != nil {
		if reservation != nil {
			_ = s.usage.Release(ctx, reservation.ID)
		}
		return nil, err
	}
	if err := s.llm.CompleteStructured(ctx, prompt, output, schemaName, schema); err != nil {
		if reservation != nil {
			_ = s.usage.Release(ctx, reservation.ID)
		}
		logger.Report(ctx, "ai review llm failed",
			common.ServiceUnavailable("review unavailable").WithCode(common.CodeAIUnavailable),
			"attemptID", attemptID, "kind", string(kind))
		return nil, common.ServiceUnavailable("review unavailable").WithCode(common.CodeAIUnavailable)
	}
	raw, err := json.Marshal(output)
	if err != nil {
		if reservation != nil {
			_ = s.usage.Release(ctx, reservation.ID)
		}
		logger.Report(ctx, "ai review marshal failed", common.Internal(), "attemptID", attemptID)
		return nil, common.Internal()
	}
	out := &models.AIOutput{
		UserID:         userID,
		SubjectType:    enums.UsageRefTypeExamAttempt,
		SubjectID:      attemptID,
		Kind:           kind,
		SchemaVersion:  1,
		Payload:        raw,
		CreditsCharged: credits,
	}
	if err := s.outputs.Create(ctx, out); err != nil {
		if reservation != nil {
			_ = s.usage.Release(ctx, reservation.ID)
		}
		appErr := common.FromDBError(err, "ai output")
		logger.Report(ctx, "ai output store failed", appErr, "attemptID", attemptID, "kind", string(kind))
		return nil, appErr
	}
	if reservation != nil {
		if err := s.usage.Commit(ctx, reservation.ID); err != nil {
			// Same policy as voice: the output exists; the sweeper
			// reconciles a failed settlement. Never strand a paid result.
			slog.WarnContext(ctx, "usage commit after ai output failed", "attemptID", attemptID, "reservationID", reservation.ID)
		}
	}
	slog.InfoContext(ctx, "ai output generated", "attemptID", attemptID, "kind", string(kind), "credits", credits, "metered", reservation != nil)
	return out, nil
}
