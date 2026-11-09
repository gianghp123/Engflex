package models

import "time"

// SavedItem is a read-only projection of bookmarks and saved vocabulary.
type SavedItem struct {
	ContentType    string
	ContentID      string
	Note           *string
	CreatedAt      time.Time
	Lesson         *Lesson         `gorm:"->;foreignKey:ContentID;references:ID"`
	VideoExercise  *VideoExercise  `gorm:"->;foreignKey:ContentID;references:ID"`
	Story          *Story          `gorm:"->;foreignKey:ContentID;references:ID"`
	VocabularyItem *VocabularyItem `gorm:"->;foreignKey:ContentID;references:ID"`
}

func (SavedItem) TableName() string { return "saved_items" }

type SavedLibrary struct {
	Items     []*SavedItem `gorm:"-"`
	Total     int64
	WithNotes int64
	Vocab     int64
	Lesson    int64
	Video     int64
	Story     int64
}
