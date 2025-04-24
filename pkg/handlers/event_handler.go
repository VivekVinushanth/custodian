package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"identity-customer-data-service/pkg/models"
	"identity-customer-data-service/pkg/service"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AddEvent handles adding a single event
func AddEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
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

	// ✅ Do not alter or enforce structure — pass as-is to the service
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

func GetEvents(c *gin.Context) {
	// Optional filters
	appId := c.Param("app_id")
	eventType := c.Query("event_type")
	eventName := c.Query("event_name")
	timeRange := c.Query("time_range") // in minutes
	//from := c.Query("from") // epoch is expected
	//to := c.Query("to")
	// todo: validate date format for from-to

	// Start building the filter
	filter := bson.M{
		"app_id": appId,
	}

	if eventType != "" {
		filter["event_type"] = strings.ToLower(eventType)
	}
	if eventName != "" {
		filter["event_name"] = strings.ToLower(eventName)
	}

	// Handle timestamp range
	//timeFilter := bson.M{}
	//if from != "" {
	//	fromTime := parseTimeToISO(from)
	//	if !fromTime.IsZero() {
	//		timeFilter["$gte"] = fromTime
	//	}
	//}
	//if to != "" {
	//	toTime := parseTimeToISO(to)
	//	if !toTime.IsZero() {
	//		timeFilter["$lte"] = toTime
	//	}
	//}
	//if len(timeFilter) > 0 {
	//	filter["event_timestamp"] = timeFilter
	//}

	// 🔸 Apply time_range as a filter on event_timestamp (Unix)
	if timeRange != "" {
		durationSec, err := strconv.Atoi(timeRange) // parse string to int
		if err != nil {
			log.Printf("Invalid time range format: %v", err)
			return
		}

		currentTime := time.Now().UTC().Unix()        // current time in seconds
		startTime := currentTime - int64(durationSec) // assuming value is in minutes

		filter["event_timestamp"] = bson.M{"$gte": startTime}
	}

	// Parse custom property filters
	for key, value := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "properties.") {
			log.Print("Found custom property filter: ", key, ":", value)
			filter[key] = value[0]
		}
	}
	events, err := service.GetEvents(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}
