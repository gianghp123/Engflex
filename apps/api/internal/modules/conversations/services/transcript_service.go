package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"engflex-api/internal/common"
	"engflex-api/internal/logger"
	"engflex-api/internal/voice"
)

// transcriptError maps an engine failure onto the API's error vocabulary,
// preserving the refusals and treating everything else as unavailable.
func transcriptError(ctx context.Context, action, conversationID string, err error) error {
	var status *voice.TranscriptStatusError
	if errors.As(err, &status) {
		switch status.Status {
		case http.StatusNotFound:
			logger.Report(ctx, "transcript refused", common.NotFound("no live session"), "conversationID", conversationID)
			return common.NotFound("no live session")
		case http.StatusConflict:
			logger.Report(ctx, "transcript refused", common.Conflict("nothing to act on"), "conversationID", conversationID)
			return common.Conflict("nothing to act on")
		case http.StatusUnprocessableEntity, http.StatusBadRequest:
			logger.Report(ctx, "transcript refused", common.UnprocessableEntity("unusable correction"), "conversationID", conversationID)
			return common.UnprocessableEntity("unusable correction")
		case http.StatusRequestEntityTooLarge:
			logger.Report(ctx, "transcript refused", common.New(http.StatusRequestEntityTooLarge, "audio too large"), "conversationID", conversationID)
			return common.New(http.StatusRequestEntityTooLarge, "audio too large")
		}
	}
	logger.Report(ctx, "transcript engine call failed", common.ServiceUnavailable("transcript unavailable"), "conversationID", conversationID)
	return common.ServiceUnavailable("transcript unavailable")
}

// Transcript applies one learner-initiated correction command. It is a
// pass-through on purpose: the engine owns the reviewed turn, resolves which turn
// is being corrected, and hands back its own copy of the text. Go checks that
// the conversation is the caller's, forwards, and maps the answer.
func (s *ConversationService) Transcript(ctx context.Context, userID, conversationID, action, text string) (*voice.TranscriptResult, error) {
	conv, err := s.getOwned(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	res, err := s.voice.Transcript(ctx, conversationID, action, text)
	if err != nil || res == nil {
		return nil, transcriptError(ctx, action, conversationID, err)
	}
	slog.InfoContext(ctx, "transcript command applied",
		"conversationID", conversationID, "action", action, "state", res.State)
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	return res, nil
}

// MaxTranscribeBytes mirrors the engine's MAX_AUDIO_BYTES: refuse before
// proxying so an oversize upload never reaches the engine. Exported: the
// controller enforces the same cap on the declared file size.
const MaxTranscribeBytes = 2 * 1024 * 1024

// MaxPronounceBytes mirrors the engine's MAX_AUDIO_BYTES on /pronounce: same
// browser source, same cap, separate name so each route reads clearly.
const MaxPronounceBytes = 2 * 1024 * 1024

func (s *ConversationService) Transcribe(ctx context.Context, userID, conversationID string, audio []byte, mime string) (*voice.TranscribeResult, error) {
	conv, err := s.getOwned(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	if len(audio) > MaxTranscribeBytes {
		return nil, common.New(http.StatusRequestEntityTooLarge, "audio too large")
	}
	res, err := s.voice.Transcribe(ctx, conversationID, audio, mime)
	if err != nil || res == nil {
		return nil, transcriptError(ctx, "transcribe", conversationID, err)
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	return res, nil
}

// Pronounce scores one exercise attempt against its reference sentence. Like
// Transcribe it is a pass-through: Go proves the conversation is the
// caller's, refuses oversize uploads, and maps the engine's answer.
func (s *ConversationService) Pronounce(ctx context.Context, userID, conversationID, expectedText, lang string, audio []byte, mime string) (*voice.PronounceResult, error) {
	conv, err := s.getOwned(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	if len(audio) > MaxPronounceBytes {
		return nil, common.New(http.StatusRequestEntityTooLarge, "audio too large")
	}
	res, err := s.pronounce.Pronounce(ctx, expectedText, lang, audio, mime)
	if err != nil || res == nil {
		return nil, transcriptError(ctx, "pronounce", conversationID, err)
	}
	slog.InfoContext(ctx, "pronunciation scored",
		"conversationID", conversationID, "score", res.Score)
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, err
	}
	return res, nil
}
