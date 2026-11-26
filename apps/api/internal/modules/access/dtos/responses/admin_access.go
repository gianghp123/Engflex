// Package responses holds the access module's wire DTOs (tygo inputs).
package responses

import (
	"engflex-api/internal/database/models"
)

// AdminPlanLimits is the GET /admin/premium/limits body: the whole catalog.
type AdminPlanLimits []*models.PlanFeatureLimit

// UserUsage is the admin usage lookup (spec §8.6).
type UserUsage struct {
	Counters []*models.UsageCounter `json:"counters"`
	Events   []*models.UsageEvent   `json:"events"`
}

// TrialStatus is the POST /premium/trial and admin grant body: whether this
// call started a trial. Full trial detail stays on GET /premium/me.
type TrialStatus struct {
	Started bool `json:"started"`
}
