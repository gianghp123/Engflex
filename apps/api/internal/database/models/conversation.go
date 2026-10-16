package models

import (
	"time"

	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
)

// Conversation is one voice session (free talk or roleplay). Duration is
// derived from started_at/ended_at.
type Conversation struct {
	ID         string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string                 `gorm:"not null" json:"userId"`
	Mode       enums.ConversationMode `gorm:"not null" json:"mode"`
	ScenarioID *string                `gorm:"type:uuid" json:"scenarioId"`
	// Scenario is the practice scenario this conversation ran, filled by
	// Joins("Scenario") on read. "->" keeps it read-only: a conversation
	// write never auto-saves the scenario. Nil for free talk and for a
	// deleted scenario (scenario_id is ON DELETE SET NULL).
	Scenario            *Scenario                `gorm:"->;foreignKey:ScenarioID;references:ID" json:"-"`
	Status              enums.ConversationStatus `gorm:"not null" json:"status"`
	SpeechSessionID     string                   `gorm:"column:speech_session_id" json:"-"`
	SpeechStartResponse datatypes.JSON           `gorm:"column:speech_start_response;type:jsonb" json:"-"`
	DurationSec         *int                     `gorm:"column:duration_sec" json:"durationSec,omitempty"`
	StartedAt           time.Time                `gorm:"not null;default:now()" json:"startedAt"`
	EndedAt             *time.Time               `json:"endedAt"`
	CreatedAt           time.Time                `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time                `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Conversation) TableName() string { return "conversations" }
