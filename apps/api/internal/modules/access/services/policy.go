// Package services owns the access domain: the pure policy evaluator, the
// policy/usage services, and the paid-AI orchestration (spec §5.3).
package services

import (
	"time"

	"github.com/google/uuid"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// Action is a user-facing metered or free capability. The action → charge
// mapping is Go code, not DB rows (spec §5.5): charges are program
// behavior, limits are admin pricing.
type Action string

const (
	ActionVoiceStart    Action = "voice_session_start"
	ActionExamReview    Action = "exam_ai_review"
	ActionStudyPlan     Action = "study_plan"
	ActionWritingScore  Action = "writing_score"
	ActionPronunciation Action = "pronunciation"
)

// Charge is one counter debit an action performs.
type Charge struct {
	Counter enums.CounterKind
	Amount  int
}

// Limit is one plan counter's allowance.
type Limit struct {
	Enabled bool
	Value   int
	Period  enums.UsagePeriod
}

// CounterUsage is the metered state one decision reads.
type CounterUsage struct {
	Used     int
	Reserved int
}

// UsageSnapshot carries the counters a decision needs, already read.
type UsageSnapshot map[enums.CounterKind]CounterUsage

// Principal is the resolved caller: plan and role only. Usage and limits
// stay out so Evaluate stays pure over its explicit inputs.
type Principal struct {
	UserID  string
	Plan    enums.PlanKey
	IsAdmin bool
}

// Decision is the PDP result. On deny, Code is one of the common.Code*
// strings (or admin_bypass on the admin allow path). ResetAt is the first
// instant the limiting counter rolls over (nil for never-resetting totals);
// the wire renders it RFC3339 so the UI can show "resets in …".
type Decision struct {
	Allow           bool
	Code            string
	Plan            enums.PlanKey
	Charges         []Charge
	LimitingCounter enums.CounterKind
	Limit           int
	Used            int
	Remaining       int
	ResetAt         *time.Time
	RequiredPlan    enums.PlanKey
}

// actionCharges is the single source of truth for what an action costs
// (spec §5.5). Completion-critical AI is free — never add a charge here.
var actionCharges = map[Action][]Charge{
	ActionVoiceStart:    {{Counter: enums.CounterKindAICredit, Amount: 1}, {Counter: enums.CounterKindVoiceSession, Amount: 1}},
	ActionExamReview:    {{Counter: enums.CounterKindAICredit, Amount: 1}},
	ActionStudyPlan:     {{Counter: enums.CounterKindAICredit, Amount: 2}},
	ActionWritingScore:  nil,
	ActionPronunciation: nil,
}

// ActionCharges returns the counter debits an action performs. Nil means
// free: no counter is read or written.
func ActionCharges(action Action) []Charge { return actionCharges[action] }

// CreditCost is the AI-credit badge cost an option card shows: the sum of
// the action's AI-credit charges only (voice sessions also debit a session
// slot, which is quota, not a credit badge). Free actions cost 0, and the
// UI hides a 0-cost badge.
func CreditCost(action Action) int {
	total := 0
	for _, ch := range ActionCharges(action) {
		if ch.Counter == enums.CounterKindAICredit {
			total += ch.Amount
		}
	}
	return total
}

// defaultLimits is the boot snapshot when the DB cannot be read (spec §5.6)
// — the process must never fail to start, and paid AI still fails closed
// later because the usage store is checked independently.
func defaultLimits() map[enums.PlanKey]map[enums.CounterKind]Limit {
	return map[enums.PlanKey]map[enums.CounterKind]Limit{
		enums.PlanKeyFree: {
			enums.CounterKindAICredit:     {Enabled: true, Value: 15, Period: enums.UsagePeriodMonth},
			enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
		},
		enums.PlanKeyTrial: {
			enums.CounterKindAICredit:     {Enabled: true, Value: 30, Period: enums.UsagePeriodTotal},
			enums.CounterKindVoiceSession: {Enabled: true, Value: 7, Period: enums.UsagePeriodTotal},
		},
		enums.PlanKey1M: {
			enums.CounterKindAICredit:     {Enabled: true, Value: 100, Period: enums.UsagePeriodMonth},
			enums.CounterKindVoiceSession: {Enabled: true, Value: 60, Period: enums.UsagePeriodMonth},
		},
		enums.PlanKey3M: {
			enums.CounterKindAICredit:     {Enabled: true, Value: 100, Period: enums.UsagePeriodMonth},
			enums.CounterKindVoiceSession: {Enabled: true, Value: 60, Period: enums.UsagePeriodMonth},
		},
		enums.PlanKey12M: {
			enums.CounterKindAICredit:     {Enabled: true, Value: 100, Period: enums.UsagePeriodMonth},
			enums.CounterKindVoiceSession: {Enabled: true, Value: 60, Period: enums.UsagePeriodMonth},
		},
		enums.PlanKeyAdmin: {}, // bypass: the evaluator short-circuits admins
	}
}

// DefaultLimits exposes the boot fallback for the snapshot loader and tests.
func DefaultLimits() map[enums.PlanKey]map[enums.CounterKind]Limit {
	return defaultLimits()
}

// Evaluate is the pure PDP (spec §5.5 evaluation order). No I/O, no clock:
// period keys and resetAt are resolved by the caller.
func Evaluate(p Principal, action Action, limits map[enums.CounterKind]Limit, usage UsageSnapshot) Decision {
	d := Decision{Plan: p.Plan, Charges: ActionCharges(action), RequiredPlan: requiredPlanFor(action)}
	if p.IsAdmin || p.Plan == enums.PlanKeyAdmin {
		d.Allow = true
		d.Code = "admin_bypass"
		return d
	}
	if len(d.Charges) == 0 {
		// Completion-critical and unmetered actions (spec §3.3): free for
		// every plan, no counter consulted.
		d.Allow = true
		return d
	}
	tightest := -1
	for _, ch := range d.Charges {
		lim, ok := limits[ch.Counter]
		if !ok || !lim.Enabled || lim.Value <= 0 {
			d.Allow = false
			d.Code = common.CodeFeatureNotInPlan
			d.LimitingCounter = ch.Counter
			return d
		}
		u := usage[ch.Counter]
		used := u.Used + u.Reserved
		if used+ch.Amount > lim.Value {
			d.Allow = false
			if ch.Counter == enums.CounterKindAICredit {
				d.Code = common.CodeInsufficientCredits
			} else {
				d.Code = common.CodeQuotaExceeded
			}
			d.LimitingCounter = ch.Counter
			d.Limit = lim.Value
			d.Used = used
			d.Remaining = lim.Value - used
			return d
		}
		remaining := lim.Value - used - ch.Amount
		if tightest == -1 || remaining < tightest {
			tightest = remaining
			d.LimitingCounter = ch.Counter
			d.Limit = lim.Value
			d.Used = used
			d.Remaining = remaining
		}
	}
	d.Allow = true
	return d
}

// requiredPlanFor names the cheapest plan that includes the action — the
// upgrade CTA target (paid for everything metered at launch).
func requiredPlanFor(action Action) enums.PlanKey {
	switch action {
	case ActionWritingScore, ActionPronunciation:
		return enums.PlanKeyFree
	default:
		return enums.PlanKey1M
	}
}

// NewReservationID mints the group id a reservation's event rows share.
func NewReservationID() string { return uuid.NewString() }
