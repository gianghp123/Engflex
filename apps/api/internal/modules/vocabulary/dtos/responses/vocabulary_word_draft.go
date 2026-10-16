package responses

// WordDraft is the generated draft the add-word dialog fills itself with.
// Details is the same wire shape as a saved item's details, minus audio — an
// LLM cannot produce audio, so those fields stay empty.
type WordDraft struct {
	Definition   string            `json:"definition"`
	CEFR         string            `json:"cefr"`
	PartOfSpeech string            `json:"partOfSpeech"`
	Details      VocabularyDetails `json:"details"`
}
