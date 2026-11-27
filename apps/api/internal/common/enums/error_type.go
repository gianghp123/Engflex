package enums

// ErrorType is the learner-facing grammar/lexis error taxonomy carried on
// every feedback span (conversation AnalysisSpan, writing Span). It answers
// "what am I weak at" for the dashboard aggregation (Plan 01, GAP-01).
type ErrorType string

const (
	ErrorTypeVerbTense     ErrorType = "verb_tense"
	ErrorTypeSubjectVerb   ErrorType = "subject_verb_agreement"
	ErrorTypeArticle       ErrorType = "article"
	ErrorTypePreposition   ErrorType = "preposition"
	ErrorTypePlural        ErrorType = "plural"
	ErrorTypeWordOrder     ErrorType = "word_order"
	ErrorTypeSpelling      ErrorType = "spelling"
	ErrorTypePunctuation   ErrorType = "punctuation"
	ErrorTypeLexicalChoice ErrorType = "lexical_choice"
	ErrorTypeCollocation   ErrorType = "collocation"
	ErrorTypeOther         ErrorType = "other"
)

// Valid reports whether the value is a known taxonomy member.
func (e ErrorType) Valid() bool {
	switch e {
	case ErrorTypeVerbTense, ErrorTypeSubjectVerb, ErrorTypeArticle,
		ErrorTypePreposition, ErrorTypePlural, ErrorTypeWordOrder,
		ErrorTypeSpelling, ErrorTypePunctuation, ErrorTypeLexicalChoice,
		ErrorTypeCollocation, ErrorTypeOther:
		return true
	}
	return false
}
