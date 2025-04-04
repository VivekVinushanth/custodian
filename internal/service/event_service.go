package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
	"time"
)

// AddEvent stores a single event in MongoDB
func AddEvent(event models.Event) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	if event.PermaID == "" {
		return fmt.Errorf("perma_id not found")
	}

	if _, err := CreateOrUpdateProfile(event); err != nil {
		return fmt.Errorf("failed to create new profile: %v", err)
	}

	// 🔸 Store the event
	profile, _ := profileRepo.FindProfileByID(event.PermaID)
	if profile != nil {
		eventRepo := repositories.NewEventRepository(mongoDB.Database, "events")
		return eventRepo.AddEvent(event)
	}
	return fmt.Errorf("profile not found")
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

// EnrichProfile updates interests list based on events
func EnrichProfile(event models.Event) error {
	profileRepo := repositories.NewProfileRepository(pkg.GetMongoDBInstance().Database, "profiles")

	profile, _ := waitForProfile(event.PermaID, 5, 100*time.Millisecond)

	if profile == nil {
		return fmt.Errorf("profile not found to enrich")
	}

	// todo: Enrich happens only for parent profile, not for child profiles. Only if it is session data enrichment happens for child profile
	// 🔹 Enrich app_context.devices if event.Context has device_id
	if event.Context != nil {
		if raw, ok := event.Context["device_id"]; ok {
			if deviceID, ok := raw.(string); ok && deviceID != "" {
				device := models.AppContextDevices{
					DeviceID: deviceID,
					LastUsed: event.EventTimestamp, // format to string
				}

				// Optional enrichment fields
				if os, ok := event.Context["os"].(string); ok {
					device.Os = os
				}
				if browser, ok := event.Context["browser"].(string); ok {
					device.Browser = browser
				}
				if version, ok := event.Context["browser_version"].(string); ok {
					device.BrowserVersion = version
				}
				if ip, ok := event.Context["ip"].(string); ok {
					device.Ip = ip
				}
				if deviceType, ok := event.Context["device_type"].(string); ok {
					device.DeviceType = deviceType
				}

				permaId := event.PermaID

				// Enriching only the master profile
				//todo: Child is enriched only for session information
				if !profile.ProfileHierarchy.IsMaster {
					permaId = profile.ProfileHierarchy.ParentProfileID
				}
				appContext := models.AppContext{
					AppID:   event.AppID,
					Devices: []models.AppContextDevices{device},
				}
				// 🔁 Update app_context
				if err := profileRepo.AddOrUpdateAppContext(permaId, appContext); err != nil {
					return fmt.Errorf("failed to enrich app context: %v", err)
				}

			}
		}
	}

	// 🔹 Enrich personality.interests if category_searched
	if strings.ToLower(event.EventType) == "track" && strings.ToLower(event.EventName) == "category_searched" {
		permaId := event.PermaID
		if profile.ProfileHierarchy != nil && !profile.ProfileHierarchy.IsMaster {
			permaId = profile.ProfileHierarchy.ParentProfileID
		}

		log.Print("Enriching interests for profile: ", event.Properties)
		action := event.Properties["action"]
		if action == "select_category" {
			if category, ok := event.Properties["objectname"]; ok {
				update := bson.M{
					"$addToSet": bson.M{
						"personality.interests": category,
					},
				}
				err := profileRepo.UpdatePreferenceData(permaId, update)

				if err != nil && strings.Contains(err.Error(), "Cannot create field 'interests' in element {personality: null}") {
					init := bson.M{
						"$set": bson.M{
							"personality": bson.M{"interests": []string{}},
						},
					}
					_ = profileRepo.UpdatePreferenceData(permaId, init)
					err = profileRepo.UpdatePreferenceData(permaId, update)
				}

				if err != nil {
					return fmt.Errorf("failed to enrich interests: %v", err)
				}
			}
		}
	}

	log.Println("Event Cat========== ", event.EventType)
	log.Println("Event Name========== ", event.EventName)
	log.Println("Event Props========== ", event.Properties)

	// 🔹 Enrich identity data if user_logged_in event
	if strings.ToLower(event.EventType) == "identify" && strings.ToLower(event.EventName) == "user_logged_in" {
		log.Println("Enriching identity data for profile========== ", event.Properties)
		permaID := event.PermaID
		if profile.ProfileHierarchy != nil && !profile.ProfileHierarchy.IsMaster {
			permaID = profile.ProfileHierarchy.ParentProfileID
		}

		identityData := models.IdentityData{}
		if val, ok := event.Properties["username"].(string); ok {
			identityData.Username = val
		}
		if val, ok := event.Properties["email"].(string); ok {
			identityData.Email = val
		}
		if val, ok := event.Properties["first_name"].(string); ok {
			identityData.FirstName = val
		}
		if val, ok := event.Properties["last_name"].(string); ok {
			identityData.LastName = val
		}
		if val, ok := event.Properties["user_id"].(string); ok {
			identityData.UserId = val
		}

		if err := profileRepo.AddOrUpdateIdentityData(permaID, identityData); err != nil {
			return fmt.Errorf("failed to enrich identity data: %v", err)
		}

	}

	return nil
}
