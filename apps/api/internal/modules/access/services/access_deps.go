// Package services owns the access domain: the pure policy evaluator, the
// policy/usage services, and the paid-AI orchestration (spec §5.3).
package services

import (
	"context"
	"time"

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
	Location() *time.Location
}

// UsageMeter is the narrow metering surface consuming modules program
// against: the full ten-arg Reserve plus Commit/Release. *UsageService
// satisfies it structurally.
type UsageMeter interface {
	Reserve(ctx context.Context, userID string, plan enums.PlanKey, charges []Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location) (*Reservation, error)
	Commit(ctx context.Context, reservationID string) error
	Release(ctx context.Context, reservationID string) error
}
