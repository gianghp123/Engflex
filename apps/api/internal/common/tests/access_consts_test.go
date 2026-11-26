package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"engflex-api/internal/common"
)

func TestAccessLaunchConstants(t *testing.T) {
	assert.Equal(t, 7, common.TrialDays)
	assert.Equal(t, 30, common.TrialCredits)
	assert.Equal(t, 7, common.TrialVoiceSession)
	assert.Equal(t, 2*time.Minute, common.VoiceLeaseTTL)
	assert.Equal(t, 5*time.Minute, common.ReviewLeaseTTL)
	assert.Equal(t, time.Minute, common.SweeperInterval)
	assert.Less(t, common.SweeperBatch, 1000)
	assert.Greater(t, common.MaxPlanLimitValue, 1000, "room for a generous premium cap without allowing absurd values")
}
