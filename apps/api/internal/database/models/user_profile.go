package models

import (
	"database/sql/driver"
	"time"

	"engflex-api/internal/common/enums"
)

// UserProfile stores the onboarding profile for one Clerk subject. The
// surrogate id is the table's PK; user_id (Clerk subject) is unique.
// Other tables keep referencing users by user_id text with no FK.
type UserProfile struct {
	ID                    string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                string      `gorm:"not null;uniqueIndex" json:"userId"`
	Preferences           Preferences `gorm:"type:jsonb;not null;default:'{}'" json:"preferences"`
	OnboardingCompletedAt *time.Time  `json:"onboardingCompletedAt"`
	CreatedAt             time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt             time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserProfile) TableName() string { return "user_profiles" }

// Preferences is the typed view over user_profiles.preferences (jsonb). The
// CEFR range is derived from Level in Go, not stored. It mirrors
// responses.Preferences field for field.
type Preferences struct {
	Level              enums.Level           `json:"level"`
	Goals              []enums.Goal          `json:"goals"`
	DailyCommitmentMin int                   `json:"dailyCommitmentMin"`
	Topics             []enums.InterestTopic `json:"topics"`
}

// GormDataType pins the column type for migrations handled outside GORM.
func (Preferences) GormDataType() string { return "jsonb" }

// Value marshals the document for writes.
func (p Preferences) Value() (driver.Value, error) { return jsonbValue(p) }

// Scan decodes the column on reads; a broken blob leaves the zero value.
func (p *Preferences) Scan(value any) error {
	jsonbScan(value, p)
	return nil
}
