package models

// Event properties for tracking user interactions
type TrackEvent struct {
	// The specific action performed (click, scroll)
	Action string `json:"action,omitempty"`
	// The type of object interacted with (button, product)
	ObjectType string `json:"object_type,omitempty"`
	// Unique identifier for the interacted object
	ObjectId string `json:"object_id,omitempty"`
	// Human-readable name of the object
	ObjectName string `json:"object_name,omitempty"`
	// A numeric value associated with the event
	Value float64 `json:"value,omitempty"`
	// Additional label for categorization
	Label string `json:"label,omitempty"`
	// Source of the interaction (website, mobile app)
	Source string `json:"source,omitempty"`
	// URL where the event occurred
	Url string `json:"url,omitempty"`
	// URL of the referring page
	Referrer string `json:"referrer,omitempty"`
}
