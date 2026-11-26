package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// UsageEvent is one counter charge's lifecycle row (spec §6.3): the audit
// ledger AND the reservation record. (user_id, idempotency_key) is unique —
// the exactly-once guard; rows are reused on retry after release/expiry,
// never deleted.
type UsageEvent struct {
	ID             string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReservationID  string                 `gorm:"type:uuid;not null" json:"reservationId"`
	UserID         string                 `gorm:"not null;uniqueIndex:idx_usage_events_user_idem" json:"userId"`
	CounterKey     enums.CounterKind      `gorm:"not null;type:text" json:"counterKey"`
	PeriodKey      string                 `gorm:"not null" json:"periodKey"`
	Amount         int                    `gorm:"not null" json:"amount"`
	Status         enums.UsageEventStatus `gorm:"not null" json:"status"`
	IdempotencyKey string                 `gorm:"not null;uniqueIndex:idx_usage_events_user_idem" json:"idempotencyKey"`
	RefType        enums.UsageRefType     `gorm:"not null" json:"refType"`
	RefID          string                 `gorm:"not null;default:''" json:"refId"`
	ExpiresAt      *time.Time             `json:"expiresAt"`
	CreatedAt      time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	SettledAt      *time.Time             `gorm:"autoUpdateTime" json:"settledAt"`
}

func (UsageEvent) TableName() string { return "usage_events" }
