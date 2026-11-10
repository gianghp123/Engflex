package responses

import "engflex-api/internal/common/enums"

type IeltsContentMetadata struct {
	Sections []IeltsSectionMetadata `json:"sections"`
}

type IeltsSectionMetadata struct {
	SectionNumber int    `json:"sectionNumber"`
	Title         string `json:"title"`
	QuestionCount int    `json:"questionCount"`
}

type IeltsTakingContent struct {
	SchemaVersion string               `json:"schemaVersion"`
	TestID        string               `json:"testId"`
	Title         string               `json:"title"`
	Skill         enums.ExamSkill      `json:"skill"`
	SubType       enums.ExamSubType    `json:"subType"`
	Sections      []IeltsTakingSection `json:"sections"`
	Assets        []Asset              `json:"assets,omitempty"`
}

type IeltsTakingSection struct {
	SectionNumber  int                        `json:"sectionNumber"`
	Title          string                     `json:"title"`
	Passages       []IeltsPassage             `json:"passages"`
	QuestionGroups []IeltsTakingQuestionGroup `json:"questionGroups"`
}

type IeltsPassage struct {
	ID      string           `json:"id"`
	Title   *string          `json:"title,omitempty"`
	Label   *string          `json:"label,omitempty"`
	Content []IeltsParagraph `json:"content"`
}

type IeltsParagraph struct {
	ID    string  `json:"id,omitempty"`
	Label *string `json:"label,omitempty"`
	Text  string  `json:"text"`
}

type IeltsTakingQuestionGroup struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Title        *string               `json:"title,omitempty"`
	Instructions *string               `json:"instructions,omitempty"`
	Questions    []IeltsTakingQuestion `json:"questions"`
}

type IeltsTakingQuestion struct {
	ID             string   `json:"id"`
	QuestionNumber int      `json:"questionNumber"`
	Stem           *string  `json:"stem,omitempty"`
	Options        []Option `json:"options,omitempty"`
}

type IeltsReviewContent struct {
	SchemaVersion string               `json:"schemaVersion"`
	TestID        string               `json:"testId"`
	Title         string               `json:"title"`
	Skill         enums.ExamSkill      `json:"skill"`
	SubType       enums.ExamSubType    `json:"subType"`
	Sections      []IeltsReviewSection `json:"sections"`
	Assets        []Asset              `json:"assets,omitempty"`
}

type IeltsReviewSection struct {
	SectionNumber  int                        `json:"sectionNumber"`
	Title          string                     `json:"title"`
	Passages       []IeltsPassage             `json:"passages"`
	QuestionGroups []IeltsReviewQuestionGroup `json:"questionGroups"`
}

type IeltsReviewQuestionGroup struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Title        *string               `json:"title,omitempty"`
	Instructions *string               `json:"instructions,omitempty"`
	Questions    []IeltsReviewQuestion `json:"questions"`
}

type IeltsReviewQuestion struct {
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
