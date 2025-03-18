package repositories

import (
	"context"
	"custodian/internal/models"
	//"custodian/internal/pkg"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UnificationRepository handles MongoDB operations for unification rules
type UnificationRepository struct {
	Collection *mongo.Collection
}

// NewUnificationRepository initializes a repository
func NewUnificationRepository(db *mongo.Database, collectionName string) *UnificationRepository {
	return &UnificationRepository{
		Collection: db.Collection(collectionName),
	}
}

// CreateUnificationRule inserts a new rule
func (repo *UnificationRepository) CreateUnificationRule(rule models.UnificationRule) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rule.CreatedAt = time.Now().Unix()
	rule.UpdatedAt = rule.CreatedAt

	_, err := repo.Collection.InsertOne(ctx, rule)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to insert unification rule: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Unification rule created successfully: "+rule.RuleName)
	return nil
}

// UpdateUnificationRule modifies an existing rule
func (repo *UnificationRepository) UpdateUnificationRule(ruleName string, updatedRule models.UnificationRule) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updatedRule.UpdatedAt = time.Now().Unix()

	filter := bson.M{"rule_name": ruleName}
	update := bson.M{
		"$set": bson.M{
			"rules":      updatedRule.Rules,
			"is_active":  updatedRule.IsActive,
			"updated_at": updatedRule.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true) // Insert if not found
	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update unification rule: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Unification rule updated: "+ruleName)
	return nil
}

// PatchUnificationRule modifies specific fields
func (repo *UnificationRepository) PatchUnificationRule(ruleName string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now().Unix()

	filter := bson.M{"rule_name": ruleName}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch unification rule: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Unification rule patched: "+ruleName)
	return nil
}

// DeleteUnificationRule removes a rule
func (repo *UnificationRepository) DeleteUnificationRule(ruleName string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"rule_name": ruleName}
	_, err := repo.Collection.DeleteOne(ctx, filter)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to delete unification rule: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Unification rule deleted: "+ruleName)
	return nil
}

// GetUnificationRules retrieves all unification rules from MongoDB
func (repo *UnificationRepository) GetUnificationRules() ([]models.UnificationRule, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 🔹 Query MongoDB for all unification rules
	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to fetch unification rules: "+err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.UnificationRule
	// 🔹 Decode all rules
	if err = cursor.All(ctx, &rules); err != nil {
		//logger.LogMessage("ERROR", "Error decoding unification rules: "+err.Error())
		return nil, err
	}

	//logger.LogMessage("INFO", "Successfully fetched unification rules")
	return rules, nil
}
