package responses

import "engflex-api/internal/common/enums"

// Scenario is one roleplay practice card. Field names mirror the model so
// utils.Map copies everything automatically — including the nested Details
// blob, whose struct shape matches models.ScenarioDetail field for field.
// A non-null Persona marks curated scenarios.
type Scenario struct {
	ID          string                   `json:"id"`
	Topic       *ScenarioTopic           `json:"topic,omitempty"`
	Persona     *Persona                 `json:"persona"`
	Title       string                   `json:"title"`
	Objective   string                   `json:"objective"`
	CEFRLevel   enums.ScenarioDifficulty `json:"cefrLevel"`
	MaxDuration int                      `json:"maxDuration"`
	Details     ScenarioDetail           `json:"details"`
	Status      enums.PublishStatus      `json:"status"`
}

// ScenarioDetail mirrors models.ScenarioDetail field for field so copier
// maps the blob without manual lines. Pure scenario substance (context,
// opening line, terms); nothing coach-generic lives here.
type ScenarioDetail struct {
	Role             string        `json:"role,omitempty"`
	InterlocutorRole string        `json:"interlocutorRole,omitempty"`
	GoalFormat       string        `json:"goalFormat,omitempty"`
	Context          []string      `json:"context,omitempty"`
	Opening          DetailOpening `json:"opening,omitempty"`
	Vocab            []string      `json:"vocab,omitempty"`
}

// DetailOpening is the partner's first line on the detail page.
type DetailOpening struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// ScenarioTopic mirrors models.ScenarioTopic: the banner taxonomy plus the
// top-k scenarios the repository attached. utils.MapSlice copies the whole
// banner — nested scenarios, their personas, and their detail documents — in
// one call, because every nested type mirrors its model by field name.
type ScenarioTopic struct {
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	Name      string     `json:"name"`
	ShortName string     `json:"shortName"`
	Position  int        `json:"position"`
	Scenarios []Scenario `json:"scenarios"`
}
