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
	IPA            string              `json:"ipa"`
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
	ItemID         string                 `json:"itemId"`
	SourceType     enums.VocabularySource `json:"sourceType"`
	SourceID       *string                `json:"sourceId"`
	Note           *string                `json:"note"`
	Mastered       bool                   `json:"mastered"`
	SRSDueAt       *time.Time             `json:"srsDueAt"`
	IntervalDays   int                    `json:"intervalDays"`
	ReviewCount    int                    `json:"reviewCount"`
	LastReviewedAt *time.Time             `json:"lastReviewedAt"`
	CreatedAt      time.Time              `json:"createdAt"`
}

// VocabularyItem is a vocabulary hub card / word detail payload.
type VocabularyItem struct {
	ID           string                        `json:"id"`
	Term         string                        `json:"term"`
	CEFR         enums.CEFR                    `json:"cefr"`
	PartOfSpeech *enums.VocabularyPartOfSpeech `json:"partOfSpeech"`
	Definition   string                        `json:"definition"`
	Details      *VocabularyDetails            `json:"details"`
	UserState    *UserVocabularyState          `json:"userState"`
}

// VocabularyStats aggregates the hub's stat cards.
type VocabularyStats struct {
	TotalSaved      int `json:"totalSaved"`
	Mastered        int `json:"mastered"`
	NeedsReview     int `json:"needsReview"`
	CustomAdditions int `json:"customAdditions"`
	DueToday        int `json:"dueToday"`
	// SavedToday counts words saved in the caller's local day (the request's
	// dayStart, or the UTC day when it is omitted).
	SavedToday int `json:"savedToday"`
}
