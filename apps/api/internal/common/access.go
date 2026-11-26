package common

import "time"

// Launch constants for the entitlement system. The DTO validation limits
// must equal these values — access_consts_test.go guards the ones that
// cross into request binding.
const (
	// Trial: once per Clerk user, no card, 7 days, one-time credits.
	TrialDays         = 7
	TrialCredits      = 30
	TrialVoiceSession = 7

	// Reservation leases: a hold older than its lease is expired by the
	// sweeper; a crash can never block a user permanently.
	VoiceLeaseTTL  = 2 * time.Minute
	ReviewLeaseTTL = 5 * time.Minute

	// Sweeper cadence and batch.
	SweeperInterval = time.Minute
	SweeperBatch    = 100

	// MaxPlanLimitValue caps what an admin may set any plan counter limit
	// to (request binding ceiling — see the admin PUT request DTO).
	MaxPlanLimitValue = 100000
)
