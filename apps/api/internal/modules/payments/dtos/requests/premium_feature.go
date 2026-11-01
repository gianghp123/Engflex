package requests

// CreatePremiumFeature registers a new named app capability in the
// premium-gating registry (admin-only).
type CreatePremiumFeature struct {
	Key         string `json:"key" binding:"required,max=100"`
	Label       string `json:"label" binding:"required,max=200"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	IsActive    bool   `json:"isActive"`
}

// UpdatePremiumFeature edits an existing feature entry.
type UpdatePremiumFeature struct {
	Label       string `json:"label" binding:"required,max=200"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	IsActive    bool   `json:"isActive"`
}
