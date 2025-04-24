package handlers

import (
	"identity-customer-data-service/pkg/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GiveConsentToCollect handles consent requests for data collection
func GiveConsentToCollect(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	err := service.GiveConsentToCollect(permaID, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to give consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent given for data collection"})
}

// GiveConsentToShare handles consent requests for data collection
func GiveConsentToShare(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	err := service.GiveConsentToShare(permaID, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to give consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent given for data collection"})
}

// GetConsentedApps handles fetching consented apps
func GetConsentedApps(c *gin.Context) {
	permaID := c.Param("perma_id")
	consents, err := service.GetConsentedApps(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consents"})
		return
	}
	c.JSON(http.StatusOK, consents)
}

// GetConsentedAppsToCollect handles HTTP requests to fetch apps where the user has given consent to collect data
func GetConsentedAppsToCollect(c *gin.Context) {
	permaID := c.Param("perma_id") // Extract perma_id from URL

	// Call the service layer
	apps, err := service.GetConsentedAppsToCollect(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consented_apps": apps})
}

// GetConsentedAppsToShare handles HTTP requests to fetch apps where the user has given consent to collect data
func GetConsentedAppsToShare(c *gin.Context) {
	permaID := c.Param("perma_id") // Extract perma_id from URL

	// Call the service layer
	apps, err := service.GetConsentedAppsToShare(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consented_apps": apps})
}

// RevokeConsentToCollect handles consent revocation
func RevokeConsentToCollect(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	err := service.RevokeConsentToCollect(permaID, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent revoked for data collection"})
}

// RevokeConsentToShare handles consent revocation
func RevokeConsentToShare(c *gin.Context) {
	permaID := c.Param("perma_id")
	appID := c.Param("app_id")

	err := service.RevokeConsentToShare(permaID, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent revoked for data collection"})
}

// RevokeAllConsents handles consent revocation
func RevokeAllConsents(c *gin.Context) {
	permaID := c.Param("perma_id")

	err := service.RevokeAllConsents(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent revoked for data collection"})
}
