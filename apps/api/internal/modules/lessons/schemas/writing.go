// Package schemas holds the writing-score LLM output shape and its
// validation. The JSON schema map is generated from the struct so the two
// can never drift.
package schemas

import (
	"fmt"
	"strings"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/llm"
)

// Span is one quoted problem with a minimal fix. The array it sits in is the
// label (grammar = real error, phrasing = awkward only) — there is no status
// field to drift.
type Span struct {
	Text       string          `json:"text" jsonschema_description:"Exact problematic substring"`
	Occurrence int             `json:"occurrence" jsonschema_description:"1-based occurrence"`
	ErrorType  enums.ErrorType `json:"errorType" jsonschema_description:"Error taxonomy tag"`
	Correction string          `json:"correction" jsonschema_description:"Fix for this span only"`
	Reason     string          `json:"reason" jsonschema_description:"Brief why"`
}

type TaskVerdict struct {
	Item     string `json:"item" jsonschema_description:"Checklist item judged"`
	Status   string `json:"status" jsonschema:"enum=covered,enum=partial,enum=missing" jsonschema_description:"Coverage verdict"`
	Evidence string `json:"evidence" jsonschema_description:"Exact learner-text quote, or empty when missing"`
}

// TaskCompletion is section 1: did the response do the task? Counts live in
// the service-built summary, never here (LLMs cannot count).
type TaskCompletion struct {
	Verdicts      []TaskVerdict `json:"verdicts" jsonschema_description:"One verdict per checklist item, same order"`
	AnswersPrompt bool          `json:"answersPrompt" jsonschema_description:"Whether the response answers the prompt"`
	AnswersNote   string        `json:"answersNote" jsonschema_description:"Short why or what is missing"`
}

// WritingFeedback is the full LLM output: six sections, no counts of its
// own, no copy of the learner paragraph.
type WritingFeedback struct {
	Task        TaskCompletion `json:"task" jsonschema_description:"Task completion verdicts"`
	Grammar     []Span         `json:"grammar" jsonschema_description:"Real grammar or spelling errors only"`
	Phrasing    []Span         `json:"phrasing" jsonschema_description:"Grammatical but unnatural phrasing only"`
	Expressions []string       `json:"expressions" jsonschema_description:"3-5 reusable task-relevant patterns"`
	Suggested   string         `json:"suggested" jsonschema_description:"Complete improved answer in word range"`
	Tip         string         `json:"tip" jsonschema_description:"One actionable recommendation"`
}

var WritingFeedbackSchemaName = "WritingFeedback"

// WritingFeedbackSchema is the strict json_schema map for the LLM call,
// via the shared generator.
var WritingFeedbackSchema = llm.GenerateSchema[WritingFeedback]()

// NormalizeWritingFeedback validates one decoded LLM reply: non-empty
// suggested and tip, every grammar/phrasing span an exact substring,
// exactly one verdict per checklist item with known statuses and substring
// evidence, at least one expression. Anything else is rejected — the caller
// maps the error to 503 and stores nothing.
func NormalizeWritingFeedback(out WritingFeedback, text string, checklistLen int) (WritingFeedback, error) {
	if strings.TrimSpace(out.Suggested) == "" || strings.TrimSpace(out.Tip) == "" {
		return WritingFeedback{}, fmt.Errorf("feedback is missing suggested or tip")
	}
	for i, sp := range out.Grammar {
		if sp.Text == "" || !strings.Contains(text, sp.Text) || sp.Occurrence < 1 {
			return WritingFeedback{}, fmt.Errorf("grammar span %d is not an exact substring", i)
		}
		if out.Grammar[i].ErrorType == "" {
			out.Grammar[i].ErrorType = enums.ErrorTypeOther
		}
		if !out.Grammar[i].ErrorType.Valid() {
			return WritingFeedback{}, fmt.Errorf("grammar span %d has unknown error type", i)
		}
	}
	for i, sp := range out.Phrasing {
		if sp.Text == "" || !strings.Contains(text, sp.Text) || sp.Occurrence < 1 {
			return WritingFeedback{}, fmt.Errorf("phrasing span %d is not an exact substring", i)
		}
		if out.Phrasing[i].ErrorType == "" {
			out.Phrasing[i].ErrorType = enums.ErrorTypeOther
		}
		if !out.Phrasing[i].ErrorType.Valid() {
			return WritingFeedback{}, fmt.Errorf("phrasing span %d has unknown error type", i)
		}
	}
	if len(out.Task.Verdicts) != checklistLen {
		return WritingFeedback{}, fmt.Errorf("verdicts %d != checklist %d", len(out.Task.Verdicts), checklistLen)
	}
	for i, v := range out.Task.Verdicts {
		if v.Status != "covered" && v.Status != "partial" && v.Status != "missing" {
			return WritingFeedback{}, fmt.Errorf("verdict %d has unknown status %q", i, v.Status)
		}
		if v.Evidence != "" && !strings.Contains(text, v.Evidence) {
			return WritingFeedback{}, fmt.Errorf("verdict %d evidence is not a substring", i)
		}
	}
	if len(out.Expressions) < 1 {
		return WritingFeedback{}, fmt.Errorf("feedback has no expressions")
	}
	if out.Grammar == nil {
		out.Grammar = []Span{}
	}
	if out.Phrasing == nil {
		out.Phrasing = []Span{}
	}
	return out, nil
}
