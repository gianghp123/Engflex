package middleware

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
)

// PremiumChecker is the minimal surface RequirePremium needs. Defined here
// (not imported from the services package) so middleware stays the
// consumer-side interface; services.SubscriptionService satisfies it
// structurally, no adapter needed.
type PremiumChecker interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
}

// RequirePremium allows only callers with an active, non-expired
// subscription. Compose after RequireAuth. Unlike RequireAdmin, this is not
// a pure JWT-claim check — premium status lives in the subscriptions table,
// so the checker is queried per request. Build once per process with
// RequirePremium(subscriptionService) and mount the returned gin.HandlerFunc.
func RequirePremium(checker PremiumChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, err := checker.IsPremium(c.Request.Context(), UserID(c))
		if err != nil {
			common.Fail(c, common.Internal())
			c.Abort()
			return
		}
		if !ok {
			slog.WarnContext(c.Request.Context(), "request rejected: premium required",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			common.Fail(c, common.Forbidden("premium required"))
			c.Abort()
			return
		}
		c.Next()
	}
}
