package requests

import "engflex-api/internal/common/enums"

// ScenarioDetailInput mirrors models.ScenarioDetail field for field (Go
// names identical) so utils.Map copies it into the model without manual
// lines; only the JSON tags differ (camelCase wire vs the model's
// snake_case jsonb storage tags).
type ScenarioDetailInput struct {
	Role             string             `json:"role"`
	InterlocutorRole string             `json:"interlocutorRole"`
	GoalFormat       string             `json:"goalFormat"`
	Context          []string           `json:"context"`
	Opening          DetailOpeningInput `json:"opening"`
	Vocab            []string           `json:"vocab"`
}

type DetailOpeningInput struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// CreateScenario adds an admin-curated roleplay card. Always starts as
// draft (service-enforced); Status is not accepted here.
type CreateScenario struct {
	TopicID     string                   `json:"topicId" binding:"required,uuid"`
	PersonaID   *string                  `json:"personaId" binding:"omitempty,uuid"`
	Title       string                   `json:"title" binding:"required"`
	Objective   string                   `json:"objective"`
	CEFRLevel   enums.ScenarioDifficulty `json:"cefrLevel" binding:"required,oneof=B1+ B2 C1"`
	MaxDuration int                      `json:"maxDuration" binding:"required,min=1"`
	Details     ScenarioDetailInput      `json:"details"`
}

// UpdateScenario edits a scenario. Status only takes effect for the
// explicit publish/unpublish action (service-enforced, matching
// UpdateVocabularyItem/UpdateVideoExercise).
type UpdateScenario struct {
	TopicID     *string                   `json:"topicId" binding:"omitempty,uuid"`
	PersonaID   *string                   `json:"personaId" binding:"omitempty,uuid"`
	Title       *string                   `json:"title"`
	Objective   *string                   `json:"objective"`
	CEFRLevel   *enums.ScenarioDifficulty `json:"cefrLevel" binding:"omitempty,oneof=B1+ B2 C1"`
	MaxDuration *int                      `json:"maxDuration" binding:"omitempty,min=1"`
	Details     *ScenarioDetailInput      `json:"details"`
	Status      *enums.PublishStatus      `json:"status" binding:"omitempty,oneof=draft published"`
}

// CreateScenarioTopic / UpdateScenarioTopic manage the fixed banner
// taxonomy. Full replace, like video categories — topics have no draft
// concept of their own.
type CreateScenarioTopic struct {
	Slug      string `json:"slug" binding:"required"`
	Name      string `json:"name" binding:"required"`
	ShortName string `json:"shortName"`
	Position  int    `json:"position"`
}

type UpdateScenarioTopic struct {
	Slug      string `json:"slug" binding:"required"`
	Name      string `json:"name" binding:"required"`
	ShortName string `json:"shortName"`
	Position  int    `json:"position"`
}

// CreatePersona / UpdatePersona manage AI conversation partners.
type CreatePersona struct {
	Name        string       `json:"name" binding:"required"`
	RoleTitle   string       `json:"roleTitle" binding:"required"`
	Personality *string      `json:"personality"`
	Style       *string      `json:"style"`
	Objective   *string      `json:"objective"`
	Gender      enums.Gender `json:"gender" binding:"required,oneof=male female"`
	DefaultCEFR *enums.CEFR  `json:"defaultCefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
}

type UpdatePersona struct {
	Name        string       `json:"name" binding:"required"`
	RoleTitle   string       `json:"roleTitle" binding:"required"`
	Personality *string      `json:"personality"`
	Style       *string      `json:"style"`
	Objective   *string      `json:"objective"`
	Gender      enums.Gender `json:"gender" binding:"required,oneof=male female"`
	DefaultCEFR *enums.CEFR  `json:"defaultCefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
}
