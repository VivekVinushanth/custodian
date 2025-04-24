package errors

var (
	// Server error codes
	// Error codes for resolution rules
	ErrWhileCreatingResolutionRules = ErrorMessage{
		Code:        "15001",
		Message:     "Error while adding resolution rules.",
		Description: "Error while adding resolution rules for the organization.",
	}

	ErrWhileFetchingResolutionRules = ErrorMessage{
		Code:        "15002",
		Message:     "Error while fetching resolution rules.",
		Description: "Error while fetching resolution rules of the organization.",
	}

	ErrWhileFetchingResolutionRule = ErrorMessage{
		Code:        "15003",
		Message:     "Error while fetching resolution rule.",
		Description: "Error while fetching resolution rule of the organization.",
	}

	ErrWhileUpdatingResolutionRule = ErrorMessage{
		Code:        "15004",
		Message:     "Error while updating resolution rule.",
		Description: "Error while updating resolution rule of the organization.",
	}

	ErrBadRequest = ErrorMessage{
		Code:    "11001",
		Message: "Invalid body format.",
	}

	ErrNoResolutionRules = ErrorMessage{
		Code:        "11002",
		Message:     "No resolution rules defined for this organization.",
		Description: "Error while retrieving API resources from the database.",
	}

	ErrResolutionRuleNotFound = ErrorMessage{
		Code:        "11003",
		Message:     "No resolution rule found.",
		Description: "No resolution rule defined for this organization for the provided rule_id..",
	}

	ErrOnlyStatusUpdatePossible = ErrorMessage{
		Code:        "11004",
		Message:     "Only status update is possible.",
		Description: "Only 'is_active' field can be updated.",
	}

	ErrNoEventProps = ErrorMessage{
		Code:        "11005",
		Message:     "No event properties.",
		Description: "At least one event property should be added to the event schema.",
	}

	ErrNoEventPropValue = ErrorMessage{
		Code:        "11005",
		Message:     "No event properties.",
		Description: "Property %s must have both name and type",
	}

	ErrImproperProperty = ErrorMessage{
		Code:        "11005",
		Message:     "Improper property name or type.",
		Description: "Allowed types are: string, int, boolean, timestamp, array",
	}

	ErrValidationProfileTrait = ErrorMessage{
		Code:        "11005",
		Message:     "Invalid value for the profile trait.",
		Description: "Allowed types are: string, int, boolean, timestamp, array",
	}
)
