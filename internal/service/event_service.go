package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"fmt"
)

// AddEvent stores a single event in MongoDB
func AddEvent(event models.Event) error {
	mongoDB := pkg.GetMongoDBInstance()
	// Check if the profile exists for the perma_id
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	existingProfile, err := profileRepo.FindProfileByID(event.PermaID)
	if err != nil || existingProfile == nil {
		return fmt.Errorf("profile_not_found")
	}
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.AddEvent(event)
}

// AddEvents stores multiple events in MongoDB
func AddEvents(events []models.Event) error {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.AddEvents(events)
}

// GetUserEvent retrieves a single event
func GetUserEvent(permaID, eventID string) (*models.Event, error) {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.GetUserEvent(permaID, eventID)
}

// GetUserEvents retrieves all events for a user
func GetUserEvents(permaID string) ([]models.Event, error) {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.GetUserEvents(permaID)
}

// DeleteEvent removes a single event by perma_id and event_id
func DeleteEvent(permaID, eventID string) error {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.DeleteEvent(permaID, eventID)
}

// DeleteEventsByPermaID removes all events for a specific user (perma_id)
func DeleteEventsByPermaID(permaID string) error {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.DeleteEventsByPermaID(permaID)
}

// DeleteEventsByAppID removes all events for a given app tied to a user
func DeleteEventsByAppID(permaID, appID string) error {
	mongoDB := pkg.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.DeleteEventsByAppID(permaID, appID)
}
