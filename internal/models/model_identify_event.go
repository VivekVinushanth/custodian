package models

// Event properties for user identity tracking
type IdentifyEvent struct {
	// Unique identifier for the user
	UserId string `json:"user_id,omitempty"`
	// Custom user attributes
	Traits string `json:"traits,omitempty"`
}
