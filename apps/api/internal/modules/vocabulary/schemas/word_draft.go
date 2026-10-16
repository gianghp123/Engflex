// Package schemas holds the word-draft LLM output shape and its validation.
// The JSON schema map is generated from the struct so the two can never drift.
package schemas

import (
	"fmt"
	"strings"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/llm"
)

const (
	wordDraftContextsCap     = 4
	wordDraftCollocationsCap = 6
	wordDraftFormsCap        = 4
	wordDraftFieldCap        = 400
)

// WordDraftContext is one example sentence with a short label.
type WordDraftContext struct {
	Label string `json:"label" jsonschema_description:"Short label, e.g. \"At work\" or \"Everyday\""`
	Quote string `json:"quote" jsonschema_description:"The full example sentence"`
}

// WordDraftCollocation is a fixed phrase with its pattern.
type WordDraftCollocation struct {
	Phrase  string `json:"phrase" jsonschema_description:"The collocation, e.g. \"a punctual start\""`
	Pattern string `json:"pattern" jsonschema_description:"Grammar pattern, e.g. \"adj + noun\""`
	Example string `json:"example" jsonschema_description:"One short example sentence"`
}

// WordDraftWordForm is one morphology entry (plural, past tense, antonym...).
type WordDraftWordForm struct {
	FormType string   `json:"formType" jsonschema_description:"Relationship, e.g. plural, past tense, noun form, antonym"`
	Forms    []string `json:"forms" jsonschema_description:"The forms themselves"`
}

// WordDraftDetails mirrors the rich details the word-detail screen renders,
// minus the audio fields — an LLM cannot produce audio (they stay empty).
type WordDraftDetails struct {
	IPA            string                 `json:"ipa" jsonschema_description:"IPA transcription without slashes, e.g. ˈpʌŋk.tʃu.əl"`
	Syllables      string                 `json:"syllables" jsonschema_description:"Syllables separated by dots, e.g. punc.tu.al"`
	StressTip      string                 `json:"stressTip" jsonschema_description:"One-line stress tip for a Vietnamese learner"`
	CommonSlip     string                 `json:"commonSlip" jsonschema_description:"The mistake learners commonly make with this word"`
	LongDefinition string                 `json:"longDefinition" jsonschema_description:"A fuller definition, 2-3 sentences"`
	Etymology      string                 `json:"etymology" jsonschema_description:"Word origin; empty string when unsure — never invent one"`
	Contexts       []WordDraftContext     `json:"contexts" jsonschema_description:"2-4 complete example sentences"`
	Collocations   []WordDraftCollocation `json:"collocations" jsonschema_description:"3-6 phrases that commonly go with the word"`
	WordForms      []WordDraftWordForm    `json:"wordForms" jsonschema_description:"Up to 4 form relationships (plural, past tense, noun form, antonym)"`
}

// WordDraft is the full LLM output: exactly the fields the add-word dialog
// fills itself with.
type WordDraft struct {
	Definition   string           `json:"definition" jsonschema_description:"One learner-friendly sentence defining the term"`
	CEFR         string           `json:"cefr" jsonschema:"enum=A1,enum=A2,enum=B1,enum=B2,enum=C1,enum=C2" jsonschema_description:"CEFR level estimate"`
	PartOfSpeech string           `json:"partOfSpeech" jsonschema:"enum=noun,enum=verb,enum=adjective,enum=adverb,enum=pronoun,enum=preposition,enum=conjunction,enum=interjection,enum=determiner,enum=phrase,enum=idiom,enum=phrasal_verb" jsonschema_description:"The term's main part of speech in this sense"`
	Details      WordDraftDetails `json:"details" jsonschema_description:"Rich word details"`
}

// WordDraftSchemaName is the schema label sent to the engine.
const WordDraftSchemaName = "word_draft"

// WordDraftSchema is generated from WordDraft; drift is impossible.
var WordDraftSchema = llm.GenerateSchema[WordDraft]()

// NormalizeWordDraft validates and tidies the model output. Garbage CEFR or an
// empty definition fails the draft (503 at the boundary); a bad part of speech
// is dropped; empty rows are removed and arrays are capped so one verbose
// answer cannot bloat the saved document.
func NormalizeWordDraft(out WordDraft) (WordDraft, error) {
	out.Definition = strings.TrimSpace(out.Definition)
	out.CEFR = strings.TrimSpace(out.CEFR)
	out.PartOfSpeech = strings.TrimSpace(out.PartOfSpeech)
	if out.Definition == "" {
		return out, fmt.Errorf("definition is empty")
	}
	switch enums.CEFR(out.CEFR) {
	case enums.CEFRA1, enums.CEFRA2, enums.CEFRB1, enums.CEFRB2, enums.CEFRC1, enums.CEFRC2:
	default:
		return out, fmt.Errorf("cefr %q is not a level", out.CEFR)
	}
	if !validPartOfSpeech(out.PartOfSpeech) {
		out.PartOfSpeech = ""
	}
	out.Details = normalizeDetails(out.Details)
	return out, nil
}

func validPartOfSpeech(pos string) bool {
	switch enums.VocabularyPartOfSpeech(pos) {
	case enums.VocabularyPartOfSpeechNoun, enums.VocabularyPartOfSpeechVerb,
		enums.VocabularyPartOfSpeechAdjective, enums.VocabularyPartOfSpeechAdverb,
		enums.VocabularyPartOfSpeechPronoun, enums.VocabularyPartOfSpeechPreposition,
		enums.VocabularyPartOfSpeechConjunction, enums.VocabularyPartOfSpeechInterjection,
		enums.VocabularyPartOfSpeechDeterminer, enums.VocabularyPartOfSpeechPhrase,
		enums.VocabularyPartOfSpeechIdiom, enums.VocabularyPartOfSpeechPhrasalVerb:
		return true
	}
	return false
}

func normalizeDetails(d WordDraftDetails) WordDraftDetails {
	d.IPA = capText(d.IPA)
	d.Syllables = capText(d.Syllables)
	d.StressTip = capText(d.StressTip)
	d.CommonSlip = capText(d.CommonSlip)
	d.LongDefinition = capText(d.LongDefinition)
	d.Etymology = capText(d.Etymology)

	contexts := make([]WordDraftContext, 0, len(d.Contexts))
	for _, c := range d.Contexts {
		c.Label = capText(c.Label)
		c.Quote = capText(c.Quote)
		if c.Quote == "" {
			continue
		}
		contexts = append(contexts, c)
		if len(contexts) == wordDraftContextsCap {
			break
		}
	}
	d.Contexts = contexts

	collocations := make([]WordDraftCollocation, 0, len(d.Collocations))
	for _, c := range d.Collocations {
		c.Phrase = capText(c.Phrase)
		c.Pattern = capText(c.Pattern)
		c.Example = capText(c.Example)
		if c.Phrase == "" {
			continue
		}
		collocations = append(collocations, c)
		if len(collocations) == wordDraftCollocationsCap {
			break
		}
	}
	d.Collocations = collocations

	forms := make([]WordDraftWordForm, 0, len(d.WordForms))
	for _, f := range d.WordForms {
		f.FormType = capText(f.FormType)
		kept := make([]string, 0, len(f.Forms))
		for _, form := range f.Forms {
			if form = capText(form); form != "" {
				kept = append(kept, form)
			}
		}
		if f.FormType == "" || len(kept) == 0 {
			continue
		}
		f.Forms = kept
		forms = append(forms, f)
		if len(forms) == wordDraftFormsCap {
			break
		}
	}
	d.WordForms = forms

	return d
}

func capText(s string) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > wordDraftFieldCap {
		return string(runes[:wordDraftFieldCap])
	}
	return s
}
