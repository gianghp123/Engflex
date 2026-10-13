package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

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

// VocabularyDetails mirrors vocabulary_items.details (jsonb): everything the
// word-detail screen renders beyond the core columns.
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

// UserVocabularyState is the user's row over a shared item. ItemID is the
// shared item (needed to correlate, display, and DELETE the row).
type UserVocabularyState struct {
	ItemID     string                 `json:"itemId"`
	SourceType enums.VocabularySource `json:"sourceType"`
	SourceID   *string                `json:"sourceId"`
	Note       *string                `json:"note"`
	Mastered   bool                   `json:"mastered"`
	SRSDueAt   *time.Time             `json:"srsDueAt"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// VocabularyItem is a vocabulary hub card / word detail payload.
type VocabularyItem struct {
	ID           string                  `json:"id"`
	Term         string                  `json:"term"`
	IPA          string                  `json:"ipa"`
	PartOfSpeech string                  `json:"partOfSpeech"`
	CEFR         enums.CEFR              `json:"cefr"`
	Domain       *enums.VocabularyDomain `json:"domain"`
	Definition   string                  `json:"definition"`
	Status       enums.PublishStatus     `json:"status"`
	Details      *VocabularyDetails      `json:"details"`
	UserState    *UserVocabularyState    `json:"userState"`
}

// VocabularyStats aggregates the hub's four stat cards (computed by
// services from user_vocabulary).
type VocabularyStats struct {
	TotalSaved      int `json:"totalSaved"`
	Mastered        int `json:"mastered"`
	NeedsReview     int `json:"needsReview"`
	CustomAdditions int `json:"customAdditions"`
	DueToday        int `json:"dueToday"`
	RetentionPct    int `json:"retentionPct"`
}
