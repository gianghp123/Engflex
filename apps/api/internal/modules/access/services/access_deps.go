// Package services owns the access domain: the pure policy evaluator, the
// policy/usage services, and the paid-AI orchestration (spec §5.3).
package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
)

// PolicyDecider is the narrow policy surface consuming modules (Task 11
// voice, Task 12 AI review) program against: resolve the caller, evaluate an
// action, and report the quota location. *PolicyService satisfies it
// structurally. Defined here — not in each consumer — so there is one
// definition and one mocks dir.
type PolicyDecider interface {
	Principal(ctx context.Context, userID string, isAdmin bool) (Principal, TrialInfo, error)
	Decide(ctx context.Context, userID string, isAdmin bool, action Action) (Decision, error)
	CreditCost(action Action) int
	Location() *time.Location
}

// CreditCostReader is the narrow cost surface controllers use to stamp option
// DTOs without importing the policy service.
type CreditCostReader interface {
	CreditCost(action Action) int
}

// DefaultCreditCostReader serves the code fallback (route-table tests and a
// nil access bundle).
type DefaultCreditCostReader struct{}

func (DefaultCreditCostReader) CreditCost(action Action) int { return DefaultCreditCost(action) }

// UsageMeter is the narrow metering surface consuming modules program
// against: the full eleven-arg Reserve plus Commit/Release/RefundCommitted.
// CommitTx is Commit's transactional body over the caller's handle, so a
// state write and its settlement commit atomically. *UsageService satisfies
// it structurally.
type UsageMeter interface {
	Reserve(ctx context.Context, userID string, plan enums.PlanKey, action enums.AccessAction, charges []Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location) (*Reservation, error)
	Commit(ctx context.Context, reservationID string) error
	CommitTx(ctx context.Context, tx *gorm.DB, reservationID string) error
	Release(ctx context.Context, reservationID string) error
	RefundCommitted(ctx context.Context, userID, idempotencyBase string) error
}
