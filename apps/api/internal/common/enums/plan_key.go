package enums

// PlanKey is the entitlement plan a principal resolves to: the paid
// subscription plan keys, the free default, the trial, and the admin
// bypass. Wire values are stable (contracts generate from these).
type PlanKey string

const (
	PlanKeyFree  PlanKey = "free"
	PlanKeyTrial PlanKey = "trial"
	PlanKey1M    PlanKey = "1m"
	PlanKey3M    PlanKey = "3m"
	PlanKey12M   PlanKey = "12m"
	PlanKeyAdmin PlanKey = "admin"
)
