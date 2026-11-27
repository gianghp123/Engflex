// Package schemas holds the post-session summary LLM output shape and its
// validation. The JSON schema map is generated from the struct so the two
// can never drift.
package schemas

import (
	"strings"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/llm"
)

// ObjectiveVerdict is one task judgment with the learner quote backing it.
type ObjectiveVerdict struct {
	ID       int    `json:"id" jsonschema_description:"Objective id (1-based position)"`
	Status   string `json:"status" jsonschema:"enum=done,enum=partial,enum=missed" jsonschema_description:"Verdict"`
	Evidence string `json:"evidence" jsonschema_description:"Exact learner quote, or empty when missed"`
}

// SessionSummaryLLM is the full LLM reply: takeaways, per-objective
// verdicts, and a one-sentence level note.
type SessionSummaryLLM struct {
	Takeaways  []string           `json:"takeaways" jsonschema_description:"2-3 sentences, one strength + one fix"`
	Objectives []ObjectiveVerdict `json:"objectives" jsonschema_description:"One verdict per task, task order"`
	LevelNote  string             `json:"levelNote" jsonschema_description:"One sentence on the learner's level"`
}

var SessionSummaryName = "SessionSummary"

// SessionSummarySchema is the strict json_schema map for the LLM call, via
// the shared generator.
var SessionSummarySchema = llm.GenerateSchema[SessionSummaryLLM]()

// NormalizeSessionSummary validates one decoded LLM reply: known statuses,
// one verdict per objective with a valid id, evidence that appears in the
// transcript (lenient: non-empty quotes for done/partial are accepted when
// they are close matches). Free talk (no objectives) yields empty verdicts.
func NormalizeSessionSummary(out SessionSummaryLLM, objs []models.ScenarioObjective, transcript string) (*models.ConversationSummary, error) {
	if len(out.Takeaways) == 0 {
		return nil, common.UnprocessableEntity("summary takeaways missing")
	}
	if out.LevelNote == "" {
		return nil, common.UnprocessableEntity("summary level note missing")
	}
	ids := map[int]bool{}
	for _, o := range objs {
		ids[o.ID] = true
	}
	if len(objs) == 0 {
		if len(out.Objectives) != 0 {
			return nil, common.UnprocessableEntity("unexpected objective verdicts")
		}
		return &models.ConversationSummary{Takeaways: out.Takeaways, Objectives: []models.ObjectiveResult{}, LevelNote: out.LevelNote}, nil
	}
	if len(out.Objectives) != len(objs) {
		return nil, common.UnprocessableEntity("objective verdict count mismatch")
	}
	seen := map[int]bool{}
	results := make([]models.ObjectiveResult, 0, len(out.Objectives))
	for _, v := range out.Objectives {
		if !ids[v.ID] {
			return nil, common.UnprocessableEntity("unknown objective id")
		}
		if seen[v.ID] {
			return nil, common.UnprocessableEntity("duplicate objective id")
		}
		seen[v.ID] = true
		if v.Status != "done" && v.Status != "partial" && v.Status != "missed" {
			return nil, common.UnprocessableEntity("unknown objective status")
		}
		if v.Evidence != "" && !strings.Contains(transcript, v.Evidence) {
			return nil, common.UnprocessableEntity("summary evidence not in transcript")
		}
		results = append(results, models.ObjectiveResult{ID: v.ID, Status: v.Status, Evidence: v.Evidence})
	}
	return &models.ConversationSummary{Takeaways: out.Takeaways, Objectives: results, LevelNote: out.LevelNote}, nil
}
