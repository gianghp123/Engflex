package services

import (
	"time"

	"engflex-api/internal/database/repositories"
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
	// Tx coordinates the state write with its settlement: SetSpeechStart
	// and Usage.CommitTx commit atomically (spec §7.1), so a cached
	// session is never uncharged and a charge never lands without the
	// cache row.
	Tx repositories.TransactionManager
}
