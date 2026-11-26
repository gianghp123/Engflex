package services

import (
	"context"
	"fmt"
	"strings"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/ielts"
	"engflex-api/internal/exams/toeic"
)

// ExamMistake is one wrong answer the review explains. Keys are opaque
// question identifiers; the prose is English-only (no user-facing copy).
type ExamMistake struct {
	QuestionID  string  `json:"questionId"`
	SelectedKey *string `json:"selectedKey,omitempty"`
	CorrectKey  string  `json:"correctKey"`
	Explanation string  `json:"explanation"`
	Stem        string  `json:"stem,omitempty"`
}

// ExamReviewOutput is the structured review document the LLM returns. The
// zero value marshals to valid JSON so an empty attempt or undecodable
// content still yields a storable payload.
type ExamReviewOutput struct {
	Summary  string        `json:"summary"`
	Mistakes []ExamMistake `json:"mistakes"`
}

// StudyPlanFocusArea names one weak area with a short remediation note.
type StudyPlanFocusArea struct {
	Area   string `json:"area"`
	Detail string `json:"detail"`
}

// StudyPlanItem is one suggested study step.
type StudyPlanItem struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// StudyPlanOutput is the structured study-plan document the LLM returns.
type StudyPlanOutput struct {
	Summary    string               `json:"summary"`
	FocusAreas []StudyPlanFocusArea `json:"focusAreas"`
	Plan       []StudyPlanItem      `json:"plan"`
}

// MaxReviewItems caps the wrong-answer detail sent to the LLM so the prompt
// cannot grow unbounded on a full 200-question exam.
const MaxReviewItems = 20

// ExamReviewSchema is the strict JSON schema for the review output.
func ExamReviewSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{"type": "string"},
			"mistakes": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"questionId":  map[string]any{"type": "string"},
						"selectedKey": map[string]any{"type": "string"},
						"correctKey":  map[string]any{"type": "string"},
						"explanation": map[string]any{"type": "string"},
						"stem":        map[string]any{"type": "string"},
					},
					"required":             []string{"questionId", "correctKey", "explanation"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"summary", "mistakes"},
		"additionalProperties": false,
	}
}

// StudyPlanSchema is the strict JSON schema for the study-plan output.
func StudyPlanSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{"type": "string"},
			"focusAreas": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"area":   map[string]any{"type": "string"},
						"detail": map[string]any{"type": "string"},
					},
					"required":             []string{"area", "detail"},
					"additionalProperties": false,
				},
			},
			"plan": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":  map[string]any{"type": "string"},
						"detail": map[string]any{"type": "string"},
					},
					"required":             []string{"title", "detail"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"summary", "focusAreas", "plan"},
		"additionalProperties": false,
	}
}

// buildReviewPrompt assembles the mistake-review prompt from the attempt's
// wrong answers plus the exam content. Empty or undecodable content is
// tolerated: the prompt degrades to attempt-level summary lines and the
// output stays a zero-valued ExamReviewOutput that marshals.
func buildReviewPrompt(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) (string, any, string, map[string]any, error) {
	out := &ExamReviewOutput{}
	items := reviewLines(ctx, tests, m)
	var b strings.Builder
	b.WriteString("You are an English exam coach. Review the learner's completed exam attempt and explain each mistake briefly in English.\n")
	if len(items) == 0 {
		b.WriteString("No per-question mistake detail is available. Summarize the overall result and give general study advice.\n")
	} else {
		b.WriteString(fmt.Sprintf("Wrong answers (up to %d shown):\n", MaxReviewItems))
		for _, line := range items {
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("For each listed question, explain why the correct answer is right and the learner's choice is wrong in one sentence.\n")
	}
	b.WriteString("Return JSON matching the exam_review schema with a short summary and a mistakes array.\n")
	return b.String(), out, "exam_review", ExamReviewSchema(), nil
}

// buildPlanPrompt assembles the study-plan prompt from the same data at
// section-level granularity.
func buildPlanPrompt(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) (string, any, string, map[string]any, error) {
	out := &StudyPlanOutput{}
	summary := planSummaryLines(ctx, tests, m)
	var b strings.Builder
	b.WriteString("You are an English study planner. Build a short remedial study plan from the learner's completed exam attempt.\n")
	if len(summary) == 0 {
		b.WriteString("No section-level detail is available. Propose a general balanced plan covering listening and reading.\n")
	} else {
		b.WriteString("Weak areas observed:\n")
		for _, line := range summary {
			b.WriteString("- " + line + "\n")
		}
		b.WriteString("Link free lessons or videos by topic name where relevant.\n")
	}
	b.WriteString("Return JSON matching the study_plan schema with a summary, focusAreas, and a plan of concrete steps.\n")
	return b.String(), out, "study_plan", StudyPlanSchema(), nil
}

// reviewLines returns up to MaxReviewItems human-readable wrong-answer
// lines. Content or result parse failures degrade to generic answer lines
// (or nothing) instead of an error.
func reviewLines(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) []string {
	if m == nil {
		return nil
	}
	if res, err := toeic.ParseResult(m.Result); err == nil && res != nil {
		return toeicReviewLines(ctx, tests, m, res)
	}
	if res, err := ielts.ParseResult(m.Result); err == nil && res != nil {
		return ieltsReviewLines(ctx, tests, m, res)
	}
	return nil
}

// toeicReviewLines matches wrong TOEIC answers against the content key
// when the content decodes; otherwise it lists raw selections.
func toeicReviewLines(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt, res *toeic.Result) []string {
	byQuestion := map[string]toeic.Question{}
	if content := loadToeicContent(ctx, tests, m); content != nil {
		for _, p := range content.Parts {
			for _, item := range p.Items {
				for _, q := range item.Questions {
					byQuestion[q.ID] = q
				}
			}
		}
	}
	var lines []string
	for _, a := range res.Answers {
		q, ok := byQuestion[a.QuestionID]
		if !ok {
			continue
		}
		if q.CorrectKey == nil {
			continue
		}
		if a.SelectedKey != nil && *a.SelectedKey == *q.CorrectKey {
			continue
		}
		sel := "<blank>"
		if a.SelectedKey != nil {
			sel = *a.SelectedKey
		}
		stem := ""
		if q.Stem != nil {
			stem = *q.Stem
		}
		lines = append(lines, fmt.Sprintf("question %s: chose %s, correct %s. %s", a.QuestionID, sel, *q.CorrectKey, stem))
		if len(lines) >= MaxReviewItems {
			break
		}
	}
	if len(byQuestion) == 0 {
		// Content unavailable: fall back to raw selections so the prompt
		// still carries bounded attempt detail.
		for _, a := range res.Answers {
			sel := "<blank>"
			if a.SelectedKey != nil {
				sel = *a.SelectedKey
			}
			lines = append(lines, fmt.Sprintf("question %s: chose %s", a.QuestionID, sel))
			if len(lines) >= MaxReviewItems {
				break
			}
		}
	}
	return lines
}

// ieltsReviewLines matches wrong IELTS answers against the content key
// when available; otherwise it lists raw selections.
func ieltsReviewLines(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt, res *ielts.Result) []string {
	byQuestion := map[string]ielts.Question{}
	if content := loadIeltsContent(ctx, tests, m); content != nil {
		for _, sec := range content.Sections {
			for _, qg := range sec.QuestionGroups {
				for _, q := range qg.Questions {
					byQuestion[q.ID] = q
				}
			}
		}
	}
	var lines []string
	for _, a := range res.Answers {
		q, ok := byQuestion[a.QuestionID]
		if !ok {
			continue
		}
		if ieltsAnswerCorrect(q, a) {
			continue
		}
		lines = append(lines, fmt.Sprintf("question %s: chose %s, correct %s", a.QuestionID, ieltsSelected(a), ieltsCorrect(q)))
		if len(lines) >= MaxReviewItems {
			break
		}
	}
	if len(byQuestion) == 0 {
		for _, a := range res.Answers {
			lines = append(lines, fmt.Sprintf("question %s: chose %s", a.QuestionID, ieltsSelected(a)))
			if len(lines) >= MaxReviewItems {
				break
			}
		}
	}
	return lines
}

// planSummaryLines aggregates weak areas at section level for the plan
// prompt. Failures degrade to an empty summary (generic plan) instead of
// an error.
func planSummaryLines(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) []string {
	if m == nil {
		return nil
	}
	if res, err := toeic.ParseResult(m.Result); err == nil && res != nil {
		return toeicPlanLines(ctx, tests, m, res)
	}
	if res, err := ielts.ParseResult(m.Result); err == nil && res != nil {
		return ieltsPlanLines(res)
	}
	return nil
}

func toeicPlanLines(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt, res *toeic.Result) []string {
	content := loadToeicContent(ctx, tests, m)
	if content == nil {
		if len(res.Answers) == 0 {
			return nil
		}
		return []string{fmt.Sprintf("%d answers recorded", len(res.Answers))}
	}
	partOf := map[string]int{}
	byQuestion := map[string]string{}
	for _, p := range content.Parts {
		for _, item := range p.Items {
			for _, q := range item.Questions {
				partOf[q.ID] = p.PartNumber
				if q.CorrectKey != nil {
					byQuestion[q.ID] = *q.CorrectKey
				}
			}
		}
	}
	if len(partOf) == 0 {
		if len(res.Answers) == 0 {
			return nil
		}
		return []string{fmt.Sprintf("%d answers recorded", len(res.Answers))}
	}
	wrongByPart := map[int]int{}
	totalByPart := map[int]int{}
	correctByPart := map[int]int{}
	for _, a := range res.Answers {
		part, ok := partOf[a.QuestionID]
		if !ok {
			continue
		}
		totalByPart[part]++
		correct := false
		if key, ok := byQuestion[a.QuestionID]; ok && a.SelectedKey != nil && *a.SelectedKey == key {
			correct = true
		}
		if correct {
			correctByPart[part]++
		} else {
			wrongByPart[part]++
		}
	}
	var lines []string
	for part, wrong := range wrongByPart {
		lines = append(lines, fmt.Sprintf("TOEIC part %d: %d wrong of %d attempted", part, wrong, totalByPart[part]))
	}
	return lines
}

func ieltsPlanLines(res *ielts.Result) []string {
	if len(res.SectionScores) == 0 {
		if len(res.Answers) == 0 {
			return nil
		}
		return []string{fmt.Sprintf("%d answers recorded", len(res.Answers))}
	}
	var lines []string
	for _, sec := range res.SectionScores {
		lines = append(lines, fmt.Sprintf("section %d: %d correct of %d", sec.SectionNumber, sec.RawCorrect, sec.RawTotal))
	}
	return lines
}

// loadToeicContent returns the decoded TOEIC content for the attempt's
// test, or nil when the test id is absent, the load fails, or the content
// is not TOEIC. Never returns an error: prompt building must not fail on
// content problems.
func loadToeicContent(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) *toeic.Content {
	raw := loadTestContent(ctx, tests, m)
	if len(raw) == 0 {
		return nil
	}
	c, err := toeic.ParseContent(raw)
	if err != nil {
		return nil
	}
	return c
}

// loadIeltsContent mirrors loadToeicContent for IELTS tests.
func loadIeltsContent(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) *ielts.Content {
	raw := loadTestContent(ctx, tests, m)
	if len(raw) == 0 {
		return nil
	}
	c, err := ielts.ParseContent(raw)
	if err != nil {
		return nil
	}
	return c
}

// loadTestContent fetches the attempt's ExamTest content bytes. A missing
// ExamTestID, a load error, or an empty catalog degrades to nil.
func loadTestContent(ctx context.Context, tests repositories.ExamTestRepository, m *models.Attempt) []byte {
	if tests == nil || m == nil || m.ExamTestID == nil || *m.ExamTestID == "" {
		return nil
	}
	t, err := tests.GetByID(ctx, *m.ExamTestID)
	if err != nil || t == nil {
		return nil
	}
	return []byte(t.Content)
}

func ieltsAnswerCorrect(q ielts.Question, a ielts.Answer) bool {
	if q.CorrectKey != nil && a.SelectedKey != nil && *a.SelectedKey == *q.CorrectKey {
		return true
	}
	return false
}

func ieltsSelected(a ielts.Answer) string {
	if a.SelectedKey != nil {
		return *a.SelectedKey
	}
	if a.TextValue != nil {
		return *a.TextValue
	}
	return "<blank>"
}

func ieltsCorrect(q ielts.Question) string {
	if q.CorrectKey != nil {
		return *q.CorrectKey
	}
	if len(q.CorrectAnswers) > 0 {
		return q.CorrectAnswers[0]
	}
	return "?"
}
