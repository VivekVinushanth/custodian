package handlers

import "github.com/gin-gonic/gin"

// RegisterRoutes initializes API routes
func RegisterRoutes(router *gin.Engine) {
	base := router.Group("/api/v1") // Base URL defined here
	{
		// Profile APIs
		profile := base.Group("/profile")
		{
			profile.POST("/", CreateProfile)
		}

		base.GET("/:perma_id/profile", GetProfile)

		// App Context APIs
		appContext := base.Group("/:perma_id/profile")
		{
			appContext.PUT("/:app_id/app_context", AddOrUpdateAppContext)
			appContext.PATCH("/:app_id/app_context", UpdateAppContextData)
			appContext.GET("/:app_id/app_context", GetAppContextData)
			appContext.GET("/app_context", GetListOfAppContextData)
		}

		// Personality Data APIs
		personality := base.Group("/:perma_id/profile/personality")
		{
			personality.PUT("/", AddOrUpdatePersonalityData)
			personality.PATCH("/", UpdatePersonalityData)
			personality.GET("/", GetPersonalityProfileData)
		}

		// Alias APIs
		//alias := base.Group("/:perma_id/alias")
		//{
		//	alias.POST("/", AliasUser)
		//	alias.GET("/", GetAlias)
		//}

		// Events APIs
		events := base.Group("/:perma_id")
		{
			events.POST("/event", AddEvent)
			events.POST("/events", AddEvents)
			events.GET("/events/:event_id", GetUserEvent)
			events.GET("/events", GetUserEvents)
		}

		// Consent APIs (Newly Added)
		consent := base.Group("/consents/:perma_id")
		{
			consent.POST("/:app_id/collect/", GiveConsentToCollect)
			consent.POST("/:app_id/share/", GiveConsentToShare)
			consent.GET("/", GetConsentedApps)
			consent.GET("/:app_id/collect", GetConsentedAppsToCollect)
			consent.GET("/:app_id/share", GetConsentedAppsToShare)
			consent.DELETE("/", RevokeAllConsents)
			consent.DELETE("/:app_id/collect/", RevokeConsentToCollect)
			consent.DELETE("/:app_id/:share", RevokeConsentToShare)
		}

	}
}
