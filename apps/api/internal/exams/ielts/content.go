package ielts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"engflex-api/internal/common/enums"
)

type Content struct {
	SchemaVersion string            `json:"schemaVersion"`
	TestID        string            `json:"testId"`
	Title         string            `json:"title"`
	Skill         enums.ExamSkill   `json:"skill"`
	SubType       enums.ExamSubType `json:"subType"`
	Sections      []Section         `json:"sections"`
	Assets        []Asset           `json:"assets,omitempty"`
}

type Asset struct {
	SourceID string  `json:"sourceId"`
	Type     string  `json:"type"`
	SHA256   *string `json:"sha256,omitempty"`
	URL      *string `json:"url,omitempty"`
}

type Section struct {
	SectionNumber  int             `json:"sectionNumber"`
	Title          string          `json:"title"`
	Passages       []Passage       `json:"passages"`
	QuestionGroups []QuestionGroup `json:"questionGroups"`
}

type Passage struct {
	ID      string      `json:"id"`
	Title   *string     `json:"title,omitempty"`
	Label   *string     `json:"label,omitempty"`
	Content []Paragraph `json:"content"`
}

type Paragraph struct {
	ID    string  `json:"id,omitempty"`
	Label *string `json:"label,omitempty"`
	Text  string  `json:"text"`
}

type QuestionGroup struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Title        *string    `json:"title,omitempty"`
	Instructions *string    `json:"instructions,omitempty"`
	Questions    []Question `json:"questions"`
}

type Question struct {
	ID                    string   `json:"id"`
	QuestionNumber        int      `json:"questionNumber"`
	Stem                  *string  `json:"stem,omitempty"`
	Options               []Option `json:"options,omitempty"`
	CorrectKey            *string  `json:"correctKey,omitempty"`
	CorrectKeys           []string `json:"correctKeys,omitempty"`
	CorrectAnswers        []string `json:"correctAnswers,omitempty"`
	Explanation           *string  `json:"explanation,omitempty"`
	VietnameseExplanation *string  `json:"vietnameseExplanation,omitempty"`
}

type Option struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// ParseContent parses and validates an IELTS test content document.
func ParseContent(raw []byte) (*Content, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty IELTS content")
	}

	var content Content
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&content); err != nil {
		return nil, fmt.Errorf("decode IELTS content: %w", err)
	}

	if content.SchemaVersion == "" {
		return nil, errors.New("missing schemaVersion")
	}
	if len(content.Sections) == 0 {
		return nil, errors.New("IELTS content must have at least one section")
	}

	seenQuestions := make(map[string]bool)
	seenNumbers := make(map[int]bool)

	for _, sec := range content.Sections {
		if sec.SectionNumber < 1 {
			return nil, fmt.Errorf("invalid section number: %d", sec.SectionNumber)
		}
		for _, qg := range sec.QuestionGroups {
			for _, q := range qg.Questions {
				if q.ID == "" {
					return nil, errors.New("question must have non-empty id")
				}
				if seenQuestions[q.ID] {
					return nil, fmt.Errorf("duplicate question id: %s", q.ID)
				}
				seenQuestions[q.ID] = true

				if q.QuestionNumber < 1 || q.QuestionNumber > 100 {
					return nil, fmt.Errorf("invalid question number: %d", q.QuestionNumber)
				}
				if seenNumbers[q.QuestionNumber] {
					return nil, fmt.Errorf("duplicate question number: %d", q.QuestionNumber)
				}
				seenNumbers[q.QuestionNumber] = true

				hasCorrect := q.CorrectKey != nil || len(q.CorrectKeys) > 0 || len(q.CorrectAnswers) > 0
				if !hasCorrect {
					return nil, fmt.Errorf("question %s (%d) has no correct answer defined", q.ID, q.QuestionNumber)
				}
			}
		}
	}

	return &content, nil
}
