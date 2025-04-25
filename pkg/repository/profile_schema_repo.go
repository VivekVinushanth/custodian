package repositories

import (
	"context"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ProfileSchemaRepository struct {
	Collection *mongo.Collection
}

func NewProfileSchemaRepository(db *mongo.Database, collection string) *ProfileSchemaRepository {
	return &ProfileSchemaRepository{
		Collection: db.Collection(collection),
	}
}

func (repo *ProfileSchemaRepository) UpsertRule(rule models.ProfileEnrichmentRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"rule_id": rule.RuleId} // assuming rule_id is unique
	update := bson.M{"$set": rule}

	opts := options.Update().SetUpsert(true)

	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *ProfileSchemaRepository) GetSchemaRules() ([]models.ProfileEnrichmentRule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.ProfileEnrichmentRule
	err = cursor.All(ctx, &rules)
	return rules, err
}

func (repo *ProfileSchemaRepository) GetSchemaRule(traitId string) (models.ProfileEnrichmentRule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_field": traitId}

	var rule models.ProfileEnrichmentRule
	err := repo.Collection.FindOne(ctx, filter).Decode(&rule)
	if err != nil {
		return models.ProfileEnrichmentRule{}, err
	}

	return rule, nil
}

func (repo *ProfileSchemaRepository) DeleteSchemaRule(attribute string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.Collection.DeleteOne(ctx, bson.M{"rule_id": attribute})
	return err
}
