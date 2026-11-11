package services

import (
	"encoding/json"
	"math"
	"slices"
	"strings"
	"unicode"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

func dictationWords(text string) []string {
	lower := cases.Lower(language.Und)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\uFEFF' || (unicode.IsSpace(r) && r != '\u0085')
	})
	out := make([]string, 0, len(words))
	for _, word := range words {
		normalized := strings.Map(func(r rune) rune {
			if unicode.IsPunct(r) {
				return -1
			}
			return r
		}, lower.String(norm.NFC.String(word)))
		if normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

func dictationPositions(n int, level string) []int {
	divisor := 1
	switch level {
	case "easy":
		divisor = 3
	case "normal":
		divisor = 2
	}
	count := (n + divisor - 1) / divisor
	positions := make([]int, count)
	for i := range positions {
		positions[i] = i * n / count
	}
	return positions
}

func decodeDictationLine[T any](value any, target *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func gradeDictationLine(content, level string, input requests.DictationLineInput, previous responses.DictationLineResult) (responses.DictationLineResult, error) {
	words := dictationWords(content)
	positions := dictationPositions(len(words), level)
	result := responses.DictationLineResult{
		Draft: input.Draft, Answers: input.Answers, Revealed: slices.Clone(previous.Revealed),
	}
	if len(input.Answers) > len(positions) || (level == "hard" && len(input.Answers) != 0) {
		return result, common.BadRequest("invalid dictation answers")
	}
	for _, position := range input.Revealed {
		if !slices.Contains(positions, position) {
			return result, common.BadRequest("invalid dictation hint position")
		}
		if !slices.Contains(result.Revealed, position) {
			result.Revealed = append(result.Revealed, position)
		}
	}
	if !input.Completed {
		return result, nil
	}
	typed := dictationWords(input.Draft)
	if len(positions) == 0 || (!input.RevealAll && level == "hard" && len(typed) != len(words)) {
		return result, common.BadRequest("invalid dictation submission")
	}
	earned := 0
	for i, position := range positions {
		var answer string
		if level == "hard" && position < len(typed) {
			answer = typed[position]
		} else if level != "hard" && i < len(input.Answers) {
			if parts := dictationWords(input.Answers[i]); len(parts) == 1 {
				answer = parts[0]
			}
		}
		hinted := slices.Contains(result.Revealed, position)
		if answer != words[position] {
			if !input.RevealAll {
				return result, common.BadRequest("dictation answer is incomplete")
			}
			if !hinted {
				result.Revealed = append(result.Revealed, position)
			}
		} else if !hinted {
			earned++
		}
	}
	total := len(positions)
	result.Completed, result.Earned, result.Total = true, &earned, &total
	return result, nil
}

// mergeDictationProgress runs while the attempt row is locked.
func mergeDictationProgress(req *requests.SaveVideoProgress, previous requests.SaveVideoProgress, lines []*models.VideoTranscript) error {
	level := "hard"
	if previous.DictationVersion == 1 && previous.Level != nil {
		level = *previous.Level
	}
	if req.Level != nil && *req.Level != level {
		return common.Conflict("dictation level is locked")
	}
	incoming := req.Transcripts
	req.Level, req.DictationVersion = &level, 1
	req.Transcripts = make(map[string]any)
	req.CompletedTranscriptIDs = []string{}
	for _, line := range lines {
		old, hadState := previous.Transcripts[line.ID]
		if slices.Contains(previous.CompletedTranscriptIDs, line.ID) {
			req.CompletedTranscriptIDs = append(req.CompletedTranscriptIDs, line.ID)
			if hadState {
				req.Transcripts[line.ID] = old
			}
			continue
		}
		raw, changed := incoming[line.ID]
		if !changed {
			if hadState {
				req.Transcripts[line.ID] = old
			}
			continue
		}
		var input requests.DictationLineInput
		if err := decodeDictationLine(raw, &input); err != nil || raw == nil {
			return common.BadRequest("invalid dictation checkpoint")
		}
		var prior responses.DictationLineResult
		_ = decodeDictationLine(old, &prior)
		if previous.DictationVersion == 0 {
			positions := dictationPositions(len(dictationWords(line.Content)), level)
			prior.Revealed = slices.DeleteFunc(prior.Revealed, func(i int) bool {
				return !slices.Contains(positions, i)
			})
		}
		result, err := gradeDictationLine(line.Content, level, input, prior)
		if err != nil {
			return err
		}
		req.Transcripts[line.ID] = result
		if result.Completed {
			req.CompletedTranscriptIDs = append(req.CompletedTranscriptIDs, line.ID)
		}
	}
	return nil
}

func dictationScore(progress requests.SaveVideoProgress, lines []*models.VideoTranscript) (*float64, error) {
	earned, total := 0, 0
	hasUnscored := progress.DictationVersion == 0
	for _, line := range lines {
		if progress.DictationVersion == 1 && len(dictationWords(line.Content)) == 0 {
			continue
		}
		if !slices.Contains(progress.CompletedTranscriptIDs, line.ID) {
			return nil, common.Conflict("dictation has unfinished transcripts")
		}
		var result responses.DictationLineResult
		if err := decodeDictationLine(progress.Transcripts[line.ID], &result); err != nil || result.Total == nil || result.Earned == nil {
			hasUnscored = true
			continue
		}
		earned += *result.Earned
		total += *result.Total
	}
	if hasUnscored || total == 0 {
		return nil, nil
	}
	score := math.Round(10000*float64(earned)/float64(total)) / 100
	return &score, nil
}
