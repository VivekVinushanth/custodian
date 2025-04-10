package models

type ProfileHierarchy struct {
	ParentProfileID string         `json:"parent_profile_id,omitempty" bson:"parent_profile_id,omitempty"`
	IsParent        bool           `json:"is_parent,omitempty" bson:"is_parent,omitempty"`
	ListProfile     bool           `json:"list_profile,omitempty" bson:"list_profile,omitempty"`
	ChildProfiles   []ChildProfile `json:"child_profile_ids,omitempty" bson:"child_profile_ids,omitempty"`
}

type ChildProfile struct {
	ChildProfileId string `json:"child_profile_id,omitempty" bson:"child_profile_id,omitempty"`
	RuleName       string `json:"rule_name,omitempty" bson:"rule_name,omitempty"`
}

type Profile struct {
	PermaId          string                 `json:"perma_id" bson:"perma_id"`
	OriginCountry    string                 `json:"origin_country" bson:"origin_country"`
	Identity         map[string]interface{} `json:"identity,omitempty" bson:"identity,omitempty"`
	Personality      map[string]interface{} `json:"personality,omitempty" bson:"personality,omitempty"`
	AppContext       []AppContext           `json:"app_context,omitempty" bson:"app_context,omitempty"`
	Session          map[string]interface{} `json:"session,omitempty" bson:"session,omitempty"`
	ProfileHierarchy *ProfileHierarchy      `json:"profile_hierarchy,omitempty" bson:"profile_hierarchy,omitempty"`
}

type ProfileEnrichmentRule struct {
	TraitId         string      `json:"trait_id,omitempty" bson:"trait_id,omitempty"`
	TraitName       string      `json:"trait_name" bson:"trait_name"`
	Description     string      `json:"description,omitempty" bson:"description,omitempty"`
	TraitType       string      `json:"trait_type" bson:"trait_type"`                         // static or computed
	Value           interface{} `json:"value,omitempty" bson:"value,omitempty"`               // required if trait_type == static
	ValueType       string      `json:"value_type,omitempty" bson:"value_type,omitempty"`     // required if trait_type == static
	Computation     string      `json:"computation,omitempty" bson:"computation,omitempty"`   // if trait_type == computed
	SourceField     string      `json:"source_field,omitempty" bson:"source_field,omitempty"` // if computation == copy
	MergeStrategy   string      `json:"merge_strategy" bson:"merge_strategy"`                 // overwrite, combine, ignore
	MaskingRequired bool        `json:"masking_required" bson:"masking_required"`
	MaskingStrategy string      `json:"masking_strategy,omitempty" bson:"masking_strategy,omitempty"` // optional if MaskingRequired == false
	Trigger         RuleTrigger `json:"trigger" bson:"trigger"`                                       // 🔸 grouped field
	CreatedAt       int64       `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt       int64       `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type RuleTrigger struct {
	EventType  string          `json:"event_type" bson:"event_type"`
	EventName  string          `json:"event_name" bson:"event_name"`
	Conditions []RuleCondition `json:"conditions" bson:"conditions"`
}

type RuleCondition struct {
	Field    string `json:"field" bson:"field"`
	Operator string `json:"operator" bson:"operator"`
	Value    string `json:"value" bson:"value"`
}
