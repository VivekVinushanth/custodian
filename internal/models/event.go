package models

import (
	"time"
)

// Event represents a user event stored in MongoDB
type Event struct {
	PermaID        string    `json:"perma_id" bson:"perma_id"`
	EventType      string    `json:"event_type" bson:"event_type"`
	EventName      string    `json:"event_name" bson:"event_name"`
	EventID        string    `json:"event_id" bson:"event_id"`
	AppID          string    `json:"app_id" bson:"app_id" binding:"required"`
	EventTimestamp time.Time `json:"event_timestamp" bson:"event_timestamp"`
	//Context        interface{}          `json:"context" bson:"context"`
	Locale     string               `json:"locale,omitempty" bson:"locale,omitempty"`
	Properties OneOfEventProperties `json:"properties,omitempty" bson:"properties,omitempty"`
}

// OneOfEventProperties represents multiple event types (PageEvent, TrackEvent, IdentifyEvent)
type OneOfEventProperties struct {
	PageEvent     *PageEvent     `json:"page_event,omitempty" bson:"page_event,omitempty"`
	TrackEvent    *TrackEvent    `json:"track_event,omitempty" bson:"track_event,omitempty"`
	IdentifyEvent *IdentifyEvent `json:"identify_event,omitempty" bson:"identify_event,omitempty"`
}
