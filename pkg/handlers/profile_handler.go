package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"github.com/wso2/identity-customer-data-service/pkg/utils"
	"log"
	"net/http"
	"strings"
)

// GetProfile handles profile retrieval requests
func (s Server) GetProfile(c *gin.Context, profileId string) {

	// Optional: Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	var profile *models.Profile
	var err error

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		profile, err = service.GetProfileWithToken(profileId, token)
	} else {
		profile, err = service.GetProfileWithToken(profileId, "")
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving profile"})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile) // Return profile in JSON format
}

// GetTraits handles profile retrieval requests
func (s Server) GetTraits(c *gin.Context, profileId string) {

	// Optional: Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	var profile *models.Profile
	var err error

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		profile, err = service.GetProfileWithToken(profileId, token)
	} else {
		profile, err = service.GetProfileWithToken(profileId, "")
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving profile"})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	traits := profile.Traits

	c.JSON(http.StatusOK, traits) // Return traits in JSON format
}

// DeleteProfile handles profile retrieval requests
func (s Server) DeleteProfile(c *gin.Context, profileId string) {

	profile, err := service.DeleteProfile(profileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving profile"})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusNoContent, profile) // Return profile in JSON format
}

// GetAllProfiles handles profile retrieval requests
func (s Server) GetAllProfiles(c *gin.Context) {
	var profiles []models.Profile
	var err error

	// Build the filter from query params
	filter := c.QueryArray("filter") // Handles multiple `filter=...` parameters

	if len(filter) > 0 {
		log.Print("Filters: ", filter)
		profiles, err = service.GetAllProfilesWithFilter(filter)
	} else {
		profiles, err = service.GetAllProfiles() // pass empty filter
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profiles"})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// CreateEnrichmentRule handles POST /enrichment-rules to create new profile enrichment rules
func (s Server) CreateEnrichmentRule(c *gin.Context) {
	var rules models.ProfileEnrichmentRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	err := service.AddEnrichmentRule(rules)
	fmt.Printf("AddEnrichmentRule called with rule: %+v\n", rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rules", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rules saved successfully"})
}

// GetEnrichmentRules handles GET /enrichment-rules to retrieve all rules or with filters
func (s Server) GetEnrichmentRules(c *gin.Context) {

	filters := c.QueryArray("filter") // Handles multiple `filter=...` parameters

	if len(filters) > 0 {
		log.Print("Filters: ", filters)
		rules, err := service.GetEnrichmentRulesByFilter(filters)
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, rules)
		return
	}

	// fallback: all rules
	rules, err := service.GetEnrichmentRules()
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rules)
}

// GetEnrichmentRule handles GET /enrichment-rules/:rule_id to retrieve a specific rule
func (s Server) GetEnrichmentRule(c *gin.Context, ruleId string) {
	rule, err := service.GetEnrichmentRule(ruleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rule"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (s Server) PutEnrichmentRule(c *gin.Context, ruleId string) {
	//TODO update the implementation
	var rules models.ProfileEnrichmentRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}
	err := service.AddEnrichmentRule(rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rules"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rules saved successfully"})
}

// DeleteEnrichmentRule handles DELETE /unification_rules/:rule_name
func (s Server) DeleteEnrichmentRule(c *gin.Context, ruleId string) {
	if ruleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_name is required"})
		return
	}

	err := service.DeleteEnrichmentRule(ruleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
