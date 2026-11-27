// Package services owns the access domain: the pure policy evaluator, the
// policy/usage services, and the paid-AI orchestration (spec §5.3).
package services

import (
	"time"

	"github.com/google/uuid"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// Action is the access policy's name for the shared closed set.
type Action = enums.AccessAction

const (
	ActionVoiceStart    = enums.AccessActionVoiceStart
	ActionExamReview    = enums.AccessActionExamReview
	ActionStudyPlan     = enums.AccessActionStudyPlan
	ActionWritingScore  = enums.AccessActionWritingScore
	ActionPronunciation = enums.AccessActionPronunciation
)

// chargeTemplate is the code-owned shape of an action's debit. Only the
// AI-credit amount is overridden by the admin catalog.
type chargeTemplate struct {
	counter enums.CounterKind
	amount  int
}

var actionChargeStructure = map[Action][]chargeTemplate{
	ActionVoiceStart:    {{counter: enums.CounterKindAICredit, amount: 1}, {counter: enums.CounterKindVoiceSession, amount: 1}},
	ActionExamReview:    {{counter: enums.CounterKindAICredit, amount: 1}},
	ActionStudyPlan:     {{counter: enums.CounterKindAICredit, amount: 2}},
	ActionWritingScore:  nil,
	ActionPronunciation: nil,
}

var defaultCreditCosts = map[Action]int{
	ActionVoiceStart: 1,
	ActionExamReview: 1,
	ActionStudyPlan:  2,
}

// paidActions is the fixed order the admin/public cost lists render in.
var paidActions = []Action{ActionVoiceStart, ActionExamReview, ActionStudyPlan}

func defaultActionCreditCosts() map[Action]int {
	out := make(map[Action]int, len(defaultCreditCosts))
	for action, credits := range defaultCreditCosts {
		out[action] = credits
	}
	return out
}

// DefaultCharges is the fallback charge list (DB-outage/tests). Production
// reads PolicyService.Charges.
func DefaultCharges(action Action) []Charge {
	return buildCharges(action, defaultCreditCosts[action])
}

// DefaultCreditCost is the fallback badge cost for an action.
func DefaultCreditCost(action Action) int {
	total := 0
	for _, ch := range DefaultCharges(action) {
		if ch.Counter == enums.CounterKindAICredit {
			total += ch.Amount
		}
	}
	return total
}

func buildCharges(action Action, creditAmount int) []Charge {
	templates := actionChargeStructure[action]
	out := make([]Charge, 0, len(templates))
	for _, t := range templates {
		amount := t.amount
		if t.counter == enums.CounterKindAICredit {
			amount = creditAmount
		}
		if amount > 0 {
			out = append(out, Charge{Counter: t.counter, Amount: amount})
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

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
func Evaluate(p Principal, action Action, charges []Charge, limits map[enums.CounterKind]Limit, usage UsageSnapshot) Decision {
	d := Decision{Plan: p.Plan, Charges: charges, RequiredPlan: requiredPlanFor(action)}
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
