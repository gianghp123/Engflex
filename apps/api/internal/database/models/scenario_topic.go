package models

import "time"

// ScenarioTopic is the fixed, ordered banner taxonomy grouping roleplay
// scenarios ("Job interview simulations").
//
// Scenarios is the inverse of Scenario.Topic, attached by
// ListTopicsWithPreview in the same call that reads the topics. Like
// ConversationTurn.Feedback, the mapping is declared here in Go — the
// feedbacks-style polymorphic problem does not exist here (topic_id is a
// real foreign key), but an explicit tag still beats GORM's guessing, which
// would look for a ScenarioTopicID column that does not exist. It stays nil
// for a topic with no eligible scenario. Never set it on a topic you insert:
// scenarios are written through their own path.
type ScenarioTopic struct {
	ID        string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug      string      `gorm:"not null" json:"slug"`
	Name      string      `gorm:"not null" json:"name"`
	ShortName string      `gorm:"column:short_name;not null;default:''" json:"shortName"`
	Position  int         `gorm:"not null" json:"position"`
	Scenarios []*Scenario `gorm:"->;foreignKey:TopicID;references:ID" json:"scenarios,omitempty"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ScenarioTopic) TableName() string { return "scenario_topics" }
