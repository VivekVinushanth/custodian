package models

// Event properties for tracking user interactions
type TrackEvent struct {
	// The specific action performed (click, scroll)
	Action string `json:"action,omitempty"`
	// The type of object interacted with (button, product)
	ObjectType string `json:"objecttype,omitempty"`
	// Unique identifier for the interacted object
	ObjectId string `json:"objectid,omitempty"`
	// Human-readable name of the object
	ObjectName string `json:"objectname,omitempty"`
	// A numeric value associated with the event
	Value string `json:"value,omitempty"`
	// Additional label for categorization
	Label string `json:"label,omitempty"`
	// Source of the interaction (website, mobile app)
	Source string `json:"source,omitempty"`
	// URL where the event occurred
	Url string `json:"url,omitempty"`
	// URL of the referring page
	Referrer string `json:"referrer,omitempty"`
}
