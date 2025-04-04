package service

import (
	//"context"
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
)

func CreateSchemaRules(rules []models.ProfileEnrichmentRule) error {
	mongoDB := pkg.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.AddSchemaRules(rules)
}

func GetSchemaRules() ([]models.ProfileEnrichmentRule, error) {
	mongoDB := pkg.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.GetSchemaRules()
}

func DeleteSchemaRule(ruleName string) error {
	mongoDB := pkg.GetMongoDBInstance()
	schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
	return schemaRepo.DeleteSchemaRule(ruleName)
}
