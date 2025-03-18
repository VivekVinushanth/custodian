package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateUnificationRule stores a new rule
func CreateUnificationRule(rule models.UnificationRule) error {
	mongoDB := pkg.GetMongoDBInstance()
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	return unificationRepo.CreateUnificationRule(rule)
}

// UpdateUnificationRule modifies a rule
func UpdateUnificationRule(ruleName string, updatedRule models.UnificationRule) error {
	mongoDB := pkg.GetMongoDBInstance()
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	return unificationRepo.UpdateUnificationRule(ruleName, updatedRule)
}

// PatchUnificationRule applies partial updates
func PatchUnificationRule(ruleName string, updates bson.M) error {
	mongoDB := pkg.GetMongoDBInstance()
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	return unificationRepo.PatchUnificationRule(ruleName, updates)
}

// DeleteUnificationRule removes a rule
func DeleteUnificationRule(ruleName string) error {
	mongoDB := pkg.GetMongoDBInstance()
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	return unificationRepo.DeleteUnificationRule(ruleName)
}

// GetUnificationRules fetches all unification rules
func GetUnificationRules() ([]models.UnificationRule, error) {
	mongoDB := pkg.GetMongoDBInstance()
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	return unificationRepo.GetUnificationRules()
}
