package handlers

import (
	"custodian/internal/models"
	"custodian/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddEvent handles adding a single event
func AddEvent(c *gin.Context) {

	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err := service.AddEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event added successfully"})
}

// AddEvents handles batch event creation
func AddEvents(c *gin.Context) {
	permaID := c.Param("perma_id")

	var events []models.Event
	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	for i := range events {
		events[i].PermaID = permaID
	}

	err := service.AddEvents(events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add events"})
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

	c.JSON(http.StatusOK, events)
}
