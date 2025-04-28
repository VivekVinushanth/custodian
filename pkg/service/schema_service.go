package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/wso2/identity-customer-data-service/pkg/constants"
	"github.com/wso2/identity-customer-data-service/pkg/errors"
	"github.com/wso2/identity-customer-data-service/pkg/locks"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/repository"
	"net/http"
	"strings"
	"time"
)

func AddEnrichmentRule(rule models.ProfileEnrichmentRule) error {

	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, constants.ProfileSchemaCollection)

	if rule.RuleId == "" {
		// if it is not existing, its new. If not its an update.
		rule.RuleId = uuid.New().String()
	}
	// 🔹 Required: Trait Name
	if rule.TraitName == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10001",
			Message:     "Trait name is required.",
			Description: "The 'trait_name' field cannot be empty.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required: Rule Type
	if rule.RuleType != "static" && rule.RuleType != "computed" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10002",
			Message:     "Invalid rule type.",
			Description: "Trait type must be either 'static' or 'computed'.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required for Static: Value
	if rule.RuleType == "static" && rule.Value == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10003",
			Message:     "Missing static value.",
			Description: "For static traits, 'value' must be provided.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required for Computed: Computation logic
	if rule.RuleType == "computed" && rule.Computation == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10004",
			Message:     "Missing computation logic.",
			Description: "For computed traits, 'computation' must be provided.",
		}, http.StatusBadRequest)
	}

	if rule.Computation == "copy" && len(rule.SourceFields) != 1 {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10004",
			Message:     "Missing source field",
			Description: "For copy computation, 'source field' must be provided.",
		}, http.StatusBadRequest)
	}

	// 🔹 Validate Trigger
	if rule.Trigger.EventType == "" || rule.Trigger.EventName == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10005",
			Message:     "Invalid trigger definition.",
			Description: "Both 'event_type' and 'event_name' must be provided inside trigger.",
		}, http.StatusBadRequest)
	}

	// 🔹 Validate Trigger Conditions
	for _, cond := range rule.Trigger.Conditions {
		if cond.Field == "" || cond.Operator == "" {
			return errors.NewClientError(errors.ErrorMessage{
				Code:        "CDS-10006",
				Message:     "Invalid trigger condition.",
				Description: "Each condition must have a field and operator defined.",
			}, http.StatusBadRequest)
		}
		if !constants.AllowedConditionOperators[strings.ToLower(cond.Operator)] {
			return errors.NewClientError(errors.ErrorMessage{
				Code:        "CDS-10007",
				Message:     "Unsupported operator.",
				Description: fmt.Sprintf("Operator '%s' is not supported.", cond.Operator),
			}, http.StatusBadRequest)
		}
	}

	// 🔹 Validate Merge Strategy
	if rule.MergeStrategy != "" && !constants.AllowedMergeStrategies[strings.ToLower(rule.MergeStrategy)] {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10008",
			Message:     "Invalid merge strategy.",
			Description: fmt.Sprintf("Merge strategy '%s' is not allowed.", rule.MergeStrategy),
		}, http.StatusBadRequest)
	}

	// 🔹 Validate Masking
	if rule.MaskingRequired {
		if rule.MaskingStrategy == "" {
			return errors.NewClientError(errors.ErrorMessage{
				Code:        "CDS-10009",
				Message:     "Missing masking strategy.",
				Description: "Masking is required, but no strategy was provided.",
			}, http.StatusBadRequest)
		}
		if !constants.AllowedMaskingStrategies[strings.ToLower(rule.MaskingStrategy)] {
			return errors.NewClientError(errors.ErrorMessage{
				Code:        "CDS-10010",
				Message:     "Invalid masking strategy.",
				Description: fmt.Sprintf("Masking strategy '%s' is not supported.", rule.MaskingStrategy),
			}, http.StatusBadRequest)
		}
	}

	rule.CreatedAt = time.Now().UTC().Unix()
	rule.UpdatedAt = time.Now().UTC().Unix()

	return schemaRepo.UpsertRule(rule)
}

func GetEnrichmentRules() ([]models.ProfileEnrichmentRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, constants.ProfileSchemaCollection)
	return schemaRepo.GetSchemaRules()
}

func GetEnrichmentRulesByFilter(filters []string) ([]models.ProfileEnrichmentRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, constants.ProfileSchemaCollection)
	return schemaRepo.GetEnrichmentRulesByFilter(filters)
}

func GetEnrichmentRule(traitId string) (models.ProfileEnrichmentRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, constants.ProfileSchemaCollection)
	return schemaRepo.GetSchemaRule(traitId)
}

func DeleteEnrichmentRule(ruleId string) error {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, constants.ProfileSchemaCollection)
	return schemaRepo.DeleteSchemaRule(ruleId)
}
