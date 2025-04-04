package models

type ProfileHierarchy struct {
	ParentProfileID string        `json:"parent_profile_id,omitempty" bson:"parent_profile_id,omitempty"`
	IsMaster        bool          `json:"is_master,omitempty" bson:"is_master,omitempty"`
	ListProfile     bool          `json:"list_profile,omitempty" bson:"list_profile,omitempty"`
	PeerProfiles    []PeerProfile `json:"peer_profile_ids,omitempty" bson:"peer_profile_ids,omitempty"`
	ChildProfiles   []string      `json:"child_profile_ids,omitempty" bson:"child_profile_ids,omitempty"`
}

type PeerProfile struct {
	PeerProfileId string `json:"peer_profile_id,omitempty" bson:"peer_profile_id,omitempty"`
	RuleName      string `json:"rule_name,omitempty" bson:"rule_name,omitempty"`
}

type Profile struct {
	PermaID          string                 `json:"perma_id" bson:"perma_id"`
	OriginCountry    string                 `json:"origin_country" bson:"origin_country"`
	UserIds          []string               `json:"user_ids,omitempty" bson:"user_ids,omitempty"`
	Identity         *IdentityData          `json:"identity,omitempty" bson:"identity,omitempty"`
	Personality      *PersonalityData       `json:"personality,omitempty" bson:"personality,omitempty"`
	AppContext       []AppContext           `json:"app_context,omitempty" bson:"app_context,omitempty"`
	Session          map[string]interface{} `json:"session,omitempty" bson:"session,omitempty"`
	ProfileHierarchy *ProfileHierarchy      `json:"profile_hierarchy,omitempty" bson:"profile_hierarchy,omitempty"`
}

type ProfileEnrichmentRule struct {
	ID           string          `json:"id,omitempty" bson:"_id,omitempty"`
	ProfileField string          `json:"profile_field" bson:"profile_field"` // e.g., personality.interests
	EventName    string          `json:"event_name" bson:"event_name"`       // e.g., category_searched
	EventType    string          `json:"event_type" bson:"event_type"`       // e.g., track
	Conditions   []RuleCondition `json:"conditions" bson:"conditions"`       // e.g. [{field: "action", op: "eq", value: "select_category"}]
}

type RuleCondition struct {
	Field    string `json:"field" bson:"field"`       // e.g. action
	Operator string `json:"operator" bson:"operator"` // eq, contains, etc.
	Value    string `json:"value" bson:"value"`       // e.g. select_category
}
