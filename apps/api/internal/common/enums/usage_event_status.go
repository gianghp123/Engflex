package enums

// UsageEventStatus is the reservation lifecycle. Rows are retained for
// audit; released/expired rows are reused on retry, never deleted.
type UsageEventStatus string

const (
	UsageEventStatusReserved  UsageEventStatus = "reserved"
	UsageEventStatusCommitted UsageEventStatus = "committed"
	UsageEventStatusReleased  UsageEventStatus = "released"
	UsageEventStatusExpired   UsageEventStatus = "expired"
)
