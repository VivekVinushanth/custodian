package models

import (
	"time"
)

type IdentityData struct {
	UserId string `json:"userId,omitempty"`

	Username string `json:"username,omitempty"`

	Email string `json:"email,omitempty"`

	PhoneNumbers []string `json:"phone_numbers,omitempty"`

	FirstName string `json:"first_name,omitempty"`

	LastName string `json:"last_name,omitempty"`

	DisplayName string `json:"display_name,omitempty"`

	PreferredUsername string `json:"preferred_username,omitempty"`

	ProfileUrl string `json:"profile_url,omitempty"`

	Picture string `json:"picture,omitempty"`

	Roles []string `json:"roles,omitempty"`

	Groups []string `json:"groups,omitempty"`

	AccountStatus string `json:"account_status,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty"`

	IdpProvider string `json:"idp_provider,omitempty"`

	MfaEnabled bool `json:"mfa_enabled,omitempty"`

	LastLogin time.Time `json:"last_login,omitempty"`

	Locale string `json:"locale,omitempty"`

	Timezone string `json:"timezone,omitempty"`
}
