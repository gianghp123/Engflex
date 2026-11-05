package controllers

import (
	"engflex-api/config"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/lessons/activityconfig"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/utils"
)

// ApplyPayload fills exactly one typed payload on the DTO from the activity's
// config. A broken blob leaves every payload nil — the header still renders.
// Derived fields (media URLs, transcript) are set after the one Map.
func ApplyPayload(dto *responses.Activity, m *models.LessonActivity, media config.MediaConfig) {
	switch m.Type {
	case enums.ActivityTypeReading:
		cfg, err := activityconfig.ParseReadingConfig(m.Config)
		if err != nil {
			return
		}
		var p responses.ReadingPayload
		if utils.Map(&p, cfg) != nil {
			return
		}
		dto.Reading = &p
	case enums.ActivityTypeListening:
		cfg, err := activityconfig.ParseListeningConfig(m.Config)
		if err != nil {
			return
		}
		var p responses.ListeningPayload
		if utils.Map(&p, cfg) != nil {
			return
		}
		p.AudioURL = media.URL(cfg.AudioKey)
		p.Transcript = cfg.Transcript()
		dto.Listening = &p
	case enums.ActivityTypeWriting:
		cfg, err := activityconfig.ParseWritingConfig(m.Config)
		if err != nil {
			return
		}
		var p responses.WritingPayload
		if utils.Map(&p, cfg) != nil {
			return
		}
		dto.Writing = &p
	case enums.ActivityTypeSpeaking:
		cfg, err := activityconfig.ParseSpeakingConfig(m.Config)
		if err != nil {
			return
		}
		var p responses.SpeakingPayload
		if utils.Map(&p, cfg) != nil {
			return
		}
		for i := range p.Items {
			p.Items[i].ModelAudioURL = media.URL(cfg.Items[i].ModelAudioKey)
		}
		dto.Speaking = &p
	}
}
