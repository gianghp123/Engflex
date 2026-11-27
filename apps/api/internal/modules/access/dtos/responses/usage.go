// Package responses holds the access module's wire DTOs (tygo inputs).
package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// UsageCounterView is one metered resource's state on GET /premium/usage
// (spec §10). Same shape as CounterView plus the counter key; ResetAt is
// omitted for never-resetting total periods (trial).
type UsageCounterView struct {
	Counter   enums.CounterKind `json:"counter"`
	Limit     int               `json:"limit"`
	Used      int               `json:"used"`
	Reserved  int               `json:"reserved"`
	Remaining int               `json:"remaining"`
	ResetAt   *time.Time        `json:"resetAt,omitempty"`
}

// UsageHistoryItem is one settled charge on GET /premium/usage: the
// user-facing action name, the debited credits, the event lifecycle status,
// and the entity the charge belongs to. Holds (reserved) rows explain
// missing credits; released/expired rows are non-charges and never appear.
type UsageHistoryItem struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Credits   int                    `json:"credits"`
	Status    enums.UsageEventStatus `json:"status"`
	CreatedAt time.Time              `json:"createdAt"`
	RefType   enums.UsageRefType     `json:"refType"`
	RefID     string                 `json:"refId"`
}

// UsageHistory is the paginated event tail on GET /premium/usage.
type UsageHistory struct {
	Items    []UsageHistoryItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	Total    int64              `json:"total"`
}

// MyUsage is the GET /premium/usage body (spec §10): the caller's plan, both
// counter views, the per-action costs table, and the paginated history of
// settled charges (committed + reserved, newest first).
type MyUsage struct {
	Plan     enums.PlanKey      `json:"plan"`
	Counters []UsageCounterView `json:"counters"`
	Costs    []ActionCost       `json:"costs"`
	History  UsageHistory       `json:"history"`
}
