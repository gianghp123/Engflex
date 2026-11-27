// Pure policy tests (spec §14 "Unit"): every plan/action/edge. Evaluate is
// deterministic and side-effect free — the whole entitlement contract is
// pinned here without a single mock.
package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	accessservices "engflex-api/internal/modules/access/services"
)

func limit(enabled bool, value int, period enums.UsagePeriod) accessservices.Limit {
	return accessservices.Limit{Enabled: enabled, Value: value, Period: period}
}

func premiumLimits() map[enums.CounterKind]accessservices.Limit {
	return map[enums.CounterKind]accessservices.Limit{
		enums.CounterKindAICredit:     limit(true, 100, enums.UsagePeriodMonth),
		enums.CounterKindVoiceSession: limit(true, 60, enums.UsagePeriodMonth),
	}
}

func TestActionCharges(t *testing.T) {
	tests := []struct {
		name   string
		action accessservices.Action
		want   []accessservices.Charge
	}{
		{name: "voice charges credit and slot", action: accessservices.ActionVoiceStart,
			want: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}, {Counter: enums.CounterKindVoiceSession, Amount: 1}}},
		{name: "exam review is one credit", action: accessservices.ActionExamReview,
			want: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}},
		{name: "study plan is two credits", action: accessservices.ActionStudyPlan,
			want: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 2}}},
		{name: "writing score is completion critical free", action: accessservices.ActionWritingScore, want: nil},
		{name: "pronunciation is completion critical free", action: accessservices.ActionPronunciation, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, accessservices.DefaultCharges(tt.action))
		})
	}
}

func TestCreditCost(t *testing.T) {
	tests := []struct {
		name   string
		action accessservices.Action
		want   int
	}{
		{name: "voice start costs one credit (session slot is not a credit)", action: accessservices.ActionVoiceStart, want: 1},
		{name: "exam review costs one credit", action: accessservices.ActionExamReview, want: 1},
		{name: "study plan costs two credits", action: accessservices.ActionStudyPlan, want: 2},
		{name: "writing score is free", action: accessservices.ActionWritingScore, want: 0},
		{name: "pronunciation is free", action: accessservices.ActionPronunciation, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, accessservices.DefaultCreditCost(tt.action))
		})
	}
}

func TestDefaultCreditCostMatchesLaunchValues(t *testing.T) {
	assert.Equal(t, 1, accessservices.DefaultCreditCost(accessservices.ActionVoiceStart))
	assert.Equal(t, 1, accessservices.DefaultCreditCost(accessservices.ActionExamReview))
	assert.Equal(t, 2, accessservices.DefaultCreditCost(accessservices.ActionStudyPlan))
	assert.Equal(t, 0, accessservices.DefaultCreditCost(accessservices.ActionWritingScore))
	assert.Equal(t, 0, accessservices.DefaultCreditCost(accessservices.ActionPronunciation))
}

func TestEvaluateMatrix(t *testing.T) {
	voiceUsage := func(credits, sessions int) accessservices.UsageSnapshot {
		return accessservices.UsageSnapshot{
			enums.CounterKindAICredit:     {Used: credits, Reserved: 0},
			enums.CounterKindVoiceSession: {Used: sessions, Reserved: 0},
		}
	}
	tests := []struct {
		name      string
		p         accessservices.Principal
		action    accessservices.Action
		limits    map[enums.CounterKind]accessservices.Limit
		usage     accessservices.UsageSnapshot
		wantAllow bool
		wantCode  string
	}{
		{name: "admin always bypasses", p: accessservices.Principal{Plan: enums.PlanKeyAdmin, IsAdmin: true},
			action: accessservices.ActionVoiceStart, limits: premiumLimits(), usage: voiceUsage(999, 999),
			wantAllow: true, wantCode: "admin_bypass"},
		{name: "premium voice allowed under both caps", p: accessservices.Principal{Plan: enums.PlanKey1M},
			action: accessservices.ActionVoiceStart, limits: premiumLimits(), usage: voiceUsage(99, 59),
			wantAllow: true},
		{name: "credits exhausted denies insufficient_credits", p: accessservices.Principal{Plan: enums.PlanKey1M},
			action: accessservices.ActionVoiceStart, limits: premiumLimits(), usage: voiceUsage(100, 10),
			wantAllow: false, wantCode: "insufficient_credits"},
		{name: "voice sub cap exhausted denies quota_exceeded", p: accessservices.Principal{Plan: enums.PlanKey1M},
			action: accessservices.ActionVoiceStart, limits: premiumLimits(), usage: voiceUsage(10, 60),
			wantAllow: false, wantCode: "quota_exceeded"},
		{name: "exam review needs only credits", p: accessservices.Principal{Plan: enums.PlanKey1M},
			action: accessservices.ActionExamReview, limits: premiumLimits(), usage: voiceUsage(100, 0),
			wantAllow: false, wantCode: "insufficient_credits"},
		{name: "study plan needs two credits", p: accessservices.Principal{Plan: enums.PlanKey1M},
			action: accessservices.ActionStudyPlan, limits: premiumLimits(), usage: voiceUsage(99, 0),
			wantAllow: false, wantCode: "insufficient_credits"},
		{name: "free writing score is always allowed", p: accessservices.Principal{Plan: enums.PlanKeyFree},
			action: accessservices.ActionWritingScore, limits: map[enums.CounterKind]accessservices.Limit{}, usage: nil,
			wantAllow: true},
		{name: "disabled counter denies feature_not_in_plan", p: accessservices.Principal{Plan: enums.PlanKeyFree},
			action: accessservices.ActionVoiceStart,
			limits: map[enums.CounterKind]accessservices.Limit{
				enums.CounterKindAICredit:     limit(false, 15, enums.UsagePeriodMonth),
				enums.CounterKindVoiceSession: limit(true, 3, enums.UsagePeriodMonth),
			}, usage: voiceUsage(0, 0),
			wantAllow: false, wantCode: "feature_not_in_plan"},
		{name: "missing limit row denies feature_not_in_plan", p: accessservices.Principal{Plan: enums.PlanKeyFree},
			action: accessservices.ActionExamReview, limits: map[enums.CounterKind]accessservices.Limit{}, usage: nil,
			wantAllow: false, wantCode: "feature_not_in_plan"},
		{name: "reserved holds count toward the limit", p: accessservices.Principal{Plan: enums.PlanKeyFree},
			action:    accessservices.ActionExamReview,
			limits:    map[enums.CounterKind]accessservices.Limit{enums.CounterKindAICredit: limit(true, 15, enums.UsagePeriodMonth)},
			usage:     accessservices.UsageSnapshot{enums.CounterKindAICredit: {Used: 10, Reserved: 5}},
			wantAllow: false, wantCode: "insufficient_credits"},
		{name: "trial totals enforce like any limit", p: accessservices.Principal{Plan: enums.PlanKeyTrial},
			action: accessservices.ActionVoiceStart,
			limits: map[enums.CounterKind]accessservices.Limit{
				enums.CounterKindAICredit:     limit(true, 30, enums.UsagePeriodTotal),
				enums.CounterKindVoiceSession: limit(true, 7, enums.UsagePeriodTotal),
			}, usage: voiceUsage(30, 3),
			wantAllow: false, wantCode: "insufficient_credits"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := accessservices.Evaluate(tt.p, tt.action, accessservices.DefaultCharges(tt.action), tt.limits, tt.usage)
			require.Equal(t, tt.wantAllow, d.Allow)
			if tt.wantCode != "" {
				assert.Equal(t, tt.wantCode, d.Code)
			}
			if d.Allow && tt.action == accessservices.ActionVoiceStart {
				assert.Equal(t, []accessservices.Charge{
					{Counter: enums.CounterKindAICredit, Amount: 1},
					{Counter: enums.CounterKindVoiceSession, Amount: 1},
				}, d.Charges)
			}
		})
	}
}

func TestEvaluateReportsTightestRemaining(t *testing.T) {
	// 99/100 credits used but 1/60 sessions: voice must report the credit
	// counter as limiting (remaining 0 beats remaining 59).
	d := accessservices.Evaluate(
		accessservices.Principal{Plan: enums.PlanKey1M},
		accessservices.ActionVoiceStart, accessservices.DefaultCharges(accessservices.ActionVoiceStart), premiumLimits(),
		accessservices.UsageSnapshot{
			enums.CounterKindAICredit:     {Used: 99, Reserved: 0},
			enums.CounterKindVoiceSession: {Used: 1, Reserved: 0},
		})
	require.True(t, d.Allow)
	assert.Equal(t, enums.CounterKindAICredit, d.LimitingCounter)
	assert.Equal(t, 0, d.Remaining)
}

func TestDefaultLimitsMatchSpecLaunchValues(t *testing.T) {
	// The boot fallback snapshot when the DB is unreachable (spec §5.6).
	got := accessservices.DefaultLimits()
	require.Equal(t, 15, got[enums.PlanKeyFree][enums.CounterKindAICredit].Value)
	require.Equal(t, 3, got[enums.PlanKeyFree][enums.CounterKindVoiceSession].Value)
	require.Equal(t, 30, got[enums.PlanKeyTrial][enums.CounterKindAICredit].Value)
	require.Equal(t, 7, got[enums.PlanKeyTrial][enums.CounterKindVoiceSession].Value)
	for _, plan := range []enums.PlanKey{enums.PlanKey1M, enums.PlanKey3M, enums.PlanKey12M} {
		require.Equal(t, 100, got[plan][enums.CounterKindAICredit].Value)
		require.Equal(t, 60, got[plan][enums.CounterKindVoiceSession].Value)
	}
}
