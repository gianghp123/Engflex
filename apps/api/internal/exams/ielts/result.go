package ielts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"engflex-api/internal/common/enums"
)

var whitespaceRegex = regexp.MustCompile(`\s+`)

type Answer struct {
	QuestionID   string   `json:"questionId"`
	SelectedKey  *string  `json:"selectedKey,omitempty"`
	SelectedKeys []string `json:"selectedKeys,omitempty"`
	TextValue    *string  `json:"textValue,omitempty"`
	Flagged      bool     `json:"flagged"`
}

type SectionScore struct {
	SectionNumber int `json:"sectionNumber"`
	RawCorrect    int `json:"rawCorrect"`
	RawTotal      int `json:"rawTotal"`
}

type Result struct {
	Revision      int                  `json:"revision"`
	Answers       []Answer             `json:"answers"`
	RawCorrect    int                  `json:"rawCorrect"`
	RawTotal      int                  `json:"rawTotal"`
	AccuracyPct   float64              `json:"accuracyPct"`
	BandScore     float64              `json:"bandScore"`
	CorrectCount  *int                 `json:"correctCount,omitempty"`
	QuestionCount int                  `json:"questionCount"`
	SectionScores map[int]SectionScore `json:"sectionScores,omitempty"`
}

// ParseResult decodes an IELTS attempt result JSON document.
func ParseResult(raw []byte) (*Result, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty IELTS result")
	}

	var res Result
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&res); err != nil {
		return nil, fmt.Errorf("decode IELTS result: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("invalid trailing result")
	}
	if res.Revision < 1 {
		return nil, errors.New("invalid IELTS result revision")
	}
	if res.Answers == nil {
		res.Answers = []Answer{}
	}

	return &res, nil
}

// NormalizeText standardizes text answers for reliable, case-insensitive, whitespace-trimmed comparison.
func NormalizeText(text string) string {
	s := strings.TrimSpace(text)
	s = strings.ToLower(s)
	s = whitespaceRegex.ReplaceAllString(s, " ")
	s = strings.TrimRight(s, ".,;!?:")
	return strings.TrimSpace(s)
}

// GradeIeltsAttempt evaluates candidate answers against the content answer key and calculates official band score.
func GradeIeltsAttempt(content *Content, answers []Answer) (*Result, error) {
	if content == nil {
		return nil, errors.New("content is nil")
	}

	answerMap := make(map[string]Answer, len(answers))
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	rawTotal := 0
	rawCorrect := 0
	sectionScores := make(map[int]SectionScore)

	for _, sec := range content.Sections {
		secScore := sectionScores[sec.SectionNumber]
		secScore.SectionNumber = sec.SectionNumber

		for _, qg := range sec.QuestionGroups {
			for _, q := range qg.Questions {
				rawTotal++
				secScore.RawTotal++

				ans, answered := answerMap[q.ID]
				if !answered {
					continue
				}

				isCorrect := false

				if q.CorrectKey != nil && ans.SelectedKey != nil {
					if strings.EqualFold(strings.TrimSpace(*ans.SelectedKey), strings.TrimSpace(*q.CorrectKey)) {
						isCorrect = true
					}
				}

				if !isCorrect && len(q.CorrectKeys) > 0 && len(ans.SelectedKeys) > 0 {
					if checkMultiKeys(ans.SelectedKeys, q.CorrectKeys) {
						isCorrect = true
					}
				}

				if !isCorrect && len(q.CorrectAnswers) > 0 && ans.TextValue != nil {
					candidate := NormalizeText(*ans.TextValue)
					if candidate != "" {
						for _, alt := range q.CorrectAnswers {
							if candidate == NormalizeText(alt) {
								isCorrect = true
								break
							}
						}
					}
				}

				if isCorrect {
					rawCorrect++
					secScore.RawCorrect++
				}
			}
		}

		sectionScores[sec.SectionNumber] = secScore
	}

	var bandScore float64
	switch content.Skill {
	case enums.ExamSkillListening:
		bandScore = CalculateListeningBand(rawCorrect)
	case enums.ExamSkillReading:
		bandScore = CalculateReadingBand(rawCorrect, content.SubType)
	default:
		bandScore = CalculateReadingBand(rawCorrect, content.SubType)
	}

	accuracyPct := 0.0
	if rawTotal > 0 {
		accuracyPct = (float64(rawCorrect) / float64(rawTotal)) * 100.0
	}

	correctCount := rawCorrect
	return &Result{
		Revision:      1,
		Answers:       answers,
		RawCorrect:    rawCorrect,
		RawTotal:      rawTotal,
		AccuracyPct:   accuracyPct,
		BandScore:     bandScore,
		CorrectCount:  &correctCount,
		QuestionCount: rawTotal,
		SectionScores: sectionScores,
	}, nil
}

func checkMultiKeys(provided, expected []string) bool {
	if len(provided) != len(expected) {
		return false
	}

	normProvided := make([]string, len(provided))
	for i, p := range provided {
		normProvided[i] = strings.ToUpper(strings.TrimSpace(p))
	}
	normExpected := make([]string, len(expected))
	for i, e := range expected {
		normExpected[i] = strings.ToUpper(strings.TrimSpace(e))
	}

	sort.Strings(normProvided)
	sort.Strings(normExpected)

	for i := range normProvided {
		if normProvided[i] != normExpected[i] {
			return false
		}
	}
	return true
}
