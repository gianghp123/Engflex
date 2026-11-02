package requests

import "engflex-api/internal/common/enums"

type StartVideoAttempt struct {
	Type enums.AttemptType `json:"type" form:"type" binding:"required,oneof=dictation shadowing"`
}

type SaveVideoProgress struct {
	Revision               int            `json:"revision" binding:"required,min=1"`
	ActiveSequence         int            `json:"activeSequence" binding:"min=0"`
	ActiveTranscriptID     *string        `json:"activeTranscriptId" binding:"omitempty,uuid"`
	Level                  *string        `json:"level" binding:"omitempty,oneof=easy normal hard"`
	CompletedTranscriptIDs []string       `json:"completedTranscriptIds"`
	Transcripts            map[string]any `json:"transcripts"`
}

type CompleteVideoAttempt struct {
	Score       *float64 `json:"score" binding:"omitempty,min=0,max=100"`
	DurationSec *int     `json:"durationSec" binding:"omitempty,min=0"`
}
