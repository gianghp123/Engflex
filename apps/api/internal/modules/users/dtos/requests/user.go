package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// ListUsers pages and searches the Clerk user list. Search matches Clerk's
// own cross-field query (name, email, username, phone).
type ListUsers struct {
	common.ListParams `tstype:",extends"`
}

// UpdateUserRole sets the caller's public_metadata.role claim.
type UpdateUserRole struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

// ListUserAttempts pages and filters attempts for a target user.
type ListUserAttempts struct {
	common.ListParams `tstype:",extends"`
	Type              *enums.AttemptType `form:"type" json:"type" binding:"omitempty"`
}

type GrantSubscription struct {
	Plan enums.SubscriptionPlan `json:"plan" binding:"required,oneof=1m 3m 12m"`
	Note string                 `json:"note" binding:"omitempty,max=500"`
} // @name UserGrantSubscription
