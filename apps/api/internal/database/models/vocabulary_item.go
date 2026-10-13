package models

import (
	"database/sql/driver"
	"time"

	"engflex-api/internal/common/enums"
)

// VocabularyItem is the shared dictionary entry (one row per term for all
// users). A non-null CreatedByUserID marks a user's custom word. details
// (jsonb) holds audio, syllables, stress, etymology, contexts, collocations,
// and word forms.
type VocabularyItem struct {
	ID              string                  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Term            string                  `gorm:"not null" json:"term"`
	Definition      string                  `gorm:"not null;default:''" json:"definition"`
	CEFR            enums.CEFR              `gorm:"column:cefr;not null" json:"cefr"`
	Domain          *enums.VocabularyDomain `gorm:"column:domain" json:"domain"`
	CreatedByUserID *string                 `json:"createdByUserId"`
	Status          enums.PublishStatus     `gorm:"not null;default:draft" json:"status"`
	Details         VocabularyDetails       `gorm:"type:jsonb;not null;default:'{}'" json:"details"`
	CreatedAt       time.Time               `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time               `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (VocabularyItem) TableName() string { return "vocabulary_items" }

// VocabularyContext is one labeled usage quote ("Webhook delivery...").
type VocabularyContext struct {
	Label string `json:"label"`
	Quote string `json:"quote"`
}

// Collocation is a fixed phrase with its pattern ("adv + adj").
type Collocation struct {
	Phrase  string `json:"phrase"`
	Pattern string `json:"pattern"`
	Example string `json:"example"`
}

// WordForm is one morphology entry (noun/adverb/antonym forms).
type WordForm struct {
	FormType string   `json:"formType"`
	Forms    []string `json:"forms"`
}

// VocabularyDetails is the typed view over vocabulary_items.details (jsonb):
// everything the word-detail screen renders beyond the core columns. It
// mirrors responses.VocabularyDetails field for field.
type VocabularyDetails struct {
	AudioUS        string              `json:"audioUs"`
	AudioUK        string              `json:"audioUk"`
	Syllables      string              `json:"syllables"`
	StressTip      string              `json:"stressTip"`
	CommonSlip     string              `json:"commonSlip"`
	LongDefinition string              `json:"longDefinition"`
	Etymology      string              `json:"etymology"`
	Contexts       []VocabularyContext `json:"contexts"`
	Collocations   []Collocation       `json:"collocations"`
	WordForms      []WordForm          `json:"wordForms"`
}

// GormDataType pins the column type for migrations handled outside GORM.
func (VocabularyDetails) GormDataType() string { return "jsonb" }

// Value marshals the document for writes.
func (d VocabularyDetails) Value() (driver.Value, error) { return jsonbValue(d) }

// Scan decodes the column on reads; a broken blob leaves the zero value.
func (d *VocabularyDetails) Scan(value any) error {
	jsonbScan(value, d)
	return nil
}
