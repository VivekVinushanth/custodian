package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
	"strings"
)

// UserInput represents input data for creating a profile
type UserInput struct {
	OriginCountry string                  `json:"origin_country" binding:"required"`
	UserIds       []string                `json:"user_ids,omitempty"`
	Identity      *models.IdentityData    `json:"identity,omitempty"`
	Personality   *models.PersonalityData `json:"personality,omitempty"`
	AppContext    []models.AppContext     `json:"app_context,omitempty"`
}

// CreateOrUpdateProfile handles the business logic of creating a new user profile
func CreateOrUpdateProfile(input UserInput) (*models.Profile, error) {

	// ✅ Validate that all `app_context` items have `app_id`
	if input.OriginCountry == "" {
		//logger.LogMessage("ERROR", "Validation failed: app_id is required in app_context")
		return nil, errors.New("origin_country is required")
	}

	// ✅ Validate that all `app_context` items have `app_id`
	for _, app := range input.AppContext {
		if app.AppID == "" {
			//logger.LogMessage("ERROR", "Validation failed: app_id is required in app_context")
			return nil, errors.New("app_id is required in app_context")
		}
	}

	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")

	newProfile := models.Profile{
		PermaID:       uuid.NewString(),
		OriginCountry: input.OriginCountry,
		UserIds:       input.UserIds,
		Identity:      input.Identity,
		Personality:   input.Personality,
		AppContext:    input.AppContext,
	}

	// 🔹 Step 1: Fetch all unification rules
	unificationRules, err := unificationRepo.GetUnificationRules()
	if err != nil {
		return nil, errors.New("failed to fetch unification rules")
	}

	// 🔹 Step 2: Fetch all existing profiles from DB
	existingProfiles, err := profileRepo.GetAllProfiles()
	if err != nil {
		return nil, errors.New("failed to fetch existing profiles")
	}

	// 🔹 Step 3: Loop through unification rules and compare profiles
	for _, rule := range unificationRules {
		sortRulesByPriority(rule.Rules)

		// Check each existing profile against the new profile
		for _, existingProfile := range existingProfiles {
			if doesProfileMatch(existingProfile, newProfile, rule) {
				// Merge profiles and update DB
				mergedProfile := mergeProfiles(existingProfile, newProfile)
				for _, appCtx := range mergedProfile.AppContext {
					err := profileRepo.AddOrUpdateAppContext(mergedProfile.PermaID, appCtx)
					if err != nil {
						log.Println("Failed to update AppContext for:", appCtx.AppID, "Error:", err)
					}
				}
				profileRepo.AddOrUpdateUserIds(mergedProfile.PermaID, mergedProfile.UserIds)

				if mergedProfile.Personality != nil {
					err := profileRepo.AddOrUpdatePersonalityData(mergedProfile.PermaID, *mergedProfile.Personality)
					if err != nil {
						log.Println("Failed to update PersonalityData:", err)
					}
				}
				profileRepo.AddOrUpdateIdentityData(mergedProfile.PermaID, *mergedProfile.Identity)

				return &mergedProfile, nil
			}
		}
	}

	// 🔹 Step 4: No match found, create a new profile
	_, err = profileRepo.InsertProfile(newProfile)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to save profile to MongoDB")
		return nil, err
	}
	return &newProfile, nil
}

// doesProfileMatch checks if two profiles have matching attributes based on a unification rule
func doesProfileMatch(existingProfile models.Profile, newProfile models.Profile, rule models.UnificationRule) bool {
	// Convert Profiles to JSON bytes (`[]byte`)
	existingJSON, _ := json.Marshal(existingProfile)
	log.Print(string(existingJSON))
	newJSON, _ := json.Marshal(newProfile)

	// Iterate over all rule attributes
	for _, attrRule := range rule.Rules {
		existingValues := extractFieldFromJSON(existingJSON, attrRule.Attribute)
		log.Print(existingValues)
		newValues := extractFieldFromJSON(newJSON, attrRule.Attribute)
		log.Print(newValues)
		if checkForMatch(existingValues, newValues) {
			return true // ✅ Match found
		}
	}

	return false // ❌ No match found
}

// extractFieldFromJSON extracts a nested field from raw JSON (`[]byte`) without pre-converting to a map
func extractFieldFromJSON(jsonData []byte, fieldPath string) []interface{} {
	var jsonObj interface{}
	err := json.Unmarshal(jsonData, &jsonObj)
	if err != nil {
		return nil // ❌ Return nil if JSON parsing fails
	}

	// Navigate the JSON dynamically
	return getNestedJSONField(jsonObj, fieldPath)
}

// getNestedJSONField retrieves a nested field from a parsed JSON object
func getNestedJSONField(jsonObj interface{}, fieldPath string) []interface{} {
	fields := strings.Split(fieldPath, ".")
	var value interface{} = jsonObj

	for _, field := range fields {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			value = nestedMap[field]
		} else if nestedSlice, ok := value.([]interface{}); ok {
			var results []interface{}
			for _, item := range nestedSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					extracted := getNestedJSONField(itemMap, strings.Join(fields[1:], "."))
					results = append(results, extracted...)
				}
			}
			return results
		} else {
			return nil
		}
	}

	if list, ok := value.([]interface{}); ok {
		return list // Return extracted values from the list
	}

	return []interface{}{value} // Wrap a single value in a list
}

// checkForMatch checks if at least one value from `newProfile` exists in `existingProfile`
func checkForMatch(existingValues, newValues []interface{}) bool {
	existingSet := make(map[string]bool)
	for _, val := range existingValues {
		if str, ok := val.(string); ok {
			existingSet[str] = true
		}
	}

	// 🔹 Check if at least one value from `newValues` exists in `existingSet`
	for _, val := range newValues {
		if str, ok := val.(string); ok {
			if existingSet[str] {
				return true
			}
		}
	}
	return false
}

// GetProfile retrieves a profile from MongoDB by `perma_id`
func GetProfile(permaID string) (*models.Profile, error) {
	//logger := pkg.GetLogger()
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	profile, err := profileRepo.FindProfileByID(permaID)
	if err != nil {
		//logger.LogMessage("ERROR", "Error retrieving profile for PermaID: "+permaID)
		return nil, err
	}
	if profile == nil {
		//logger.LogMessage("INFO", "Profile not found for PermaID: "+permaID)
		return nil, nil
	}

	return profile, nil
}

// DeleteProfile removes a profile from MongoDB by `perma_id`
func DeleteProfile(permaID string) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events") // assuming your event collection name is "events"

	// 🔹 Fetch the existing profile before deletion
	existingProfile, err := profileRepo.FindProfileByID(permaID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	// 🔹 Delete related events
	if err := eventRepo.DeleteEventsByPermaID(permaID); err != nil {
		// Optional: log the error but still return the deleted profile
		log.Println("Failed to delete events for PermaID:", permaID)
	}

	// 🔹 Delete the profile
	err = profileRepo.DeleteProfile(permaID)
	if err != nil {
		return nil, errors.New("failed to delete profile")
	}

	return existingProfile, nil
}

func GetAllProfiles() ([]models.Profile, error) {
	//logger := pkg.GetLogger()
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	existingProfiles, err := profileRepo.GetAllProfiles()
	if err != nil {
		//logger.LogMessage("ERROR", "Error retrieving profile for PermaID: "+permaID)
		return nil, err
	}

	return existingProfiles, nil
}

// AddOrUpdateAppContext replaces (PUT) or inserts a new AppContext inside Profile
func AddOrUpdateAppContext(permaID string, appContext models.AppContext) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.AddOrUpdateAppContext(permaID, appContext)
}

// UpdateAppContextData applies PATCH updates to specific fields of AppContext
func UpdateAppContextData(permaID, appID string, updates bson.M) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.PatchAppContext(permaID, appID, updates)
}

// GetAppContextData fetches app context for a specific app inside a profile
func GetAppContextData(permaID, appID string) (*models.AppContext, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.GetAppContext(permaID, appID)
}

// GetListOfAppContextData fetches all app contexts for a user
func GetListOfAppContextData(permaID string) ([]models.AppContext, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.GetListOfAppContext(permaID)
}

// AddOrUpdatePersonalityData replaces (PUT) the personality data inside Profile
func AddOrUpdatePersonalityData(permaID string, personalityData models.PersonalityData) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.AddOrUpdatePersonalityData(permaID, personalityData)
}

// UpdatePersonalityData_0 applies PATCH updates to specific fields of PersonalityData
func UpdatePersonalityData(permaID string, updates bson.M) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.UpdatePersonalityData(permaID, updates)
}

// GetPersonalityProfileData fetches personality data from a profile
func GetPersonalityProfileData(permaID string) (*models.PersonalityData, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.GetPersonalityProfileData(permaID)
}

// sortRulesByPriority sorts unification rule attributes by priority (lowest first)
func sortRulesByPriority(rules []models.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})
}

// mergeProfiles merges two profiles using `unify` or `combine` strategies
func mergeProfiles(existing models.Profile, newProfile models.Profile) models.Profile {
	mergedProfile := existing

	// 🔹 Merge `user_ids`
	mergedProfile.UserIds = mergeUserIDs(existing.UserIds, newProfile.UserIds)

	// 🔹 Merge `identity`
	if newProfile.Identity != nil {
		mergedProfile.Identity = newProfile.Identity
	}

	// 🔹 Merge `personality`
	if newProfile.Personality != nil {
		if mergedProfile.Personality == nil {
			mergedProfile.Personality = newProfile.Personality
		} else {
			mergedProfile.Personality.Interests = mergeLists(existing.Personality.Interests, newProfile.Personality.Interests)
			//mergedProfile.Personality.CommunicationPreferences = mergeCommunicationPreferences(existing.Personality.CommunicationPreferences, newProfile.Personality.CommunicationPreferences)
		}
	}

	// 🔹 Merge `app_context` grouped by `app_id`
	if newProfile.AppContext != nil {
		mergedProfile.AppContext = mergeAppContexts(existing.AppContext, newProfile.AppContext)
	}

	return mergedProfile
}

// mergeUserIDs combines two lists of user IDs and removes duplicates
func mergeUserIDs(existing, incoming []string) []string {
	idSet := make(map[string]bool)
	var merged []string

	for _, id := range existing {
		if !idSet[id] {
			merged = append(merged, id)
			idSet[id] = true
		}
	}

	for _, id := range incoming {
		if !idSet[id] {
			merged = append(merged, id)
			idSet[id] = true
		}
	}

	return merged
}

// mergeAppContexts merges app contexts, ensuring grouping by `app_id`
func mergeAppContexts(existing []models.AppContext, newContexts []models.AppContext) []models.AppContext {
	appContextMap := make(map[string]models.AppContext)

	// 🔹 Add existing app contexts to the map
	for _, app := range existing {
		appContextMap[app.AppID] = app
	}

	// 🔹 Merge new app contexts
	for _, newApp := range newContexts {
		if existingApp, found := appContextMap[newApp.AppID]; found {
			// 🔹 Merge attributes if `app_id` exists
			existingApp.SubscriptionPlan = highestTier(existingApp.SubscriptionPlan, newApp.SubscriptionPlan)
			existingApp.AppPermissions = mergeLists(existingApp.AppPermissions, newApp.AppPermissions)
			existingApp.Devices = mergeDeviceLists(existingApp.Devices, newApp.Devices)
			appContextMap[newApp.AppID] = existingApp
		} else {
			// 🔹 Add new app context if `app_id` doesn't exist
			appContextMap[newApp.AppID] = newApp
		}
	}

	// 🔹 Convert map back to list
	var mergedAppContexts []models.AppContext
	for _, app := range appContextMap {
		mergedAppContexts = append(mergedAppContexts, app)
	}

	return mergedAppContexts
}

// mergeLists merges lists and removes duplicates
func mergeLists(existing, newList []string) []string {
	uniqueMap := make(map[string]bool)
	for _, item := range existing {
		uniqueMap[item] = true
	}
	for _, item := range newList {
		uniqueMap[item] = true
	}

	var mergedList []string
	for item := range uniqueMap {
		mergedList = append(mergedList, item)
	}
	return mergedList
}

// highestTier returns the highest tier between two subscription plans
func highestTier(existing, newTier string) string {
	tierOrder := map[string]int{"free": 1, "basic": 2, "premium": 3, "enterprise": 4}

	existingPriority, exists := tierOrder[existing]
	newPriority, newExists := tierOrder[newTier]

	if !exists && newExists {
		return newTier
	}
	if newExists && newPriority > existingPriority {
		return newTier
	}
	return existing
}

// mergeDeviceLists merges devices, ensuring no duplicates based on `device_id`
func mergeDeviceLists(existingDevices, newDevices []models.AppContextDevices) []models.AppContextDevices {
	deviceMap := make(map[string]models.AppContextDevices)

	for _, device := range existingDevices {
		deviceMap[device.DeviceID] = device
	}
	for _, device := range newDevices {
		deviceMap[device.DeviceID] = device
	}

	var mergedDevices []models.AppContextDevices
	for _, device := range deviceMap {
		mergedDevices = append(mergedDevices, device)
	}
	return mergedDevices
}
