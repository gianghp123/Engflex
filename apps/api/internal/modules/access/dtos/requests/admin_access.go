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

// ActionCostInput is one paid action's credit price.
type ActionCostInput struct {
	Action  enums.AccessAction `json:"action" binding:"required,oneof=voice_session_start exam_ai_review study_plan"`
	Credits int                `json:"credits" binding:"min=0,max=100000"`
}

// UpdateActionCosts batch-upserts all paid action costs.
type UpdateActionCosts struct {
	Costs []ActionCostInput `json:"costs" binding:"required,len=3,dive"`
}

// UpdateTrialTerms rewrites the trial duration and both trial caps.
type UpdateTrialTerms struct {
	Days          int `json:"days" binding:"required,min=1,max=365"`
	Credits       int `json:"credits" binding:"min=0,max=100000"`
	VoiceSessions int `json:"voiceSessions" binding:"min=0,max=100000"`
}
