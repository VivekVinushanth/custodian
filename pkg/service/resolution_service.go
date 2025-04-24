package service

import (
	"go.mongodb.org/mongo-driver/bson"
	"identity-customer-data-service/pkg/constants"
	"identity-customer-data-service/pkg/errors"
	"identity-customer-data-service/pkg/locks"
	"identity-customer-data-service/pkg/models"
	"identity-customer-data-service/pkg/repository"
	"net/http"
	"time"
)

// AddResolutionRule Adds a new resolution rule.
func AddResolutionRule(rule models.UnificationRule) error {
	mongoDB := locks.GetMongoDBInstance()
	resolutionRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)
	rule.CreatedAt = time.Now().UTC().Unix()
	rule.UpdatedAt = rule.CreatedAt
	//todo: check if rule already exists and also the property is added available in profile
	return resolutionRepo.AddResolutionRule(rule)
}

// GetResolutionRules Fetches all resolution rules.
func GetResolutionRules() ([]models.UnificationRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	resolutionRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)
	rules, err := resolutionRepo.GetResolutionRules()
	if rules == nil {
		clientError := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrNoResolutionRules.Code,
			Message:     errors.ErrNoResolutionRules.Message,
			Description: errors.ErrNoResolutionRules.Description,
		}, http.StatusNotFound)

		return nil, clientError
	}
	return rules, err
}

// GetResolutionRule Fetches a specific resolution rule.
func GetResolutionRule(ruleId string) (models.UnificationRule, error) {
	mongoDB := locks.GetMongoDBInstance()
	resolutionRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)
	rule, err := resolutionRepo.GetUnificationRule(ruleId)
	if rule.RuleId == "" {
		clientError := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrResolutionRuleNotFound.Code,
			Message:     errors.ErrResolutionRuleNotFound.Message,
			Description: errors.ErrResolutionRuleNotFound.Description,
		}, http.StatusNotFound)

		return rule, clientError
	}
	return rule, err
}

// PatchResolutionRule Applies a partial update on a specific resolution rule.
func PatchResolutionRule(ruleId string, updates bson.M) error {
	mongoDB := locks.GetMongoDBInstance()
	resolutionRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)

	// Validate only "is_active" is being patched
	allowedFields := map[string]bool{"is_active": true}
	for field := range updates {
		if !allowedFields[field] {
			clientError := errors.NewClientError(errors.ErrorMessage{
				Code:        errors.ErrOnlyStatusUpdatePossible.Code,
				Message:     errors.ErrOnlyStatusUpdatePossible.Message,
				Description: errors.ErrOnlyStatusUpdatePossible.Description,
			}, http.StatusBadRequest)
			return clientError
		}
	}

	return resolutionRepo.PatchUnificationRule(ruleId, updates)
}

// DeleteResolutionRule Removes a  resolution rule.
func DeleteResolutionRule(ruleId string) error {
	mongoDB := locks.GetMongoDBInstance()
	resolutionRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)
	return resolutionRepo.DeleteUnificationRule(ruleId)
}
