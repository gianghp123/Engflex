package common

import (
	"context"
	"sync/atomic"
)

// contentGating is the process-wide CONTENT_PREMIUM_GATING switch, set once
// at boot from config and read on every content read. Option A ships it
// false: RequirePremiumAccess returns nil without touching the subscription
// lookup, so the flag also removes that lookup's failure mode from every
// content read.
var contentGating atomic.Bool

// SetContentGating records the boot-time gating switch. Called exactly once
// from main; tests may flip it and must restore it via t.Cleanup.
func SetContentGating(enabled bool) { contentGating.Store(enabled) }

// ContentGatingEnabled reports whether per-item premium gating is active.
func ContentGatingEnabled() bool { return contentGating.Load() }

// PremiumChecker is the minimal surface admin-curated content services
// need to gate a premium-only item. Defined here (not in the payments
// module) so every content module can depend on it without an inter-module
// import; services.SubscriptionService satisfies this structurally, no
// adapter needed — same shape as middleware.PremiumChecker.
type PremiumChecker interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
}

// RequirePremiumAccess enforces per-item premium gating: content flagged
// IsPremium is blocked for a caller who is neither an admin (preview) nor
// currently premium. Non-premium content and admin callers never touch the
// subscription lookup at all.
func RequirePremiumAccess(ctx context.Context, checker PremiumChecker, userID string, isAdmin, contentIsPremium bool) error {
	if !contentIsPremium || isAdmin {
		return nil
	}
	if !ContentGatingEnabled() {
		// Option A: static content is free for everyone (spec §5.4).
		return nil
	}
	ok, err := checker.IsPremium(ctx, userID)
	if err != nil {
		return Internal().withCause(err)
	}
	if !ok {
		return Forbidden("this content requires a premium subscription")
	}
	return nil
}
