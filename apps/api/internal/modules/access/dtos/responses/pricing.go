package responses

import "engflex-api/internal/common/enums"

type PricingCatalog struct {
	Free  PricingEntitlements `json:"free"`
	Trial PricingTrial        `json:"trial"`
	Plans []PricingPlan       `json:"plans"`
	Costs []ActionCost        `json:"costs"`
}

type PricingEntitlements struct {
	Credits       int               `json:"credits"`
	VoiceSessions int               `json:"voiceSessions"`
	CreditsPeriod enums.UsagePeriod `json:"creditsPeriod"`
	VoicePeriod   enums.UsagePeriod `json:"voicePeriod"`
}

type PricingTrial struct {
	Days          int `json:"days"`
	Credits       int `json:"credits"`
	VoiceSessions int `json:"voiceSessions"`
}

type PricingPlan struct {
	Plan          enums.SubscriptionPlan `json:"plan"`
	Months        int                    `json:"months"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Credits       int                    `json:"credits"`
	VoiceSessions int                    `json:"voiceSessions"`
}
