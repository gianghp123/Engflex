package requests

import "engflex-api/internal/common"

// ListScenarios pages the scenario catalog, optionally topic-filtered.
type ListScenarios struct {
	common.ListParams `tstype:",extends"`
	TopicID           *string `form:"topicId" json:"topicId"`
	// Random orders by random() instead of newest-first — the home card picks
	// two surprises, no schema support needed.
	Random bool `form:"random" json:"random"`
}
