package enums

// SubscriptionPlan is a fixed-duration premium package. Pricing per plan
// lives in services.PlanCatalog (Go, not a DB table) — no admin UI to edit
// prices in this pass.
type SubscriptionPlan string

const (
	SubscriptionPlanOneMonth    SubscriptionPlan = "1m"
	SubscriptionPlanThreeMonth  SubscriptionPlan = "3m"
	SubscriptionPlanTwelveMonth SubscriptionPlan = "12m"
)
