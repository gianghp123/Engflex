package services

import (
	"encoding/json"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/lessons/activityconfig"
	"strconv"
)

// Scores use all expected questions/items, so one answer cannot earn full credit.
func normalizePart(a *models.LessonActivity, p *PartEntry) error {
	if a.Type != "" && string(p.Type) != string(a.Type) {
		return common.BadRequest("result type does not match activity")
	}
	raw, err := json.Marshal(p.Detail)
	if err != nil {
		return common.Internal()
	}
	switch a.Type {
	case enums.ActivityTypeReading, enums.ActivityTypeListening:
		questions, err := activityconfig.QuestionsFor(a.Type, a.Config)
		if err != nil || len(questions) == 0 {
			return common.Internal()
		}
		var detail CheckPartDetail
		if json.Unmarshal(raw, &detail) != nil {
			return common.BadRequest("invalid question outcomes")
		}
		seen := map[int]bool{}
		correct := 0
		for _, o := range detail.Outcomes {
			if o.Index < 0 || o.Index >= len(questions) || seen[o.Index] {
				return common.BadRequest("invalid question index")
			}
			seen[o.Index] = true
			if o.Correct {
				correct++
			}
		}
		p.Score = 100 * float64(correct) / float64(len(questions))
	case enums.ActivityTypeSpeaking:
		cfg, err := activityconfig.ParseSpeakingConfig(a.Config)
		if err != nil || len(cfg.Items) == 0 {
			return common.Internal()
		}
		var detail PronouncePartDetail
		if json.Unmarshal(raw, &detail) != nil {
			return common.BadRequest("invalid pronunciation outcomes")
		}
		sum := 0.0
		for key, score := range detail.Items {
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(cfg.Items) || score < 0 || score > 100 {
				return common.BadRequest("invalid speaking item")
			}
			sum += score
		}
		p.Score = sum / float64(len(cfg.Items))
	}
	return nil
}
