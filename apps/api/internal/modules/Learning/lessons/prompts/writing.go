// Package prompts builds writing-score LLM prompts. Pure functions only:
// no config, no network, no database.
package prompts

import (
	"fmt"
	"strings"

	"engflex-api/internal/modules/lessons/activityconfig"
)

// BuildWritingScorePrompt renders the six-section review brief. The model
// answer is a reference for what good looks like — never the correction
// target, never copied into suggested. The learner text is the only thing
// judged. Static few-shot examples use the Kenji brief (tested live); template
// variables stay strictly for dynamic parts.
func BuildWritingScorePrompt(cfg activityconfig.WritingConfig, text string, wordCount int) string {
	checklist := make([]string, 0, len(cfg.Checklist))
	for i, item := range cfg.Checklist {
		checklist = append(checklist, fmt.Sprintf("%d. %s", i+1, item))
	}
	return "You are an English coach reviewing ONE learner writing response. " +
		"Return feedback in exactly six sections.\n\n" +
		"Task: " + cfg.Task + "\n" +
		"Brief: " + cfg.Instructions + "\n" +
		"Context: " + cfg.Stimulus + "\n" +
		fmt.Sprintf("Length: %d words (required %d-%d)\n", wordCount, cfg.MinWords, cfg.MaxWords) +
		"Checklist:\n" + strings.Join(checklist, "\n") + "\n\n" +
		"Reference answer (what good looks like; do NOT copy it into suggested):\n" + cfg.ModelAnswer + "\n\n" +
		"Learner response to review:\n" + text + "\n\n" +
		"Rules:\n" +
		"1. task.verdicts has exactly one entry per checklist item, in order; evidence must be an EXACT " +
		"substring of the learner response, or \"\" when missing. task.answersPrompt is whether the response " +
		"actually answers the prompt; task.answersNote says why in one short sentence.\n" +
		"2. grammar holds REAL grammar or spelling errors only; text must be an EXACT substring of the learner " +
		"response; occurrence is 1-based; correction fixes only that span; reason briefly says why. Empty array " +
		"when there are none — never invent errors.\n" +
		"3. phrasing holds expressions that are grammatical but unnatural or less appropriate; same span rules. " +
		"Never label a stylistic preference as an error and never put a grammar mistake here. Empty array when none.\n" +
		"4. expressions holds 3-5 reusable phrases directly relevant to THIS task; practical sentence patterns " +
		"with ... slots that work in other similar tasks too, not isolated vocabulary and not sentences " +
		"specific to this one email.\n" +
		"5. suggested is a complete improved version covering every checklist point within the required word " +
		"range; natural phrasing, never a copy of the reference answer.\n" +
		"6. tip is one concise actionable recommendation targeting the learner's main weakness; do not repeat " +
		"a span reason. Never mention model names.\n\n" +
		"Reply with a single JSON object and nothing else, keys exactly: " +
		"task, grammar, phrasing, expressions, suggested, tip.\n\n" +
		writingScoreExamples
}

// writingScoreExamples are static few-shot examples (sona-voice pattern): one
// scored email with mixed verdicts, one perfect response, one NOT
// anti-example. They use the real seeded Kenji brief so prompt and fixtures
// agree. Template variables stay strictly for dynamic parts.
const writingScoreExamples = `Examples:

Learner response:
Hello everyone,

My name is Kenji Sato and I is a sales assistant. I joined the company this month. I handle customer orders. Happy to meet you all.

Output:
{"task": {"verdicts": [{"item": "You greeted the team", "status": "covered", "evidence": "Hello everyone"}, {"item": "You gave your name and role", "status": "covered", "evidence": "My name is Kenji Sato"}, {"item": "You said where you are based", "status": "missing", "evidence": ""}, {"item": "You stayed within 40-60 words", "status": "missing", "evidence": ""}], "answersPrompt": true, "answersNote": "Asks to introduce but omits the base location."}, "grammar": [{"text": "I is", "occurrence": 1, "correction": "I am", "reason": "Be verb agrees with I as am"}], "phrasing": [{"text": "Happy to meet you all", "occurrence": 1, "correction": "I am happy to meet you all", "reason": "More complete and natural closing"}], "expressions": ["My name is ... and I am a ... in ...", "I joined ... this month", "I look forward to working with ..."], "suggested": "Hello everyone,\n\nMy name is Kenji Sato and I am a sales assistant in the Osaka office. I joined the company this month. I handle customer orders and shipping schedules. I am happy to meet you all and look forward to working together.\n\nBest regards,\nKenji", "tip": "Remember the be verb: I am, not I is."}

---

Learner response:
Hello everyone,

My name is Kenji Sato and I am a sales assistant in the Osaka office. I joined the company this month. I handle customer orders and shipping schedules. I am happy to meet you all and look forward to working together.

Best regards,
Kenji

Output:
{"task": {"verdicts": [{"item": "You greeted the team", "status": "covered", "evidence": "Hello everyone"}, {"item": "You gave your name and role", "status": "covered", "evidence": "My name is Kenji Sato"}, {"item": "You said where you are based", "status": "covered", "evidence": "Osaka office"}, {"item": "You stayed within 40-60 words", "status": "covered", "evidence": ""}], "answersPrompt": true, "answersNote": "Answers the prompt fully."}, "grammar": [], "phrasing": [], "expressions": ["My name is ... and I am a ... in ...", "I joined ... this month", "I look forward to working with ..."], "suggested": "Hello everyone,\n\nMy name is Kenji Sato and I am a sales assistant in the Osaka office. I joined the company this month. I handle customer orders and shipping schedules. I am happy to meet you all and look forward to working together.\n\nBest regards,\nKenji", "tip": "Clean email covering every point."}

---

NOT (suggested must improve the learner's response — never copy the reference answer; phrasing must never hold a grammar mistake).
`
