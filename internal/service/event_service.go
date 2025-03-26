package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// AddEvent stores a single event in MongoDB
func AddEvent(event models.Event) error {
	mongoDB := pkg.GetMongoDBInstance()

	// Step 1: Check if the profile exists
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	existingProfile, err := profileRepo.FindProfileByID(event.PermaID)
	if err != nil || existingProfile == nil {
		return fmt.Errorf("profile_not_found")
	}

	// Step 2: If category_searched, enrich personality.interests
	if strings.ToLower(event.EventType) == "track" && strings.ToLower(event.EventName) == "category_searched" {
		if props := event.Properties; props != nil {
			if raw, ok := props["track_event"]; ok {
				rawBytes, _ := json.Marshal(raw)
				var trackEvent models.TrackEvent
				if err := json.Unmarshal(rawBytes, &trackEvent); err == nil {
					if category, ok := extractCategoryFromTrackEvent(trackEvent); ok {

						// Step 2.1: Try appending first
						update := bson.M{
							"$addToSet": bson.M{
								"personality.interests": category,
							},
						}
						err := profileRepo.UpdatePreferenceData(event.PermaID, update)

						// Step 2.2: If interests is not an array, fix and retry
						if err != nil && strings.Contains(err.Error(), "Cannot apply $addToSet to non-array field") {
							// Fix field
							init := bson.M{
								"$set": bson.M{
									"personality.interests": []string{},
								},
							}
							_ = profileRepo.UpdatePreferenceData(event.PermaID, init)

							// Retry append
							err = profileRepo.UpdatePreferenceData(event.PermaID, update)
						}

						if err != nil {
							return fmt.Errorf("failed to enrich personality with category: %v", err)
						}
					}
				}
			}
		}
	}

	// Step 3: Store the event
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
	return eventRepo.AddEvent(event)
}

func extractCategoryFromTrackEvent(event models.TrackEvent) (string, bool) {
	if strings.ToLower(event.Action) == "select_category" && event.ObjectName != "" {
		return event.ObjectName, true
	}
	return "", false
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

func DecodeEventProperties(event *models.Event) {
	if event.Properties == nil {
		return
	}

	original := event.Properties // backup

	switch strings.ToLower(event.EventType) {
	case "track":
		if raw, ok := original["track_event"]; ok {
			rawBytes, _ := json.Marshal(raw)
			var track models.TrackEvent
			if err := json.Unmarshal(rawBytes, &track); err == nil && !isStructEmpty(track) {
				event.Properties = structToMap(track)
				return
			}
		}
	case "page":
		if raw, ok := original["page_event"]; ok {
			rawBytes, _ := json.Marshal(raw)
			var page models.PageEvent
			if err := json.Unmarshal(rawBytes, &page); err == nil && !isStructEmpty(page) {
				event.Properties = structToMap(page)
				return
			}
		}
	case "identify":
		if raw, ok := original["identify_event"]; ok {
			rawBytes, _ := json.Marshal(raw)
			var id models.IdentifyEvent
			if err := json.Unmarshal(rawBytes, &id); err == nil && !isStructEmpty(id) {
				event.Properties = structToMap(id)
				return
			}
		}
	}

	// fallback if decoding fails
	event.Properties = original
}

func structToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	bytes, _ := json.Marshal(v)
	_ = json.Unmarshal(bytes, &result)
	return result
}

func isStructEmpty(v interface{}) bool {
	bytes, _ := json.Marshal(v)
	return string(bytes) == "{}"
}
