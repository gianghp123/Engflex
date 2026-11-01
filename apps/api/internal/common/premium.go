package common

import "context"

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
	ok, err := checker.IsPremium(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return Forbidden("this content requires a premium subscription")
	}
	return nil
}
