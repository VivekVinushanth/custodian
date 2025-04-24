package service

import (
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"identity-customer-data-service/pkg/constants"
	"identity-customer-data-service/pkg/errors"
	"identity-customer-data-service/pkg/locks"
	"identity-customer-data-service/pkg/models"
	"identity-customer-data-service/pkg/repository"
	"identity-customer-data-service/pkg/utils"
	"net/http"
	"strings"
	"time"
)

func AddEventSchema(schema models.EventSchema) error {
	mongoDB := locks.GetMongoDBInstance()
	eventSchemaRepo := repositories.NewEventSchemaRepository(mongoDB.Database, constants.EventSchemaCollection)
	if schema.EventSchemaId == "" {
		schema.EventSchemaId = uuid.NewString()
	}
	if schema.Properties == nil {
		clientError := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrNoEventProps.Code,
			Message:     errors.ErrNoEventProps.Message,
			Description: errors.ErrNoEventProps.Description,
		}, http.StatusBadRequest)
		return clientError
	}
	for _, prop := range schema.Properties {
		if prop.PropertyName == "" || prop.PropertyType == "" {
			clientError := errors.NewClientError(errors.ErrorMessage{
				Code:        errors.ErrNoEventPropValue.Code,
				Message:     errors.ErrNoEventPropValue.Message,
				Description: errors.ErrNoEventPropValue.Description,
			}, http.StatusBadRequest)
			return clientError
		}
		if !constants.AllowedPropertyTypes[strings.ToLower(prop.PropertyType)] {
			clientError := errors.NewClientError(errors.ErrorMessage{
				Code:        errors.ErrNoEventPropValue.Code,
				Message:     errors.ErrNoEventPropValue.Message,
				Description: "Invalid property type: " + prop.PropertyType + errors.ErrNoEventPropValue.Description,
			}, http.StatusBadRequest)
			return clientError
		}
		normalizedType, err := utils.NormalizePropertyType(prop.PropertyType)
		if err != nil {
			return errors.NewClientError(errors.ErrorMessage{
				Code:        errors.ErrImproperProperty.Code,
				Message:     errors.ErrImproperProperty.Message,
				Description: "Invalid property type: " + prop.PropertyType,
			}, http.StatusBadRequest)
		}

		// Store normalized type instead of original
		prop.PropertyType = normalizedType
	}
	return eventSchemaRepo.AddEventSchema(schema)
}

func GetEventSchemas() ([]models.EventSchema, error) {
	mongoDB := locks.GetMongoDBInstance()
	eventSchemaRepo := repositories.NewEventSchemaRepository(mongoDB.Database, constants.EventSchemaCollection)
	return eventSchemaRepo.GetAllEventSchemas()
}

func GetEventSchema(id string) (*models.EventSchema, error) {
	mongoDB := locks.GetMongoDBInstance()
	eventSchemaRepo := repositories.NewEventSchemaRepository(mongoDB.Database, constants.EventSchemaCollection)
	return eventSchemaRepo.GetById(id)
}

func PatchEventSchema(id string, updates bson.M) error {
	mongoDB := locks.GetMongoDBInstance()
	eventSchemaRepo := repositories.NewEventSchemaRepository(mongoDB.Database, constants.EventSchemaCollection)
	return eventSchemaRepo.Patch(id, updates)
}

func DeleteEventSchema(id string) error {
	mongoDB := locks.GetMongoDBInstance()
	eventSchemaRepo := repositories.NewEventSchemaRepository(mongoDB.Database, constants.EventSchemaCollection)
	return eventSchemaRepo.Delete(id)
}

func AddProfileTrait(rule models.ProfileEnrichmentRule) error {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")

	if rule.TraitId == "" {
		// if it is not existing, its new. If not its an update.
		rule.TraitId = uuid.New().String()
	}
	// 🔹 Required: Trait Name
	if rule.TraitName == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10001",
			Message:     "Trait name is required.",
			Description: "The 'trait_name' field cannot be empty.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required: Trait Type
	if rule.TraitType != "static" && rule.TraitType != "computed" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10002",
			Message:     "Invalid trait type.",
			Description: "Trait type must be either 'static' or 'computed'.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required for Static: Value
	if rule.TraitType == "static" && rule.Value == "" {
		return errors.NewClientError(errors.ErrorMessage{
			Code:        "CDS-10003",
			Message:     "Missing static value.",
			Description: "For static traits, 'value' must be provided.",
		}, http.StatusBadRequest)
	}

	// 🔹 Required for Computed: Computation logic
	if rule.TraitType == "computed" && rule.Computation == "" {
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

	return schemaRepo.UpsertTrait(rule)
}

func GetProfileTraits() ([]models.ProfileEnrichmentRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.GetSchemaRules()
}

func GetProfileTrait(traitId string) (models.ProfileEnrichmentRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.GetSchemaRule(traitId)
}

func DeleteProfileTrait(ruleName string) error {
	mongoDB := locks.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.DeleteSchemaRule(ruleName)
}
