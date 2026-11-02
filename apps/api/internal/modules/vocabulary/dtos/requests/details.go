package requests

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
