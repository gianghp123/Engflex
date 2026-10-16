package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// ConversationSummary is one row in the history list: the conversation plus
// the scenario/partner the row renders. Turn data is detail-only.
type ConversationSummary struct {
	ID          string                   `json:"id"`
	Mode        enums.ConversationMode   `json:"mode"`
	Status      enums.ConversationStatus `json:"status"`
	ScenarioID  *string                  `json:"scenarioId"`
	StartedAt   time.Time                `json:"startedAt"`
	DurationSec *int                     `json:"durationSec,omitempty"`
	Scenario    *ScenarioSummary         `json:"scenario,omitempty"`
}

// ScenarioSummary is the scenario as a history row needs it — never the
// whole detail document.
type ScenarioSummary struct {
	Title   string          `json:"title"`
	Persona *PersonaSummary `json:"persona,omitempty"`
}

// PersonaSummary is the partner line ("Maya · Barista").
type PersonaSummary struct {
	Name      string `json:"name"`
	RoleTitle string `json:"roleTitle"`
}
