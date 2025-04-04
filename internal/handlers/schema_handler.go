package handlers

import (
	"custodian/internal/models"
	"custodian/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateSchemaRules handles POST /unification_rules to create new profile enrichment rules
func CreateSchemaRules(c *gin.Context) {
	var rules []models.ProfileEnrichmentRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	err := service.CreateSchemaRules(rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rules"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rules saved successfully"})
}

// GetSchemaRules handles GET /unification_rules to retrieve all rules
func GetSchemaRules(c *gin.Context) {
	rules, err := service.GetSchemaRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// DeleteSchemaRule handles DELETE /unification_rules/:rule_name
func DeleteSchemaRule(c *gin.Context) {
	ruleName := c.Param("rule_name")
	if ruleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_name is required"})
		return
	}

	err := service.DeleteSchemaRule(ruleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
