package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
)

// AddEvent stores a single event in MongoDB
func AddEvent(event models.Event) error {
	mongoDB := pkg.GetMongoDBInstance()
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
