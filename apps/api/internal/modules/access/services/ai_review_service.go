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
// generate = decide → reserve → LLM → store output + commit. The output row
// and the charge commit together in ONE transaction: no output without a
// charge, no charge without an output. An admin bypasses metering entirely:
// after Principal, Decide denials are ignored and no hold is ever reserved
// or committed. Shadow mode evaluates Decide for logging only and never
// fails on a store error.
type AIReviewService struct {
	attempts repositories.AttemptRepository
	tests    repositories.ExamTestRepository
	outputs  repositories.AIOutputRepository
	llm      llm.LLMClient
	policy   PolicyDecider
	usage    UsageMeter
	enforce  bool
	tx       repositories.TransactionManager
}

// NewAIReviewService builds the paid-AI orchestrator. tx coordinates the
// output insert with its settlement; it must be non-nil whenever generate
// can meter (enforce mode), like every other constructor dependency.
func NewAIReviewService(attempts repositories.AttemptRepository, tests repositories.ExamTestRepository, outputs repositories.AIOutputRepository, llm llm.LLMClient, policy PolicyDecider, usage UsageMeter, enforce bool, tx repositories.TransactionManager) *AIReviewService {
	return &AIReviewService{attempts: attempts, tests: tests, outputs: outputs, llm: llm, policy: policy, usage: usage, enforce: enforce, tx: tx}
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

// GenerateExamReview charges the configured credit cost and builds the
// mistake review for a completed attempt (spec §8.3).
func (s *AIReviewService) GenerateExamReview(ctx context.Context, userID, attemptID string, isAdmin bool) (*models.AIOutput, error) {
	return s.generate(ctx, userID, attemptID, ActionExamReview,
		enums.AIOutputKindExamReview, "exam_review:"+attemptID, enums.UsageRefTypeExamAttempt, isAdmin, buildReviewPrompt)
}

// GenerateStudyPlan charges the configured credit cost and links free
// lessons/videos.
func (s *AIReviewService) GenerateStudyPlan(ctx context.Context, userID, attemptID string, isAdmin bool) (*models.AIOutput, error) {
	return s.generate(ctx, userID, attemptID, ActionStudyPlan,
		enums.AIOutputKindStudyPlan, "study_plan:"+attemptID, enums.UsageRefTypeStudyPlan, isAdmin, buildPlanPrompt)
}

type promptBuilder func(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) (string, any, string, map[string]any, error)

func (s *AIReviewService) generate(ctx context.Context, userID, attemptID string, action Action, kind enums.AIOutputKind, idempotencyBase string, refType enums.UsageRefType, isAdmin bool, build promptBuilder) (*models.AIOutput, error) {
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

	p, _, err := s.policy.Principal(ctx, userID, isAdmin)
	var reservation *Reservation
	// credits stamps the reserved amount, so it must equal the hold: derive
	// it from the decided charges, not from a second policy read an admin
	// refresh could have moved in between. The CreditCost fallback runs only
	// when there is no successful decision (shadow-mode store failures) —
	// that path never reserves, so there is nothing to mismatch.
	credits := 0
	if err != nil {
		if s.enforce {
			return nil, err
		}
		logger.Report(ctx, "ai review principal failed, proceeding unmetered", err, "attemptID", attemptID)
		credits = s.policy.CreditCost(action)
	} else if dd, derr := s.policy.Decide(ctx, userID, isAdmin, action); derr != nil {
		if s.enforce {
			return nil, derr
		}
		logger.Report(ctx, "ai review decide failed, proceeding unmetered", derr, "attemptID", attemptID, "kind", string(kind))
		credits = s.policy.CreditCost(action)
	} else {
		d := dd
		credits = aiCreditTotal(d.Charges)
		if p.IsAdmin {
			slog.InfoContext(ctx, "ai review admin bypass", "attemptID", attemptID, "userID", userID, "kind", string(kind))
		} else if !d.Allow && s.enforce {
			return nil, AppError(d)
		} else if s.enforce && len(d.Charges) > 0 {
			reservation, err = s.usage.Reserve(ctx, userID, p.Plan, action, d.Charges, idempotencyBase,
				refType, attemptID, common.ReviewLeaseTTL, time.Now().UTC(), s.policy.Location())
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
	if reservation != nil {
		// The output row and its settlement commit atomically (spec §7.1):
		// no output without a charge, no charge without an output. A tx
		// error rolls back and releases the hold — never strand a paid
		// result the ledger cannot see.
		if s.tx == nil {
			_ = s.usage.Release(ctx, reservation.ID)
			appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
			logger.Report(ctx, "ai review miswired: transaction manager missing", appErr, "attemptID", attemptID)
			return nil, appErr
		}
		txErr := s.tx.Run(ctx, func(tx *gorm.DB) error {
			if err := s.outputs.WithTx(tx).Create(ctx, out); err != nil {
				return &outputStoreError{err}
			}
			return s.usage.CommitTx(ctx, tx, reservation.ID)
		})
		if txErr != nil {
			_ = s.usage.Release(ctx, reservation.ID)
			var ose *outputStoreError
			if errors.As(txErr, &ose) {
				appErr := common.FromDBError(ose.err, "ai output")
				logger.Report(ctx, "ai output store failed", appErr, "attemptID", attemptID, "kind", string(kind))
				return nil, appErr
			}
			appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
			logger.Report(ctx, "ai output settle failed", appErr, "attemptID", attemptID, "reservationID", reservation.ID, "cause", txErr.Error())
			return nil, appErr
		}
	} else if err := s.outputs.Create(ctx, out); err != nil {
		// Unmetered (shadow/admin) generations still persist: the output is
		// the billable artifact, and there is no hold to reconcile.
		appErr := common.FromDBError(err, "ai output")
		logger.Report(ctx, "ai output store failed", appErr, "attemptID", attemptID, "kind", string(kind))
		return nil, appErr
	}
	slog.InfoContext(ctx, "ai output generated", "attemptID", attemptID, "kind", string(kind), "credits", credits, "metered", reservation != nil)
	return out, nil
}

// aiCreditTotal sums the ai_credit charges of a decision: the stamped
// CreditsCharged must equal the reserved hold, never a second policy read
// that an admin refresh could have moved in between.
func aiCreditTotal(charges []Charge) int {
	total := 0
	for _, ch := range charges {
		if ch.Counter == enums.CounterKindAICredit {
			total += ch.Amount
		}
	}
	return total
}

// outputStoreError marks an ai_outputs insert failure inside the settle
// transaction so the caller can map it (duplicate → 409, else 500) instead
// of reporting a settlement outage.
type outputStoreError struct{ err error }

func (e *outputStoreError) Error() string { return e.err.Error() }
func (e *outputStoreError) Unwrap() error { return e.err }
