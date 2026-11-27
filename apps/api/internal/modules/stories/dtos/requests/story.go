package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

type ListStories struct {
	common.ListParams `tstype:",extends"`
	CEFR              enums.CEFR `form:"cefr" json:"cefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Topic             string     `form:"topic" json:"topic" binding:"omitempty,max=100"`
}

// SentenceInput is one line of a story's reading timeline. No ID — every
// save replaces the whole set (see StoryRepository.ReplaceSentences), so
// the admin form just submits the full current set as one array.
type SentenceInput struct {
	Sequence       int     `json:"sequence"`
	EnglishText    string  `json:"englishText"`
	VietnameseText string  `json:"vietnameseText"`
	ImageURL       *string `json:"imageUrl"`
	StartTimestamp float64 `json:"startTimestamp"`
	EndTimestamp   float64 `json:"endTimestamp"`
}

type CreateStory struct {
	Slug            string          `json:"slug"`
	Title           string          `json:"title"`
	TitleVietnamese *string         `json:"titleVietnamese"`
	CEFRLevel       *enums.CEFR     `json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Topic           string          `json:"topic"`
	ThumbnailURL    *string         `json:"thumbnailUrl"`
	AudioURL        string          `json:"audioUrl"`
	Duration        float64         `json:"duration"`
	IsPremium       bool            `json:"isPremium"`
	Sentences       []SentenceInput `json:"sentences"`
}

// UpdateStory replaces every field below Status. Status is the one pointer
// field — optional, so a plain metadata edit never has to resend it
// (mirrors UpdateVideoExercise). Sentences is nil-vs-empty sensitive: a
// request that omits the key leaves the existing timeline untouched; `[]`
// clears it. See storyService.Update.
type UpdateStory struct {
	Slug            string               `json:"slug"`
	Title           string               `json:"title"`
	TitleVietnamese *string              `json:"titleVietnamese"`
	CEFRLevel       *enums.CEFR          `json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Topic           string               `json:"topic"`
	ThumbnailURL    *string              `json:"thumbnailUrl"`
	AudioURL        string               `json:"audioUrl"`
	Duration        float64              `json:"duration"`
	Status          *enums.PublishStatus `json:"status" binding:"omitempty,oneof=draft published"`
	IsPremium       *bool                `json:"isPremium"`
	Sentences       []SentenceInput      `json:"sentences,omitempty"`
}
