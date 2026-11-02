package models

import "time"

// DirectoryUser is a read graph owned by the external Clerk directory.
// It has no local table and must never be migrated or persisted by GORM.
type DirectoryUser struct {
	ID           string
	Email        string
	FirstName    string
	LastName     string
	Role         string
	Banned       bool
	CreatedAt    time.Time
	LastActiveAt *time.Time
	Subscription *Subscription
}
