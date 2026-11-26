package models

import (
	"time"

	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
)

// AIOutput is one cached paid AI result (spec §6.4). The
// (subject_type, subject_id, kind) triple is unique: re-opening a
// generated review or plan is free — the charge happened once.
type AIOutput struct {
	ID             string             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         string             `gorm:"not null" json:"userId"`
	SubjectType    enums.UsageRefType `gorm:"not null" json:"subjectType"`
	SubjectID      string             `gorm:"not null" json:"subjectId"`
	Kind           enums.AIOutputKind `gorm:"not null" json:"kind"`
	SchemaVersion  int                `gorm:"not null;default:1" json:"schemaVersion"`
	Payload        datatypes.JSON     `gorm:"type:jsonb;not null" json:"payload"`
	CreditsCharged int                `gorm:"not null;default:0" json:"creditsCharged"`
	CreatedAt      time.Time          `gorm:"autoCreateTime" json:"createdAt"`
}

func (AIOutput) TableName() string { return "ai_outputs" }
