// Package requests holds the access module's wire requests.
package requests

import (
	"engflex-api/internal/common/enums"
)

// PlanLimitInput is one counter's limit for one plan.
type PlanLimitInput struct {
	CounterKey enums.CounterKind `json:"counterKey" binding:"required,oneof=ai_credit voice_session"`
	IsEnabled  bool              `json:"isEnabled"`
	LimitValue int               `json:"limitValue" binding:"min=0,max=100000"`
	Period     enums.UsagePeriod `json:"period" binding:"required,oneof=day month total"`
}

// UpdatePlanLimits batch-upserts one plan's limits.
type UpdatePlanLimits struct {
	Counters []PlanLimitInput `json:"counters" binding:"required,dive"`
}
