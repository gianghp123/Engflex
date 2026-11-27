package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// AdminActionCost is one paid action's credit price for the admin table.
type AdminActionCost struct {
	Action    enums.AccessAction `json:"action"`
	Credits   int                `json:"credits"`
	UpdatedAt time.Time          `json:"updatedAt"`
	UpdatedBy string             `json:"updatedBy"`
}

// AdminActionCosts is the GET /admin/premium/action-costs body.
type AdminActionCosts []AdminActionCost

// AdminTrialTerms is the GET /admin/premium/trial-terms body.
type AdminTrialTerms struct {
	Days          int       `json:"days"`
	Credits       int       `json:"credits"`
	VoiceSessions int       `json:"voiceSessions"`
	UpdatedAt     time.Time `json:"updatedAt"`
	UpdatedBy     string    `json:"updatedBy"`
}
