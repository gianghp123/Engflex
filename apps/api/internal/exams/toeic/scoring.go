package toeic

import "fmt"

// ScoringVersion identifies the product's estimated score table, shared by both skills.
const ScoringVersion = "toeic-estimate-v1"

// EstimateScore converts a raw score for a complete 100-question skill.
func EstimateScore(correct int) (float64, error) {
	if correct < 0 || correct > 100 {
		return 0, fmt.Errorf("invalid TOEIC raw score: %d", correct)
	}
	return float64(max(5, 5*(correct-1))), nil
}

func validateScoring(result *Result) error {
	if result.ScoringVersion == nil && result.Sections == nil {
		return nil // Pre-scoring snapshots remain readable without being regraded.
	}
	invalid := func() error { return fmt.Errorf("invalid TOEIC scoring snapshot") }
	if result.ScoringVersion == nil || *result.ScoringVersion != ScoringVersion || result.Sections == nil || result.ScaledScore != nil {
		return invalid()
	}
	count, correct := 0, 0
	for _, section := range []*SectionResult{result.Sections.Listening, result.Sections.Reading} {
		if section == nil {
			continue
		}
		if section.QuestionCount < 1 || section.QuestionCount > 100 || section.CorrectCount < 0 || section.CorrectCount > section.QuestionCount {
			return invalid()
		}
		if section.QuestionCount == 100 {
			score, _ := EstimateScore(section.CorrectCount)
			if section.ScaledScore == nil || *section.ScaledScore != score {
				return invalid()
			}
		} else if section.ScaledScore != nil {
			return invalid()
		}
		count += section.QuestionCount
		correct += section.CorrectCount
	}
	if count != len(result.Answers) || count == 0 || result.CorrectCount == nil || *result.CorrectCount != correct || result.AccuracyPct == nil || *result.AccuracyPct != 100*float64(correct)/float64(count) {
		return invalid()
	}
	return nil
}
