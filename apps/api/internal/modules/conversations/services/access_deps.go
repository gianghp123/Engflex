package services

import (
	"time"

	accessservices "engflex-api/internal/modules/access/services"
)

// AccessDependencies is what voice enforcement needs from the access
// module, as one explicit bundle (spec §5.2: services own the PEP).
type AccessDependencies struct {
	Policy  accessservices.PolicyDecider
	Usage   accessservices.UsageMeter
	Enforce bool
	// Loc is the quota-period location (APP_TIMEZONE): Reserve cuts period
	// keys with it so a hold's keys always match the Decide snapshot's.
	Loc *time.Location
}
