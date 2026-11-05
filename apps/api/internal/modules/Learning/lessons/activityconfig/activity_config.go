// Package activityconfig holds the document schemas of lesson_activities.config.
// Parse-only: these are not GORM models, have no table and no migration —
// models/ stays GORM structs and their field types (e.g. LessonDetails).
package activityconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
)

// ConfigOption is one MCQ choice inside an activity config.
type ConfigOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// ConfigQuestion is one checkable item. AnswerKey and Explanation are
// server-side only: the check endpoint reads them, the wire DTO never copies
// them (copier drops fields the DTO lacks).
type ConfigQuestion struct {
	Stem        string         `json:"stem"`
	Instruction string         `json:"instruction"`
	Options     []ConfigOption `json:"options"`
	AnswerKey   string         `json:"answerKey"`
	Explanation string         `json:"explanation"`
}

// ReadingConfig is lesson_activities.config for type=reading.
type ReadingConfig struct {
	Text      string           `json:"text"`
	Questions []ConfigQuestion `json:"questions"`
}

// ScriptLine is one spoken turn of a listening script. Voice is a TTS voice id,
// used by scripts/seed_audio.py only — it never crosses the wire.
type ScriptLine struct {
	Speaker string `json:"speaker,omitempty"`
	Voice   string `json:"voice"`
	Text    string `json:"text"`
}

// ListeningConfig is lesson_activities.config for type=listening.
type ListeningConfig struct {
	AudioKey  string           `json:"audioKey"`
	Script    []ScriptLine     `json:"script"`
	Questions []ConfigQuestion `json:"questions"`
}

// Transcript renders the script for display ("Speaker: text" per line).
func (c ListeningConfig) Transcript() string {
	lines := make([]string, 0, len(c.Script))
	for _, l := range c.Script {
		if l.Speaker == "" {
			lines = append(lines, l.Text)
			continue
		}
		lines = append(lines, l.Speaker+": "+l.Text)
	}
	return strings.Join(lines, "\n")
}

// WritingConfig is lesson_activities.config for type=writing.
type WritingConfig struct {
	Task         string   `json:"task"` // "email" | "essay"
	Instructions string   `json:"instructions"`
	Stimulus     string   `json:"stimulus"`
	MinWords     int      `json:"minWords"`
	MaxWords     int      `json:"maxWords"`
	ModelAnswer  string   `json:"modelAnswer"`
	Checklist    []string `json:"checklist"`
}

// SpeakingItem is one read-aloud item. ModelVoice is TTS-only.
type SpeakingItem struct {
	Text          string `json:"text"`
	ModelAudioKey string `json:"modelAudioKey"`
	ModelVoice    string `json:"modelVoice"`
}

// SpeakingConfig is lesson_activities.config for type=speaking.
type SpeakingConfig struct {
	Items []SpeakingItem `json:"items"`
}

func parseConfig[T any](raw datatypes.JSON) (T, error) {
	var out T
	if len(raw) == 0 {
		return out, fmt.Errorf("empty activity config")
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("parse activity config: %w", err)
	}
	return out, nil
}

func ParseReadingConfig(raw datatypes.JSON) (ReadingConfig, error) {
	return parseConfig[ReadingConfig](raw)
}
func ParseListeningConfig(raw datatypes.JSON) (ListeningConfig, error) {
	return parseConfig[ListeningConfig](raw)
}
func ParseWritingConfig(raw datatypes.JSON) (WritingConfig, error) {
	return parseConfig[WritingConfig](raw)
}
func ParseSpeakingConfig(raw datatypes.JSON) (SpeakingConfig, error) {
	return parseConfig[SpeakingConfig](raw)
}

// ErrNotCheckable reports an activity whose type has no checkable questions.
var ErrNotCheckable = errors.New("activity type is not checkable")

// QuestionsFor returns the checkable items of a reading/listening config. The
// check endpoint is the only consumer; keeping the type switch here means a new
// checkable type is a change to this package, never to the service.
func QuestionsFor(t enums.ActivityType, raw datatypes.JSON) ([]ConfigQuestion, error) {
	switch t {
	case enums.ActivityTypeReading:
		cfg, err := ParseReadingConfig(raw)
		if err != nil {
			return nil, err
		}
		return cfg.Questions, nil
	case enums.ActivityTypeListening:
		cfg, err := ParseListeningConfig(raw)
		if err != nil {
			return nil, err
		}
		return cfg.Questions, nil
	default:
		return nil, ErrNotCheckable
	}
}
