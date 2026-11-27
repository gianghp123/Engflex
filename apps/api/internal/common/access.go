package common

import "time"

// Launch constants for the entitlement system. The DTO validation limits
// must equal these values — access_consts_test.go guards the ones that
// cross into request binding.
// Trial snapshot fallbacks. The live values come from trial_settings and the
// trial plan_feature_limits rows; these are used only when the catalog cannot
// be loaded at boot.
const (
	DefaultTrialDays          = 7
	DefaultTrialCredits       = 30
	DefaultTrialVoiceSessions = 7
	MaxTrialDays              = 365
)

const (
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
