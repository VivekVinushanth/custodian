package models

type Event struct {
	ProfileId      string                 `json:"profile_id" bson:"perma_id"`
	EventType      string                 `json:"event_type" bson:"event_type"`
	EventName      string                 `json:"event_name" bson:"event_name"`
	EventId        string                 `json:"event_id" bson:"event_id"`
	AppId          string                 `json:"app_id" bson:"app_id"`
	OrgId          string                 `json:"org_id" bson:"org_id"`
	EventTimestamp int                    `json:"event_timestamp" bson:"event_timestamp"`
	Properties     map[string]interface{} `json:"properties,omitempty" bson:"properties,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty" bson:"context,omitempty"`
}
