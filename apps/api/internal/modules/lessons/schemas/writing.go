// Package schemas holds the writing-score LLM output shape and its
// validation. The JSON schema map is generated from the struct so the two
// can never drift.
package schemas

import (
	"fmt"
	"strings"

	"engflex-api/internal/database/models"
	"engflex-api/internal/llm"
)

type Span = models.WritingSpan
type TaskVerdict = models.WritingTaskVerdict
type TaskCompletion = models.WritingTaskCompletion
type WritingFeedback = models.WritingFeedback

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
	}
	for i, sp := range out.Phrasing {
		if sp.Text == "" || !strings.Contains(text, sp.Text) || sp.Occurrence < 1 {
			return WritingFeedback{}, fmt.Errorf("phrasing span %d is not an exact substring", i)
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
