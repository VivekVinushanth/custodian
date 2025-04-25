package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/service"
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
	permaID := c.Param("perma_id") // Extract `perma_id` from URL

	profile, err := service.DeleteProfile(permaID)
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

	profile, err := service.GetAllProfiles()
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

//// AddOrUpdateAppContext handles inserting/updating app context inside a user's profile
//func AddOrUpdateAppContext(c *gin.Context) {
//	permaID := c.Param("perma_id")
//	appID := c.Param("app_id")
//
//	var appContext models.ApplicationData
//	if err := c.ShouldBindJSON(&appContext); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	appContext.AppId = appID
//
//	err := service.AddOrUpdateAppContext(permaID, appContext)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app context"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "App context updated successfully"})
//}
//
//// UpdateAppContextData handles PATCH updates for app context
//func UpdateAppContextData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//	appID := c.Param("app_id")
//
//	var updates bson.M
//	if err := c.ShouldBindJSON(&updates); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	err := service.UpdateAppContextData(permaID, appID, updates)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app context"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "App context updated successfully"})
//}

// GetAppContextData handles retrieving app context from a user's profile
//func GetAppContextData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//	appID := c.Param("app_id")
//
//	appContext, err := service.GetAppContextData(permaID, appID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app context"})
//		return
//	}
//
//	if appContext == nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "No app context found for this app"})
//		return
//	}
//
//	c.JSON(http.StatusOK, appContext)
//}

// GetListOfAppContextData fetches all app contexts for a user
//func GetListOfAppContextData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//
//	appContexts, err := service.GetListOfAppContextData(permaID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app context list"})
//		return
//	}
//
//	if appContexts == nil || len(appContexts) == 0 {
//		c.JSON(http.StatusNotFound, gin.H{"error": "No app contexts found for this user"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"app_contexts": appContexts})
//}

//// AddOrUpdatePersonalityData handles inserting/updating personality data inside a user's profile
//func AddOrUpdatePersonalityData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//
//	var personalityData map[string]interface{}
//	if err := c.ShouldBindJSON(&personalityData); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	err := service.AddOrUpdatePersonalityData(permaID, personalityData)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update personality data"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Personality data updated successfully"})
//}
//
//// UpdatePersonalityData handles PATCH updates for personality data
//func UpdatePersonalityData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//
//	var updates bson.M
//	if err := c.ShouldBindJSON(&updates); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	err := service.UpdatePersonalityData(permaID, updates)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update personality data"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Personality data updated successfully"})
//}
//
//// GetPersonalityProfileData handles retrieving personality data from a user's profile
//func GetPersonalityProfileData(c *gin.Context) {
//	permaID := c.Param("perma_id")
//
//	personalityData, err := service.GetPersonalityProfileData(permaID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve personality data"})
//		return
//	}
//
//	if personalityData == nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "No personality data found for this user"})
//		return
//	}
//
//	c.JSON(http.StatusOK, personalityData)
//}

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

// GetEnrichmentRules handles GET /enrichment-rules to retrieve all rules
func (s Server) GetEnrichmentRules(c *gin.Context) {
	rules, err := service.GetEnrichmentRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
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
