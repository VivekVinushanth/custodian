package handlers

import (
	"custodian/internal/models"
	"custodian/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateUnificationRule handles adding a new rule
func CreateUnificationRule(c *gin.Context) {
	var rule models.UnificationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.CreateUnificationRule(rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rule created successfully"})
}

// GetUnificationRules handles adding a new rule
func GetUnificationRules(c *gin.Context) {

	rules, err := service.GetUnificationRules()
	if err != nil {
		if rules == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
	}

	c.JSON(http.StatusOK, rules) // Return profile in JSON format
}

// UpdateUnificationRule handles modifying an existing rule
func UpdateUnificationRule(c *gin.Context) {
	ruleName := c.Param("rule_name")

	var updatedRule models.UnificationRule
	if err := c.ShouldBindJSON(&updatedRule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.UpdateUnificationRule(ruleName, updatedRule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule updated successfully"})
}

// PatchUnificationRule applies partial updates to a unification rule
func PatchUnificationRule(c *gin.Context) {
	ruleName := c.Param("rule_name")

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.PatchUnificationRule(ruleName, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch unification rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unification rule patched successfully"})
}

// DeleteUnificationRule removes a unification rule from the database
func DeleteUnificationRule(c *gin.Context) {
	ruleName := c.Param("rule_name")

	err := service.DeleteUnificationRule(ruleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unification rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unification rule deleted successfully"})
}
