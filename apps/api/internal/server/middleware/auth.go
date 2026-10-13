// Package middleware holds shared gin middleware.
package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/logger"
)

// userIDKey is the gin context key holding the authenticated user ID
// (Clerk session subject). Canonical name is userID everywhere:
// context key, logs (user_id), DB user_id columns, DTOs.
const userIDKey = "userID"

// roleKey is the gin context key holding the caller's role (Clerk
// public-metadata `role` claim). Fail-closed: absent means user.
const roleKey = "role"

// clerkRoleClaims is the TOP-LEVEL token shape the middleware honors:
// {"role": "admin"}. Clerk only puts this in the JWT if the Dashboard
// Sessions → Customize session token editor maps it, e.g.
// {"role": "{{user.public_metadata.role}}"}. The nested
// {"metadata": {"role": ...}} shape is NOT honored (see TestRoleFromClaims).
// Only "admin" is honored; everything else resolves to UserRoleUser.
type clerkRoleClaims struct {
	Role string `json:"role"`
}

// parseRole is a strict allowlist: exact "admin" or user.
func parseRole(v string) enums.UserRole {
	if v == string(enums.UserRoleAdmin) {
		return enums.UserRoleAdmin
	}
	return enums.UserRoleUser
}

// roleFromClaims resolves the caller role from verified custom claims.
// Anything unexpected (nil, wrong type, nil pointer, unknown string)
// fails closed to UserRoleUser.
func roleFromClaims(custom any) enums.UserRole {
	if cc, ok := custom.(*clerkRoleClaims); ok && cc != nil {
		return parseRole(cc.Role)
	}
	return enums.UserRoleUser
}

// noopWriter discards writes: WithHeaderAuthorization never writes, it only
// enriches the request context, so nothing is lost in the gin adaptation.
type noopWriter struct{ header http.Header }

func (w noopWriter) Header() http.Header       { return w.header }
func (noopWriter) Write(b []byte) (int, error) { return len(b), nil }
func (noopWriter) WriteHeader(int)             {}

// RequireAuth verifies the Clerk Bearer session token and stores the userID
// (token subject) in the gin context. Requests without valid claims get the
// standard {message} 401 and the chain stops.
func RequireAuth() gin.HandlerFunc {
	verifier := clerkhttp.WithHeaderAuthorization(
		clerkhttp.CustomClaimsConstructor(func(context.Context) any {
			return &clerkRoleClaims{}
		}),
	)
	return func(c *gin.Context) {
		verifier(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			c.Request = r
		})).ServeHTTP(noopWriter{header: http.Header{}}, c.Request)

		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		if !ok || claims.Subject == "" {
			slog.WarnContext(c.Request.Context(), "request rejected: unauthenticated",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			common.Fail(c, common.Unauthorized("sign in required"))
			c.Abort()
			return
		}
		c.Set(userIDKey, claims.Subject)
		c.Set(roleKey, roleFromClaims(claims.Custom))
		// Attribute Sentry events to the caller. Nil-guarded like RequestID:
		// no hub without the Sentry middleware, and Fail-closed Role stays
		// untouched — this is telemetry, not authorization.
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			hub.Scope().SetUser(sentry.User{ID: claims.Subject})
		}
		// Stamp user_id on downstream logs (services, request logger).
		c.Request = c.Request.WithContext(
			logger.WithUserID(c.Request.Context(), claims.Subject),
		)
		c.Next()
	}
}

// UserID returns the authenticated userID, or "" when absent
// (controllers behind RequireAuth always have one).
func UserID(c *gin.Context) string {
	id, _ := c.Get(userIDKey)
	sub, _ := id.(string)
	return sub
}

// Role returns the caller's role, defaulting to UserRoleUser when absent
// (fail-closed: a gate without authentication context denies).
func Role(c *gin.Context) enums.UserRole {
	if v, ok := c.Get(roleKey); ok {
		if r, ok := v.(enums.UserRole); ok {
			return r
		}
	}
	return enums.UserRoleUser
}

// RequireAdmin allows only admin-role callers. Compose after RequireAuth;
// without it Role defaults to user and every request 403s (fail-closed).
// No route mounts this yet — first consumer is the videos authoring
// endpoints after the dev merge (see Task 5 of the policy plan).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if Role(c) != enums.UserRoleAdmin {
			slog.WarnContext(c.Request.Context(), "request rejected: admin required",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			common.Fail(c, common.Forbidden("admin required"))
			c.Abort()
			return
		}
		c.Next()
	}
}
