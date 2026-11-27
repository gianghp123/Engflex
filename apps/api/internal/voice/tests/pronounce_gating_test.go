package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/voice/pronunciation"
)

// TestGatingRecordedFixtures pins the gating policy against the recorded
// container output. The mismatch fixture is a fully wrong attempt: every word
// clears the word floor (min 0.749), the phone floor drops the two
// low-confidence phone claims (0.314, 0.498), and truncation keeps the three
// highest-confidence words. The clean fixture has no errors at all.
func TestGatingRecordedFixtures(t *testing.T) {
	t.Run("mismatch attempt is gated and truncated", func(t *testing.T) {
		engine := testPronounceEngine(t, recordedPronounceServer(t,
			readRecorded(t, "container_pronunciation.json"),
			readRecorded(t, "container_phonemes.json"),
		))
		res, err := pronunciation.NewScorer(engine).Score(context.Background(), "Good morning my name is Kenji Sato", []byte("fake-webm"), "audio/webm")
		require.NoError(t, err)
		require.Len(t, res.Errors, 7)

		gated := pronunciation.GateErrors(res.Errors)
		require.Len(t, gated, 3)
		assert.Equal(t, "good", gated[0].Word)
		assert.Equal(t, "morning", gated[1].Word)
		assert.Equal(t, "name", gated[2].Word)
		// name's /eɪ/→/æ/ claim was 0.314; the other two phones survive.
		assert.Len(t, gated[2].Phones, 2)
	})

	t.Run("clean attempt stays empty", func(t *testing.T) {
		engine := testPronounceEngine(t, recordedPronounceServer(t,
			readRecorded(t, "container_pronunciation_clean.json"),
			readRecorded(t, "container_phonemes.json"),
		))
		res, err := pronunciation.NewScorer(engine).Score(context.Background(), "Good morning my name is Kenji Sato", []byte("fake-webm"), "audio/webm")
		require.NoError(t, err)
		assert.Empty(t, res.Errors)
		assert.Empty(t, pronunciation.GateErrors(res.Errors))
	})
}
