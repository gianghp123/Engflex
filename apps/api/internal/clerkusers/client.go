// Package clerkusers wraps the Clerk Users Backend API behind a small,
// mockable interface so callers never touch the SDK's raw clerk.User shape
// (nested email list, json.RawMessage metadata, epoch-millis timestamps).
package clerkusers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

// User is the normalized view of a Clerk account this codebase needs.
type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Role      string
	CreatedAt time.Time
}

// Client reads Clerk accounts. Read-only by design: account mutation
// (ban/lock) stays a manual Clerk Dashboard action, not an API surface here.
type Client interface {
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)
	Get(ctx context.Context, id string) (*User, error)
}

type httpClient struct {
	sdk *user.Client
}

// NewHTTPClient builds the client over the Clerk SDK's default backend
// (the package-level secret key set once via clerk.SetKey in main).
func NewHTTPClient() Client {
	return &httpClient{sdk: user.NewClient(&clerk.ClientConfig{})}
}

func (c *httpClient) List(ctx context.Context, limit, offset int) ([]*User, int64, error) {
	l, o := int64(limit), int64(offset)
	res, err := c.sdk.List(ctx, &user.ListParams{
		ListParams: clerk.ListParams{Limit: &l, Offset: &o},
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*User, 0, len(res.Users))
	for _, u := range res.Users {
		out = append(out, fromClerkUser(u))
	}
	return out, res.TotalCount, nil
}

func (c *httpClient) Get(ctx context.Context, id string) (*User, error) {
	u, err := c.sdk.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromClerkUser(u), nil
}

func fromClerkUser(u *clerk.User) *User {
	email := ""
	if u.PrimaryEmailAddressID != nil {
		for _, e := range u.EmailAddresses {
			if e.ID == *u.PrimaryEmailAddressID {
				email = e.EmailAddress
				break
			}
		}
	}
	var meta struct {
		Role string `json:"role"`
	}
	_ = json.Unmarshal(u.PublicMetadata, &meta)
	first, last := "", ""
	if u.FirstName != nil {
		first = *u.FirstName
	}
	if u.LastName != nil {
		last = *u.LastName
	}
	return &User{
		ID:        u.ID,
		Email:     email,
		FirstName: first,
		LastName:  last,
		Role:      meta.Role,
		CreatedAt: time.UnixMilli(u.CreatedAt),
	}
}
