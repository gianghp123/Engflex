// Package prompts builds the post-session summary LLM prompt. Pure functions
// only: no config, no network, no database.
package prompts

import (
	"fmt"
	"strings"

	"engflex-api/internal/database/models"
)

// BuildSessionSummaryPrompt renders the objective-check brief: objectives as
// numbered tasks plus the learner's own turns, truncated server-side. The
// model judges each objective against the transcript only — no pronunciation
// claims, no mid-conversation grading.
func BuildSessionSummaryPrompt(objs []models.ScenarioObjective, transcript string) string {
	if len(objs) == 0 {
		return "You score a spoken English practice session from its transcript.\n" +
			"Return JSON {takeaways, objectives, levelNote}.\n" +
			"takeaways: 2-3 sentences on what the learner did well and one fix.\n" +
			"objectives: [] (free talk has no objectives).\n" +
			"levelNote: 1 sentence on the learner's level.\n" +
			"Never comment on pronunciation.\n\nTranscript:\n" + transcript
	}
	var lines []string
	for _, o := range objs {
		lines = append(lines, fmt.Sprintf("%d. %s (%s)", o.ID, o.Label, o.Hint))
	}
	return "You score a spoken English roleplay session from its transcript.\n" +
		"The learner had " + fmt.Sprintf("%d", len(objs)) + " tasks:\n" + strings.Join(lines, "\n") + "\n\n" +
		"Return JSON {takeaways, objectives, levelNote}.\n" +
		"takeaways: 2-3 sentences, one strength + one fix.\n" +
		"objectives: one per task {id, status done|partial|missed, evidence: exact learner quote or ''}.\n" +
		"Be conservative: missed unless the transcript shows it.\n" +
		"levelNote: 1 sentence on the learner's level.\n" +
		"Never comment on pronunciation.\n\nTranscript:\n" + transcript
}
