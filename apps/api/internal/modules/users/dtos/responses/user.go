package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// User is one Clerk account row for the admin console. Clerk is the
// source of truth — there is no local users table — so this is built by
// hand from clerk.User rather than copied from a GORM model.
type User struct {
	ID           string               `json:"id"`
	Email        string               `json:"email"`
	FirstName    string               `json:"firstName"`
	LastName     string               `json:"lastName"`
	Role         string               `json:"role"`
	Banned       bool                 `json:"banned"`
	CreatedAt    time.Time            `json:"createdAt"`
	LastActiveAt *time.Time           `json:"lastActiveAt"`
	Subscription *SubscriptionSummary `json:"subscription"`
}

// UserAttemptLesson provides the lesson title for a user attempt.
type UserAttemptLesson struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UserAttemptExamTest provides the exam test title for a user attempt.
type UserAttemptExamTest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UserAttemptVideoExercise provides the video exercise title for a user attempt.
type UserAttemptVideoExercise struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UserAttempt is one learning run for the admin attempt history view.
type UserAttempt struct {
	ID            string                    `json:"id"`
	Type          enums.AttemptType         `json:"type"`
	Status        enums.AttemptStatus       `json:"status"`
	Score         *float64                  `json:"score,omitempty"`
	DurationSec   *int                      `json:"durationSec,omitempty"`
	StartedAt     *time.Time                `json:"startedAt,omitempty"`
	SubmittedAt   *time.Time                `json:"submittedAt,omitempty"`
	CreatedAt     time.Time                 `json:"createdAt"`
	Lesson        *UserAttemptLesson        `json:"lesson,omitempty"`
	ExamTest      *UserAttemptExamTest      `json:"examTest,omitempty"`
	VideoExercise *UserAttemptVideoExercise `json:"videoExercise,omitempty"`
}

type SubscriptionSummary struct {
	Plan      enums.SubscriptionPlan   `json:"plan"`
	Status    enums.SubscriptionStatus `json:"status"`
	ExpiresAt *time.Time               `json:"expiresAt"`
}

type Subscription struct {
	ID        string                   `json:"id"`
	UserID    string                   `json:"userId"`
	Plan      enums.SubscriptionPlan   `json:"plan"`
	Status    enums.SubscriptionStatus `json:"status"`
	Amount    int64                    `json:"amount"`
	Currency  string                   `json:"currency"`
	StartedAt *time.Time               `json:"startedAt"`
	ExpiresAt *time.Time               `json:"expiresAt"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
} // @name UserSubscription
