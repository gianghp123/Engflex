package enums

// UsagePeriod is the window a limit applies over. month cuts at 00:00 in
// APP_TIMEZONE (YYYY-MM key), day at YYYY-MM-DD, total never rolls
// (trial credits, key "all").
type UsagePeriod string

const (
	UsagePeriodDay   UsagePeriod = "day"
	UsagePeriodMonth UsagePeriod = "month"
	UsagePeriodTotal UsagePeriod = "total"
)
