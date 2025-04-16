package handlers

import "github.com/gin-gonic/gin"

// RegisterRoutes initializes API routes
func RegisterRoutes(router *gin.Engine) {
	base := router.Group("/api/v1") // Base URL defined here
	{
		// Profile APIs
		profile := base.Group("/profile")
		{
			profile.GET("/", GetAllProfile)
		}

		base.GET("/:perma_id/profile", GetProfile)
		base.DELETE("/:perma_id/profile", DeleteProfile)

		// Personality Data APIs
		personality := base.Group("/:perma_id/profile/personality")
		{
			personality.PUT("/", AddOrUpdatePersonalityData)
			personality.PATCH("/", UpdatePersonalityData)
			personality.GET("/", GetPersonalityProfileData)
		}

		base.POST("/event", AddEvent)
		base.POST("/events", AddEvents)
		base.GET("/app/:app_id/events", GetEvents)

		// Events APIs
		events := base.Group("/:perma_id")
		{
			events.GET("/events/:event_id", GetUserEvent)
			events.GET("/events", GetUserEvents)
		}

		// Event Schema
		eventSchema := base.Group("/event-schema")
		{
			eventSchema.POST("", AddEventSchema)
			eventSchema.GET("/", GetEventSchemas)
			eventSchema.GET("/{event_schema_id}", GetEventSchema)
			eventSchema.PATCH("/{event_schema_id}", PatchEventSchema)
			eventSchema.DELETE("/{event_schema_id}", DeleteEventSchema)
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

		resolution := base.Group("/resolution-rules")
		{
			resolution.POST("/", AddResolutionRule)
			resolution.GET("/", GetResolutionRules)
			resolution.GET("/:rule_id", GetResolutionRule)
			resolution.PATCH("/:rule_id", PatchResolutionRule)
			resolution.DELETE("/:rule_id", DeleteResolutionRule)
		}

		traits := base.Group("/profile-traits")
		{
			traits.POST("", CreateTraits)
			traits.GET("/", GetTraits)
			traits.GET("/:trait_id", GetTrait)
			traits.PUT("/:trait_id", PutTrait)
			traits.DELETE("/:trait_id", DeleteTrait)
		}

	}
}
