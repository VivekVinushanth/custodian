package constants

// Define collection names
const (
	ResolutionRulesCollection = "resolution_rules"
	EventSchemaCollection     = "event_schemas"
	EventCollection           = "events"
	ProfileCollection         = "profiles"
	ProfileSchemaCollection   = "profile_schema"
)
const MaxRetryAttempts = 10

var AllowedPropertyTypes = map[string]bool{
	"string":        true,
	"int":           true,
	"boolean":       true,
	"date":          true,
	"arrayOfString": true,
	"arrayOfInt":    true,
	//"object":   true,
	// "arrayOfPredefinedObjects": true, // for example: array of event_type
	// todo: add more types as needed and ensure how to preserve them
}

var GoTypeMapping = map[string]string{
	"string":        "string",
	"int":           "int",
	"boolean":       "bool",
	"date":          "time.Time",
	"arrayofstring": "[]string",
	"arrayofint":    "[]int",
}

var AllowedTraitTypes = map[string]bool{
	"static":   true,
	"computed": true,
}

var AllowedMergeStrategies = map[string]bool{
	"overwrite": true,
	"combine":   true,
	"ignore":    true,
}

var AllowedMaskingStrategies = map[string]bool{
	"partial": true,
	"hash":    true,
	"redact":  true,
}

var AllowedEventTypes = map[string]bool{
	"track":    true,
	"identify": true,
	"page":     true,
}

var AllowedProfileDataScopes = map[string]bool{
	"identity":    true,
	"personality": true,
	"app_context": true,
	"session":     true,
}

var AllowedConditionOperators = map[string]bool{
	"equals":              true,
	"not_equals":          true,
	"exists":              true,
	"not_exists":          true,
	"contains":            true,
	"not_contains":        true,
	"greater_than":        true,
	"greater_than_equals": true,
	"less_than":           true,
	"less_than_equals":    true,
}
