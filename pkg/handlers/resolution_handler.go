package handlers

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"identity-customer-data-service/pkg/errors"
	"identity-customer-data-service/pkg/models"
	"identity-customer-data-service/pkg/service"
	"identity-customer-data-service/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddResolutionRule handles adding a new rule
func AddResolutionRule(c *gin.Context) {

	var rule models.UnificationRule
	if rule.RuleId == "" {
		rule.RuleId = uuid.NewString()
	}
	if err := c.ShouldBindJSON(&rule); err != nil {
		badReq := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrBadRequest.Code,
			Message:     errors.ErrBadRequest.Message,
			Description: err.Error(),
		}, http.StatusBadRequest)

		utils.HandleError(c, badReq)
		return
	}
	err := service.AddResolutionRule(rule)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule) // Return the created rule in JSON format
}

// GetResolutionRules handles adding a new rule
func GetResolutionRules(c *gin.Context) {

	rules, err := service.GetResolutionRules()
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rules)
}

// GetResolutionRule Fetches a specific resolution rule.
func GetResolutionRule(c *gin.Context) {

	ruleId := c.Param("rule_id")
	rule, err := service.GetResolutionRule(ruleId)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// PatchResolutionRule applies partial updates to a resolution rule.
func PatchResolutionRule(c *gin.Context) {
	ruleId := c.Param("rule_id")

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		badReq := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrBadRequest.Code,
			Message:     errors.ErrBadRequest.Message,
			Description: err.Error(),
		}, http.StatusBadRequest)

		utils.HandleError(c, badReq)
		return
	}

	err := service.PatchResolutionRule(ruleId, updates)
	if err != nil {
		utils.HandleError(c, err)
	}

	rule, err := service.GetResolutionRule(ruleId)

	c.JSON(http.StatusOK, rule)
}

// DeleteResolutionRule removes a resolution rule.
func DeleteResolutionRule(c *gin.Context) {
	ruleId := c.Param("rule_id")
	err := service.DeleteResolutionRule(ruleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unification rule"})
		//logger.Log.Error("Error happened when updating resolution rule. " + err.Error())
		return
	}

	c.JSON(http.StatusNoContent, "Unification rule deleted successfully")
}
