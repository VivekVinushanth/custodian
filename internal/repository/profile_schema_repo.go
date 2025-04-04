package repositories

import (
	"context"
	"custodian/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (repo *ProfileSchemaRepository) AddSchemaRules(rules []models.ProfileEnrichmentRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var docs []interface{}
	for _, rule := range rules {
		docs = append(docs, rule)
	}

	_, err := repo.Collection.InsertMany(ctx, docs)
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

func (repo *ProfileSchemaRepository) DeleteSchemaRule(attribute string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.Collection.DeleteOne(ctx, bson.M{"attribute": attribute})
	return err
}
