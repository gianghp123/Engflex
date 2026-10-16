package requests

import "engflex-api/internal/common"

// ListUsers pages and searches the Clerk user list. Search matches Clerk's
// own cross-field query (name, email, username, phone).
type ListUsers struct {
	common.ListParams `tstype:",extends"`
}

// UpdateUserRole sets the caller's public_metadata.role claim.
type UpdateUserRole struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}
