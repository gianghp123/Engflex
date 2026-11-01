package responses

import "time"

// PremiumFeature is one entry in the admin-managed premium-gating registry.
type PremiumFeature struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
