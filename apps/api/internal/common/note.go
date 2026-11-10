package common

import "strings"

const NoteMaxLength = 2000

type UpdateSavedNote struct {
	Note *string `json:"note" binding:"required,max=2000" tstype:"string"`
}

func (r UpdateSavedNote) Normalized() *string {
	if r.Note == nil {
		return nil
	}
	note := strings.TrimSpace(*r.Note)
	if note == "" {
		return nil
	}
	return &note
}
