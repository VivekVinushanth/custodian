package handlers

import (
	"custodian/internal/models"
	"custodian/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Your data Custodian is up and running"})
}

// GetProfile handles profile retrieval requests
func GetProfile(c *gin.Context) {
	permaID := c.Param("perma_id") // Extract `perma_id` from URL

	// Optional: Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	var profile *models.Profile
	var err error

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		profile, err = service.GetProfileWithToken(permaID, token)
	} else {
		profile, err = service.GetProfileWithToken(permaID, "")
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

// DeleteProfile handles profile retrieval requests
func DeleteProfile(c *gin.Context) {
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

// GetAllProfile handles profile retrieval requests
func GetAllProfile(c *gin.Context) {

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

// PutAppContextData handles inserting/updating app context inside a user's profile
func AddOrUpdateAppContext(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	var appContext models.AppContext
	if err := c.ShouldBindJSON(&appContext); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	appContext.AppID = appID

	err := service.AddOrUpdateAppContext(permaID, appContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App context updated successfully"})
}

// UpdateAppContextData handles PATCH updates for app context
func UpdateAppContextData(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.UpdateAppContextData(permaID, appID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App context updated successfully"})
}

// GetAppContextData handles retrieving app context from a user's profile
func GetAppContextData(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	appContext, err := service.GetAppContextData(permaID, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app context"})
		return
	}

	if appContext == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No app context found for this app"})
		return
	}

	c.JSON(http.StatusOK, appContext)
}

// GetListOfAppContextData fetches all app contexts for a user
func GetListOfAppContextData(c *gin.Context) {
	permaID := c.Param("perma_id")

	appContexts, err := service.GetListOfAppContextData(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app context list"})
		return
	}

	if appContexts == nil || len(appContexts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No app contexts found for this user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"app_contexts": appContexts})
}

// AddOrUpdatePersonalityData handles inserting/updating personality data inside a user's profile
func AddOrUpdatePersonalityData(c *gin.Context) {
	permaID := c.Param("perma_id")

	var personalityData map[string]interface{}
	if err := c.ShouldBindJSON(&personalityData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.AddOrUpdatePersonalityData(permaID, personalityData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update personality data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Personality data updated successfully"})
}

// UpdatePersonalityData handles PATCH updates for personality data
func UpdatePersonalityData(c *gin.Context) {
	permaID := c.Param("perma_id")

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := service.UpdatePersonalityData(permaID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update personality data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Personality data updated successfully"})
}

// GetPersonalityProfileData handles retrieving personality data from a user's profile
func GetPersonalityProfileData(c *gin.Context) {
	permaID := c.Param("perma_id")

	personalityData, err := service.GetPersonalityProfileData(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve personality data"})
		return
	}

	if personalityData == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No personality data found for this user"})
		return
	}

	c.JSON(http.StatusOK, personalityData)
}

// CreateTraits handles POST /unification_rules to create new profile enrichment rules
func CreateTraits(c *gin.Context) {
	var rules models.ProfileEnrichmentRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	err := service.AddProfileTrait(rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rules"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rules saved successfully"})
}

// GetTraits handles GET /unification_rules to retrieve all rules
func GetTraits(c *gin.Context) {
	rules, err := service.GetProfileTraits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func GetTrait(c *gin.Context) {
	rules, err := service.GetProfileTraits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func PutTrait(c *gin.Context) {
	var rules models.ProfileEnrichmentRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}
	err := service.AddProfileTrait(rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rules"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rules saved successfully"})
}

// DeleteTrait handles DELETE /unification_rules/:rule_name
func DeleteTrait(c *gin.Context) {
	ruleName := c.Param("trait_id")
	if ruleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_name is required"})
		return
	}

	err := service.DeleteProfileTrait(ruleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
