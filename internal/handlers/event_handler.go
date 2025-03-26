package handlers

import (
	"custodian/internal/models"
	"custodian/internal/service"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AddEvent handles adding a single event
func AddEvent(c *gin.Context) {

	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	switch strings.ToLower(event.EventType) {
	case "track":
		var trackEvent models.TrackEvent
		propBytes, _ := json.Marshal(event.Properties)
		if err := json.Unmarshal(propBytes, &trackEvent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track event properties"})
			return
		}
		event.Properties = map[string]interface{}{"track_event": trackEvent}

	case "page":
		var pageEvent models.PageEvent
		propBytes, _ := json.Marshal(event.Properties)
		if err := json.Unmarshal(propBytes, &pageEvent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page event properties"})
			return
		}
		event.Properties = map[string]interface{}{"page_event": pageEvent}

	case "identify":
		var identifyEvent models.IdentifyEvent
		propBytes, _ := json.Marshal(event.Properties)
		if err := json.Unmarshal(propBytes, &identifyEvent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid identify event properties"})
			return
		}
		event.Properties = map[string]interface{}{"identify_event": identifyEvent}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		return
	}

	if err := service.AddEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event added successfully"})
}

// AddEvents handles adding multiple events
func AddEvents(c *gin.Context) {
	var events []models.Event

	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Decode and structure each event's properties correctly
	for i, event := range events {
		switch strings.ToLower(event.EventType) {
		case "track":
			var trackEvent models.TrackEvent
			propBytes, _ := json.Marshal(event.Properties)
			if err := json.Unmarshal(propBytes, &trackEvent); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track event properties"})
				return
			}
			events[i].Properties = map[string]interface{}{"track_event": trackEvent}

		case "page":
			var pageEvent models.PageEvent
			propBytes, _ := json.Marshal(event.Properties)
			if err := json.Unmarshal(propBytes, &pageEvent); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page event properties"})
				return
			}
			events[i].Properties = map[string]interface{}{"page_event": pageEvent}

		case "identify":
			var identifyEvent models.IdentifyEvent
			propBytes, _ := json.Marshal(event.Properties)
			if err := json.Unmarshal(propBytes, &identifyEvent); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid identify event properties"})
				return
			}
			events[i].Properties = map[string]interface{}{"identify_event": identifyEvent}

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type in batch"})
			return
		}
	}

	if err := service.AddEvents(events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Events added successfully"})
}

// GetUserEvent fetches a specific event
func GetUserEvent(c *gin.Context) {
	permaID := c.Param("perma_id")
	eventID := c.Param("event_id")

	event, err := service.GetUserEvent(permaID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Re-decode properties based on type
	service.DecodeEventProperties(event)

	c.JSON(http.StatusOK, event)
}

// GetUserEvents fetches all events for a user
func GetUserEvents(c *gin.Context) {
	permaID := c.Param("perma_id")

	events, err := service.GetUserEvents(permaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	// Decode all event properties
	for i := range events {
		service.DecodeEventProperties(&events[i])
	}

	c.JSON(http.StatusOK, events)
}
