package models

// Span is one quoted problem with a minimal fix. The array it sits in is the
// label (grammar = real error, phrasing = awkward only) — there is no status
// field to drift.
type WritingSpan struct {
	Text       string `json:"text" jsonschema_description:"Exact problematic substring"`
	Occurrence int    `json:"occurrence" jsonschema_description:"1-based occurrence"`
	Correction string `json:"correction" jsonschema_description:"Fix for this span only"`
	Reason     string `json:"reason" jsonschema_description:"Brief why"`
}

type WritingTaskVerdict struct {
	Item     string `json:"item" jsonschema_description:"Checklist item judged"`
	Status   string `json:"status" jsonschema:"enum=covered,enum=partial,enum=missing" jsonschema_description:"Coverage verdict"`
	Evidence string `json:"evidence" jsonschema_description:"Exact learner-text quote, or empty when missing"`
}

// TaskCompletion is section 1: did the response do the task? Counts live in
// the service-built summary, never here (LLMs cannot count).
type WritingTaskCompletion struct {
	Verdicts      []WritingTaskVerdict `json:"verdicts" jsonschema_description:"One verdict per checklist item, same order"`
	AnswersPrompt bool                 `json:"answersPrompt" jsonschema_description:"Whether the response answers the prompt"`
	AnswersNote   string               `json:"answersNote" jsonschema_description:"Short why or what is missing"`
}

// WritingFeedback is the full LLM output: six sections, no counts of its
// own, no copy of the learner paragraph.
type WritingFeedback struct {
	Task        WritingTaskCompletion `json:"task" jsonschema_description:"Task completion verdicts"`
	Grammar     []WritingSpan         `json:"grammar" jsonschema_description:"Real grammar or spelling errors only"`
	Phrasing    []WritingSpan         `json:"phrasing" jsonschema_description:"Grammatical but unnatural phrasing only"`
	Expressions []string              `json:"expressions" jsonschema_description:"3-5 reusable task-relevant patterns"`
	Suggested   string                `json:"suggested" jsonschema_description:"Complete improved answer in word range"`
	Tip         string                `json:"tip" jsonschema_description:"One actionable recommendation"`
}
