package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

type Bookmark struct {
	ContentType enums.ContentType `json:"contentType"`
	ContentID   string            `json:"contentId"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type SavedLibrary struct {
	Items     []SavedItem `json:"items"`
	Total     int64       `json:"total"`
	WithNotes int64       `json:"withNotes"`
	Vocab     int64       `json:"vocab"`
	Lesson    int64       `json:"lesson"`
	Video     int64       `json:"video"`
	Story     int64       `json:"story"`
}

type SavedItem struct {
	ContentType    string               `json:"contentType"`
	ContentID      string               `json:"contentId"`
	Note           *string              `json:"note,omitempty"`
	CreatedAt      time.Time            `json:"createdAt"`
	Lesson         *SavedLesson         `json:"lesson,omitempty"`
	VideoExercise  *SavedVideo          `json:"videoExercise,omitempty"`
	Story          *SavedStory          `json:"story,omitempty"`
	VocabularyItem *SavedVocabularyItem `json:"vocabularyItem,omitempty"`
}

type SavedLesson struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type SavedVideo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SavedStory struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type SavedVocabularyItem struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}
