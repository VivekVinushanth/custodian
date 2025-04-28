package service

import (
	"fmt"
	"github.com/wso2/identity-customer-data-service/pkg/constants"
	"github.com/wso2/identity-customer-data-service/pkg/locks"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	repositories "github.com/wso2/identity-customer-data-service/pkg/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strconv"
	"strings"
	"time"
)

// AddEvents stores a single event in MongoDB
func AddEvents(event models.Event) error {

	// Step 1: Ensure profile exists (with lock protection)
	_, err := CreateOrUpdateProfile(event)
	if err != nil {
		return fmt.Errorf("failed to create or fetch profile: %v", err)
	}

	// Step 2: Store the event
	mongoDB := locks.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, constants.EventCollection)
	event.EventType = strings.ToLower(event.EventType)
	event.EventName = strings.ToLower(event.EventName)
	if err := eventRepo.AddEvent(event); err != nil {

		return fmt.Errorf("failed to store event: %v", err)
	}

	// Step 3: Enqueue the event for enrichment/unification (async)
	EnqueueEventForProcessing(event)

	return nil
}

// GetEvents retrieves all events
func GetEvents(filters []string, timeFilter bson.M) ([]models.Event, error) {
	mongoDB := locks.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, constants.EventCollection)
	return eventRepo.FindEvents(filters, timeFilter)
}

func GetEvent(eventId string) (*models.Event, error) {
	mongoDB := locks.GetMongoDBInstance()
	eventRepo := repositories.NewEventRepository(mongoDB.Database, constants.EventCollection)
	return eventRepo.FindEvent(eventId)
}

// EnrichProfile updates interests list based on events
func EnrichProfile(event models.Event) error {
	profileRepo := repositories.NewProfileRepository(locks.GetMongoDBInstance().Database, "profiles")

	profile, _ := waitForProfile(event.ProfileId, 5, 100*time.Millisecond)

	if profile == nil {
		return fmt.Errorf("profile not found to enrich")
	}

	// todo: Enrich happens only for parent profile, not for child profiles. Only if it is session data enrichment happens for child profile
	// 🔹 Enrich app_context.devices if event.Context has device_id
	if event.Context != nil {
		if raw, ok := event.Context["device_id"]; ok {
			if deviceID, ok := raw.(string); ok && deviceID != "" {
				devices := models.Devices{
					DeviceId: deviceID,
					LastUsed: event.EventTimestamp, // format to string
				}

				// Optional enrichment fields
				if os, ok := event.Context["os"].(string); ok {
					devices.Os = os
				}
				if browser, ok := event.Context["browser"].(string); ok {
					devices.Browser = browser
				}
				if version, ok := event.Context["browser_version"].(string); ok {
					devices.BrowserVersion = version
				}
				if ip, ok := event.Context["ip"].(string); ok {
					devices.Ip = ip
				}
				if deviceType, ok := event.Context["device_type"].(string); ok {
					devices.DeviceType = deviceType
				}

				profileId := event.ProfileId

				// Enriching only the master profile
				//todo: Enrich only the permanent profile
				if !profile.ProfileHierarchy.IsParent {
					profileId = profile.ProfileHierarchy.ParentProfileID
				}
				appContext := models.ApplicationData{
					AppId:   event.AppId,
					Devices: []models.Devices{devices},
				}
				// 🔁 Update app_context
				if err := profileRepo.AddOrUpdateAppContext(profileId, appContext); err != nil {
					return fmt.Errorf("failed to enrich app context: %v", err)
				}

			}
		}
	}

	// 🔹 Enrich personality.interests if event_name is category_searched

	schemaRepo := repositories.NewProfileSchemaRepository(locks.GetMongoDBInstance().Database, constants.ProfileSchemaCollection)
	rules, _ := schemaRepo.GetSchemaRules()
	for _, rule := range rules {
		if strings.ToLower(rule.Trigger.EventType) != strings.ToLower(event.EventType) ||
			strings.ToLower(rule.Trigger.EventName) != strings.ToLower(event.EventName) {
			continue
		}

		// Step 2: Evaluate conditions
		if !EvaluateConditions(event, rule.Trigger.Conditions) {
			continue
		}

		// Step 3: Get value to assign
		var value interface{}
		if rule.RuleType == "static" {
			value = rule.Value
		} else if rule.RuleType == "computed" {
			// Basic "copy" computation
			switch strings.ToLower(rule.Computation) {
			case "copy":
				if len(rule.SourceFields) != 1 {
					log.Printf("Invalid SourceFields for 'copy' computation. Expected 1, got: %d", len(rule.SourceFields))
					continue
				}
				value = GetFieldFromEvent(event, rule.SourceFields[0])
				log.Printf("Copying value from event: =====v", value)
			case "concat":
				if rule.SourceFields != nil && len(rule.SourceFields) >= 2 {
					var parts []string
					for _, field := range rule.SourceFields {
						fieldVal := GetFieldFromEvent(event, field)
						if fieldVal != nil {
							parts = append(parts, fmt.Sprintf("%v", fieldVal))
						}
					}
					if len(parts) > 0 {
						value = strings.Join(parts, "") // You can use a separator if needed
					}
				}
			case "count":
				// You'd call a service or repo to count events based on:
				//   - eventType + eventName
				//   - rule.Trigger.Conditions
				//   - rule.TimeRange (e.g., 7d)
				count, err := CountEventsMatchingRule(profile.ProfileId, rule.Trigger, rule.TimeRange)
				if err != nil {
					log.Printf("Failed to compute count for rule %s: %v", rule.RuleId, err)
					continue
				}
				value = count
			default:
				log.Printf("Unsupported computation: %s", rule.Computation)
				continue
			}
			// todo: Shall we support increment also here?
			// todo: Other aggergators might be added in future
			// todo: Other aggergators might be added in future
			// todo: Add more complex computation logic if needed
		}

		if value == nil {
			continue // skip if value couldn't be extracted
		}

		// Step 4: Apply masking if needed
		//if rule.MaskingRequired {
		//todo: see if we need to support this at this point - can be moved when fetching the data but not in persist
		//	if strVal, ok := value.(string); ok {
		//		value = utils.ApplyMasking(strVal, rule.MaskingStrategy)
		//	}
		//}

		//// Step 5: Apply merge strategy (existing value + new value)
		existingValue := GetNestedTraitValue(*profile, rule.TraitName) // e.g., identity.preferences or session.last_search
		value = MergeTraitValue(existingValue, value, rule.MergeStrategy, rule.ValueType)

		// Step 6: Apply merge strategy (existing value + new value)
		traitPath := strings.Split(rule.TraitName, ".")
		if len(traitPath) == 0 {
			log.Printf("Invalid trait path: %s", rule.TraitName)
			continue
		}

		namespace := traitPath[0] // e.g., identity
		traitName := traitPath[1] // e.g., email
		fieldPath := fmt.Sprintf("%s.%s", namespace, traitName)
		update := bson.M{fieldPath: value}
		switch namespace {
		case "traits":
			err := profileRepo.UpsertTraits(profile.ProfileId, update)
			if err != nil {
				log.Println("Error updating personality data:", err)
			}
		case "identity_attributes":
			log.Println("Updating identity data:", update)
			err := profileRepo.UpsertIdentityAttributes(profile.ProfileId, update)
			if err != nil {
				log.Println("Error updating identity data:", err)
			}
			continue
		case "application_data":
			continue
			//err = profileRepo.AddOrUpdateAppContext(permaID, traitPath[1:], value, rule.MergeStrategy, rule.ValueType)
		default:
			log.Printf("Unsupported trait namespace: %s", namespace)
			continue
		}
	}

	//// 🔹 Enrich identity data if user_logged_in event
	if strings.ToLower(event.EventType) == "identify" {

		permaID := event.ProfileId
		if profile.ProfileHierarchy != nil && !profile.ProfileHierarchy.IsParent {
			permaID = profile.ProfileHierarchy.ParentProfileID
		}

		identityData := make(map[string]interface{})
		if email, ok := event.Properties["email"].(string); ok && email != "" {
			identityData["email"] = email
		}
		if username, ok := event.Properties["user_name"].(string); ok && username != "" {
			identityData["user_name"] = username
		}
		if username, ok := event.Properties["first_name"].(string); ok && username != "" {
			identityData["first_name"] = username
		}
		if username, ok := event.Properties["last_name"].(string); ok && username != "" {
			identityData["last_name"] = username
		}
		if userID, ok := event.Properties["user_id"].(string); ok && userID != "" {
			identityData["user_id"] = userID
		}
		if userID, ok := event.Properties["phone_number"].(string); ok && userID != "" {
			identityData["phone_number"] = userID
		}
		if err := profileRepo.UpsertIdentityData(permaID, identityData); err != nil {
			return fmt.Errorf("failed to enrich identity data: %v", err)
		}
	}

	return nil
}

func CountEventsMatchingRule(permaID string, trigger models.RuleTrigger, timeRange string) (int, error) {
	eventRepo := repositories.NewEventRepository(locks.GetMongoDBInstance().Database, "events")

	// Parse duration in minutes
	durationInSec, err := strconv.Atoi(timeRange) // parse string to int
	if err != nil {
		log.Printf("Invalid time range format: %v", err)
		//return
	}

	currentTime := time.Now().UTC().Unix()          // current time in seconds
	startTime := currentTime - int64(durationInSec) // assuming value is in minutes

	// Build MongoDB filter
	filter := bson.M{
		"perma_id":   permaID,
		"event_type": strings.ToLower(trigger.EventType),
		"event_name": strings.ToLower(trigger.EventName),
		"event_timestamp": bson.M{
			"$gte": startTime,
		},
	}

	// Fetch matching events
	events, err := eventRepo.FindEventsWithFilter(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch events for counting: %v", err)
	}

	count := 0
	for _, event := range events {
		if EvaluateConditions(event, trigger.Conditions) {
			count++
		}
	}

	return count, nil
}

func GetNestedTraitValue(profile models.Profile, traitPath string) interface{} {
	parts := strings.Split(traitPath, ".")
	if len(parts) < 2 {
		return nil
	}
	namespace := parts[0] // identity, session, etc.
	fieldPath := parts[1:]

	var data map[string]interface{}
	switch namespace {
	case "identityAttributes":
		data = profile.IdentityAttributes
	case "applicationData":
		//data = profile.AppContext
		// todo: need to fix this place
		return nil
	case "traits":
		data = profile.Traits
	default:
		return nil
	}

	curr := data
	for i, part := range fieldPath {
		if i == len(fieldPath)-1 {
			return curr[part]
		}
		if next, ok := curr[part].(map[string]interface{}); ok {
			curr = next
		} else {
			return nil
		}
	}

	return nil
}

func EvaluateConditions(event models.Event, triggerConditions []models.RuleCondition) bool {
	for _, cond := range triggerConditions {
		fieldVal := GetFieldFromEvent(event, cond.Field)
		if !EvaluateCondition(fieldVal, cond.Operator, cond.Value) {
			return false
		}
	}
	return true
}
func EvaluateCondition(actual interface{}, operator string, expected string) bool {
	switch strings.ToLower(operator) {
	case "equals":
		return fmt.Sprintf("%v", actual) == expected

	case "not_equals":
		return fmt.Sprintf("%v", actual) != expected

	case "exists":
		return actual != nil && fmt.Sprintf("%v", actual) != ""

	case "not_exists":
		return actual == nil || fmt.Sprintf("%v", actual) == ""

	case "contains":
		if str, ok := actual.(string); ok {
			return strings.Contains(str, expected)
		}
		return false

	case "not_contains":
		if str, ok := actual.(string); ok {
			return !strings.Contains(str, expected)
		}
		return false

	case "greater_than":
		return compareNumeric(actual, expected, ">")

	case "greater_than_equals":
		return compareNumeric(actual, expected, ">=")

	case "less_than":
		return compareNumeric(actual, expected, "<")

	case "less_than_equals":
		return compareNumeric(actual, expected, "<=")

	default:
		return false
	}
}

func compareNumeric(actual interface{}, expected string, op string) bool {
	actualFloat, err1 := toFloat(actual)
	expectedFloat, err2 := strconv.ParseFloat(expected, 64)
	if err1 != nil || err2 != nil {
		return false
	}

	switch op {
	case ">":
		return actualFloat > expectedFloat
	case ">=":
		return actualFloat >= expectedFloat
	case "<":
		return actualFloat < expectedFloat
	case "<=":
		return actualFloat <= expectedFloat
	default:
		return false
	}
}

func GetFieldFromEvent(event models.Event, field string) interface{} {
	if event.Properties == nil {
		return nil
	}

	if val, ok := event.Properties[field]; ok {
		return val
	}
	return nil
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float")
	}
}

func MergeTraitValue(existing interface{}, incoming interface{}, strategy string, valueType string) interface{} {
	switch strings.ToLower(strategy) {
	case "overwrite":
		if incoming == nil {
			return existing
		}
		if incoming == "" {
			return existing
		}
		return incoming

	case "ignore":
		if existing != nil {
			return existing
		}
		return incoming

	case "combine":
		switch strings.ToLower(valueType) {
		case "arrayofint":
			return combineUniqueInts(toIntSlice(existing), toIntSlice(incoming))
		case "arrayofstring":
			existingArr := toStringSlice(existing)
			incomingArr := toStringSlice(incoming)
			return combineUniqueStrings(existingArr, incomingArr)
		default:
			return incoming
		}

	default:
		// fallback to overwrite
		return incoming
	}
}

func toStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case string:
		return []string{v}
	case []interface{}:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case primitive.A:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}

func toIntSlice(value interface{}) []int {
	switch v := value.(type) {
	case []int:
		return v
	case []interface{}:
		result := make([]int, 0, len(v))
		for _, item := range v {
			if i, ok := item.(float64); ok {
				result = append(result, int(i))
			} else if i, ok := item.(int); ok {
				result = append(result, i)
			}
		}
		return result
	case int:
		return []int{v}
	case float64:
		return []int{int(v)}
	default:
		return []int{}
	}
}

func combineUniqueStrings(a, b []string) []string {
	seen := make(map[string]bool)
	var combined []string
	for _, val := range append(a, b...) {
		if !seen[val] {
			seen[val] = true
			combined = append(combined, val)
		}
	}
	return combined
}

func combineUniqueInts(a, b []int) []int {
	seen := make(map[int]bool)
	var combined []int
	for _, val := range append(a, b...) {
		if !seen[val] {
			seen[val] = true
			combined = append(combined, val)
		}
	}
	return combined
}
