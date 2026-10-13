package services

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"

	"engflex-api/internal/common"
	"engflex-api/internal/logger"
)

// CommandRunner abstracts process execution so tests can fake the uv/python
// subprocess without actually invoking it.
type CommandRunner interface {
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner runs real OS processes via os/exec.
type ExecRunner struct{}

func (ExecRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// AudioSynthesisService integrates the existing scripts/seed_audio.py TTS
// pipeline (edge-tts + MinIO upload) behind an admin-triggerable call,
// scoped to one activity via --activity-id. No synthesis logic is
// duplicated in Go: the script stays the single source of truth for how
// listening/speaking audio is built, same as `make seed-audio`.
type AudioSynthesisService interface {
	SynthesizeForActivity(ctx context.Context, activityID string) error
}

type audioSynthesisService struct {
	runner CommandRunner
}

// NewAudioSynthesisService builds the service over the given process runner
// (ExecRunner in production).
func NewAudioSynthesisService(runner CommandRunner) AudioSynthesisService {
	return &audioSynthesisService{runner: runner}
}

func (s *audioSynthesisService) SynthesizeForActivity(ctx context.Context, activityID string) error {
	if _, err := uuid.Parse(activityID); err != nil {
		return common.BadRequest("invalid activity id")
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Mirrors `make seed-audio` exactly (same ad-hoc deps, same script) with
	// --activity-id added to scope it to one activity. activityID is
	// UUID-validated above and passed as its own argv entry (never shell
	// string concatenation), so this cannot be used for command injection.
	out, err := s.runner.CombinedOutput(ctx, "uv",
		"run", "--with", "edge-tts", "--with", "minio", "--with", "psycopg[binary]", "--with", "python-dotenv",
		"scripts/seed_audio.py", "--activity-id="+activityID,
	)
	output := strings.TrimSpace(string(out))
	if err != nil {
		msg := output
		if msg == "" {
			msg = err.Error()
		}
		appErr := common.UnprocessableEntity(msg)
		logger.Report(ctx, "lesson audio synthesis failed", appErr, "activityID", activityID)
		return appErr
	}
	slog.InfoContext(ctx, "lesson audio synthesized", "activityID", activityID, "output", output)
	return nil
}
