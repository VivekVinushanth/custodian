package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"strconv"
	"time"
)

// AddEvent handles adding a single event
func (s Server) AddEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := service.AddEvents(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event added successfully"})
}

// GetUserEvent fetches a specific event
func (s Server) GetEvent(c *gin.Context, eventId string) {
	//TODO implement me
	panic("implement me")
}

// TODO remove
func (s Server) GetEvents(c *gin.Context) {
	// Step 1: Extract raw filters (e.g., event_type+eq+Identify)
	rawFilters := c.QueryArray("filter")
	log.Println("Filters11: ", rawFilters)

	// Step 2: Parse optional time range
	var timeFilter bson.M
	if timeStr := c.Query("time_range"); timeStr != "" {
		log.Println("Time Rangedfff: ", timeStr)
		durationSec, _ := strconv.Atoi(timeStr)       // parse string to int
		currentTime := time.Now().UTC().Unix()        // current time in seconds
		startTime := currentTime - int64(durationSec) // assuming value is in minutes
		timeFilter = bson.M{
			"event_timestamp": bson.M{"$gte": startTime},
		}
	}

	// Step 3: Fetch events with filter strings
	events, err := service.GetEvents(rawFilters, timeFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}
