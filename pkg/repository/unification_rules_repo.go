package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"identity-customer-data-service/pkg/errors"
	"identity-customer-data-service/pkg/logger"
	"identity-customer-data-service/pkg/models"
	"time"
)

// ResolutionRuleRepository handles DB operations for resolution rules
type ResolutionRuleRepository struct {
	Collection *mongo.Collection
}

// NewResolutionRuleRepository initializes a repository
func NewResolutionRuleRepository(db *mongo.Database, collectionName string) *ResolutionRuleRepository {
	return &ResolutionRuleRepository{
		Collection: db.Collection(collectionName),
	}
}

// AddResolutionRule Inserts a new resolution rule
func (repo *ResolutionRuleRepository) AddResolutionRule(rule models.UnificationRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.Collection.InsertOne(ctx, rule)
	if err != nil {
		return errors.NewServerError(errors.ErrWhileCreatingResolutionRules, err)
	}

	logger.Log.Info("Unification rule created successfully: " + rule.RuleName)
	return nil
}

// GetResolutionRules  Retrieves all unification rules
func (repo *ResolutionRuleRepository) GetResolutionRules() ([]models.UnificationRule, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		logger.Log.Info("Error occurred while fetching resolution rules.")
		return nil, errors.NewServerError(errors.ErrWhileFetchingResolutionRules, err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		// should be debug
		if err != nil {
			logger.Log.Info("Error occurred while closing cursor.")
		}
	}(cursor, ctx)
	var rules []models.UnificationRule
	if err = cursor.All(ctx, &rules); err != nil {
		// should be debug
		logger.Log.Info("Error occurred while decoding resolution rules.")
		return nil, errors.NewServerError(errors.ErrWhileFetchingResolutionRules, err)
	}
	logger.Log.Info("Successfully fetched resolution rules")
	return rules, nil
}

// GetUnificationRule retrieves a specific resolution rule by rule_id.
func (repo *ResolutionRuleRepository) GetUnificationRule(ruleId string) (models.UnificationRule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"rule_id": ruleId}
	var rule models.UnificationRule

	err := repo.Collection.FindOne(ctx, filter).Decode(&rule)
	if err != nil {
		// should be debug
		logger.Log.Info("Error occurred while fetching resolution rule with rule_id: " + ruleId)
		return rule, errors.NewServerError(errors.ErrWhileFetchingResolutionRule, err)
	}

	logger.Log.Info("Successfully fetched resolution rule for rule_id: " + ruleId)
	return rule, nil
}

// PatchUnificationRule modifies specific fields
func (repo *ResolutionRuleRepository) PatchUnificationRule(ruleId string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now().UTC().Unix()

	filter := bson.M{"rule_id": ruleId}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.NewServerError(errors.ErrWhileUpdatingResolutionRule, err)
	}
	logger.Log.Info("Successfully updated resolution rule for rule_id: " + ruleId)
	return nil
}

// DeleteUnificationRule Removes a resolution rule.
func (repo *ResolutionRuleRepository) DeleteUnificationRule(ruleId string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"rule_id": ruleId}
	_, err := repo.Collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Log.Error(err, "Error while deleting resolution rule for rule_id: "+ruleId)
		return err
	}
	logger.Log.Info("Successfully deleted resolution rule with rule_id: " + ruleId)
	return nil
}
