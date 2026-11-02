package identity

import (
	"encoding/json"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"github.com/clerk/clerk-sdk-go/v2"
	"time"
)

type publicMetadata struct {
	Role string `json:"role"`
}

func DirectoryUser(u *clerk.User) *models.DirectoryUser {
	email := ""
	for _, addr := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && addr.ID == *u.PrimaryEmailAddressID {
			email = addr.EmailAddress
			break
		}
	}
	if email == "" && len(u.EmailAddresses) > 0 {
		email = u.EmailAddresses[0].EmailAddress
	}

	var meta publicMetadata
	_ = json.Unmarshal(u.PublicMetadata, &meta)
	role := meta.Role
	if role == "" {
		role = string(enums.UserRoleUser)
	}

	firstName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	lastName := ""
	if u.LastName != nil {
		lastName = *u.LastName
	}

	var lastActive *time.Time
	if u.LastActiveAt != nil {
		t := time.UnixMilli(*u.LastActiveAt)
		lastActive = &t
	}

	return &models.DirectoryUser{
		ID:           u.ID,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		Banned:       u.Banned,
		CreatedAt:    time.UnixMilli(u.CreatedAt),
		LastActiveAt: lastActive,
	}
}

func DirectoryUsers(users []*clerk.User) []*models.DirectoryUser {
	out := make([]*models.DirectoryUser, 0, len(users))
	for _, user := range users {
		out = append(out, DirectoryUser(user))
	}
	return out
}
