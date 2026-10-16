package responses

import "time"

// User is one Clerk account row for the admin console. Clerk is the
// source of truth — there is no local users table — so this is built by
// hand from clerk.User rather than copied from a GORM model.
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	Role         string     `json:"role"`
	Banned       bool       `json:"banned"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastActiveAt *time.Time `json:"lastActiveAt"`
}
