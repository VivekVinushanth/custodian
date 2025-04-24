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

		base.GET("/profile/:perma_id", GetProfile)
		base.DELETE("/profile/:perma_id", DeleteProfile)

		// Personality Data APIs
		personality := base.Group("/profile/:perma_id/personality")
		{
			personality.GET("/", GetPersonalityProfileData)
		}

		base.POST("/event", AddEvent)
		base.POST("/events", AddEvents)

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

		unification := base.Group("/unification-rules")
		{
			unification.POST("/", AddResolutionRule)
			unification.GET("/", GetResolutionRules)
			unification.GET("/:rule_id", GetResolutionRule)
			unification.PATCH("/:rule_id", PatchResolutionRule)
			unification.DELETE("/:rule_id", DeleteResolutionRule)
		}

		enrichment := base.Group("/enrichment-rules")
		{
			enrichment.POST("", CreateProfileEnrichmentRules)
			enrichment.GET("/", GetProfileEnrichmentRules)
			enrichment.GET("/:rule_id", GetProfileEnrichmentRule)
			enrichment.PUT("/:rule_id", PutGetProfileEnrichmentRule)
			enrichment.DELETE("/:rule_id", DeleteGetProfileEnrichmentRule)
		}

	}
}
