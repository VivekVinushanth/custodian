package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"
)

// UserInput represents input data for creating a profile
type UserInput struct {
	OriginCountry string                  `json:"origin_country" binding:"required"`
	UserIds       []string                `json:"user_ids,omitempty"`
	Identity      *models.IdentityData    `json:"identity,omitempty"`
	Personality   *models.PersonalityData `json:"personality,omitempty"`
	AppContext    []models.AppContext     `json:"app_context,omitempty"`
}

func CreateOrUpdateProfile(event models.Event) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	lock := pkg.GetDistributedLock()
	lockKey := "lock:profile:" + event.PermaID

	// 🔁 Retry logic for acquiring the lock
	maxAttempts := 5
	retryDelay := 100 * time.Millisecond
	var acquired bool
	var err error

	for i := 0; i < maxAttempts; i++ {
		acquired, err = lock.Acquire(lockKey, 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire lock for unification: %v", err)
		}
		if acquired {
			break
		}
		log.Printf("Lock busy for %s, retrying (%d/%d)...", event.PermaID, i+1, maxAttempts)
		time.Sleep(retryDelay)
	}

	if !acquired {
		return nil, fmt.Errorf("could not acquire lock for profile %s after %d attempts", event.PermaID, maxAttempts)
	}
	defer lock.Release(lockKey)

	// 🔸 Step 1: Insert profile (upsert behavior)
	initialProfile := models.Profile{
		PermaID: event.PermaID,
		ProfileHierarchy: &models.ProfileHierarchy{
			IsMaster:    true,
			ListProfile: true,
		},
	}
	_ = profileRepo.InsertProfile(initialProfile)

	// 🔁 Step 2: Wait for profile to be available (optional for visibility delay)
	profile, err := waitForProfile(event.PermaID, 10, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile after insert attempt: %v", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("profile still not visible after retries")
	}

	// 🔸 Step 3: Enrich profile
	if err := EnrichProfile(event); err != nil {
		return nil, err
	}

	// 🔸 Step 4: Re-fetch and unify
	enrichedProfile, err := profileRepo.FindProfileByID(event.PermaID)
	if err != nil {
		return nil, fmt.Errorf("failed to re-fetch profile after enrichment: %v", err)
	}
	if enrichedProfile == nil {
		return nil, fmt.Errorf("profile unexpectedly missing after enrichment")
	}

	_, err = unifyProfiles(*enrichedProfile)
	if err != nil {
		return nil, err
	}

	return enrichedProfile, nil
}

func unifyProfiles(newProfile models.Profile) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()

	lock := pkg.GetDistributedLock()
	lockKey := "lock:unify:" + newProfile.PermaID

	// Try to acquire the lock before doing unification
	acquired, err := lock.Acquire(lockKey, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for unification: %v", err)
	}
	if !acquired {
		log.Println("Unification already in progress for:", newProfile.PermaID)
		return nil, nil // Or retry logic if needed
	}
	defer lock.Release(lockKey) // Always release

	unificationRepo := repositories.NewUnificationRepository(mongoDB.Database, "unification_rules")
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	// Step 1: Fetch all unification rules
	unificationRules, err := unificationRepo.GetUnificationRules()
	if err != nil {
		return nil, errors.New("failed to fetch unification rules")
	}

	// 🔹 Step 2: Fetch all existing profiles from DB
	existingMasterProfiles, err := profileRepo.GetAllMasterProfilesExceptForCurrent(newProfile)
	if err != nil {
		return nil, errors.New("failed to fetch existing profiles")
	}

	// 🔹 Step 3: Loop through unification rules and compare profiles
	for _, rule := range unificationRules {
		sortRulesByPriority(rule.Rules)

		for _, existingProfile := range existingMasterProfiles {

			if doesProfileMatch(existingProfile, newProfile, rule) {

				// 🔄 Merge the existing master to the old master of current
				newMasterProfile := mergeProfiles(existingProfile, newProfile)

				if len(existingProfile.ProfileHierarchy.ChildProfiles) == 0 {
					newMasterProfile.PermaID = uuid.New().String()
					newMasterProfile.ProfileHierarchy = &models.ProfileHierarchy{
						IsMaster:      true,
						ListProfile:   false,
						ChildProfiles: []string{newProfile.PermaID, existingProfile.PermaID},
					}

					// creating and inserting the new master profile
					err := profileRepo.InsertProfile(newMasterProfile)
					if err != nil {
						return nil, err
					}

					// Attaching peer profiles for each of the child profiles of old master profile
					profileRepo.LinkPeers(newProfile.PermaID, existingProfile.PermaID, rule.RuleName)
					err = profileRepo.UpdateParent(newMasterProfile, newProfile)
					err = profileRepo.UpdateParent(newMasterProfile, existingProfile)
					if err != nil {
						return nil, err
					}

				} else if (len(existingProfile.ProfileHierarchy.ChildProfiles) > 0) && existingProfile.ProfileHierarchy.IsMaster {
					// Loop through all child profile and attach the peer
					for _, childProfileID := range existingProfile.ProfileHierarchy.ChildProfiles {
						if childProfileID == newProfile.PermaID {
							continue
						}
						if err != nil {
							return nil, errors.New("failed to fetch child profile")
						}
						profileRepo.LinkPeers(newProfile.PermaID, childProfileID, rule.RuleName)
					}

					err = profileRepo.UpdateParent(newMasterProfile, newProfile)
					if err != nil {
						return nil, err
					}
				}

				// Update AppContext
				for _, appCtx := range newMasterProfile.AppContext {
					err := profileRepo.AddOrUpdateAppContext(newMasterProfile.PermaID, appCtx)
					if err != nil {
						log.Println("Failed to update AppContext for:", appCtx.AppID, "Error:", err)
					}
				}

				// Update UserIds
				profileRepo.AddOrUpdateUserIds(newMasterProfile.PermaID, newMasterProfile.UserIds)

				// Update Personality
				if newMasterProfile.Personality != nil {
					err := profileRepo.AddOrUpdatePersonalityData(newMasterProfile.PermaID, *newMasterProfile.Personality)
					if err != nil {
						log.Println("Failed to update PersonalityData:", err)
					}
				}

				// Update Identity
				if newMasterProfile.Identity != nil {
					profileRepo.AddOrUpdateIdentityData(newMasterProfile.PermaID, *newMasterProfile.Identity)
				}

				return &newMasterProfile, nil

			}
		}
	}

	// No unification match found, return newProfile as-is
	return &newProfile, nil
}

// doesProfileMatch checks if two profiles have matching attributes based on a unification rule
func doesProfileMatch(existingProfile models.Profile, newProfile models.Profile, rule models.UnificationRule) bool {
	// Convert Profiles to JSON bytes (`[]byte`)
	existingJSON, _ := json.Marshal(existingProfile)
	//log.Print(string(existingJSON))
	newJSON, _ := json.Marshal(newProfile)

	// Iterate over all rule attributes
	for _, attrRule := range rule.Rules {
		existingValues := extractFieldFromJSON(existingJSON, attrRule.Attribute)
		newValues := extractFieldFromJSON(newJSON, attrRule.Attribute)
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

	profile, _ := profileRepo.FindProfileByID(permaID)
	if profile == nil {
		return nil, errors.New("profile not found")
	}

	if profile.ProfileHierarchy.IsMaster {
		return profile, nil
	} else {
		// fetching merged master profile
		masterProfile, err := profileRepo.FindProfileByID(profile.ProfileHierarchy.ParentProfileID)

		// setting the current profile hierarchy to the master profile
		//masterProfile.ProfileHierarchy = buildProfileHierarchy(profile, masterProfile)
		masterProfile.ProfileHierarchy = profile.ProfileHierarchy
		masterProfile.PermaID = profile.PermaID
		if err != nil {
			//logger.LogMessage("ERROR", "Error retrieving profile for PermaID: "+permaID)
			return nil, err
		}
		if masterProfile == nil {
			//logger.LogMessage("INFO", "Profile not found for PermaID: "+permaID)
			return nil, nil
		}
		return masterProfile, nil
	}
}

// DeleteProfile removes a profile from MongoDB by `perma_id`
func DeleteProfile(permaID string) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	eventRepo := repositories.NewEventRepository(mongoDB.Database, "events") // assuming your event collection name is "events"

	// 🔹 Fetch the existing profile before deletion
	profile, err := profileRepo.FindProfileByID(permaID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	// 🔹 Delete related events
	if err := eventRepo.DeleteEventsByPermaID(permaID); err != nil {
		// Optional: log the error but still return the deleted profile
		log.Println("Failed to delete events for PermaID:", permaID)
	}

	if profile.ProfileHierarchy.IsMaster && len(profile.ProfileHierarchy.ChildProfiles) == 0 {
		// Delete the master with no children
		err = profileRepo.DeleteProfile(permaID)
		if err != nil {
			return nil, errors.New("failed to delete profile")
		}
	}

	if profile.ProfileHierarchy.IsMaster && len(profile.ProfileHierarchy.ChildProfiles) >= 0 {
		//get all child profiles and delete
		for _, childProfileID := range profile.ProfileHierarchy.ChildProfiles {
			_, err := profileRepo.FindProfileByID(childProfileID)
			if err != nil {
				return nil, errors.New("failed to delete child profile")
			}
			err = profileRepo.DeleteProfile(childProfileID)
			if err != nil {
				return nil, errors.New("failed to delete child profile")
			}
		}
		// now delete master
		err = profileRepo.DeleteProfile(permaID)
		if err != nil {
			return nil, errors.New("failed to delete profile")
		}
	}

	if !(profile.ProfileHierarchy.IsMaster) {
		err = profileRepo.DeleteProfile(permaID)
		for _, childProfileID := range profile.ProfileHierarchy.ChildProfiles {
			if childProfileID != permaID {
				profileRepo.DetachPeer(permaID, childProfileID)
				profileRepo.DetachPeer(childProfileID, permaID)
			}
		}
		profileRepo.DetachChildFromParent(profile.ProfileHierarchy.ParentProfileID, permaID)

	}

	return profile, nil
}

func waitForProfile(permaID string, maxRetries int, delay time.Duration) (*models.Profile, error) {
	profileRepo := repositories.NewProfileRepository(pkg.GetMongoDBInstance().Database, "profiles")

	for i := 0; i < maxRetries; i++ {
		profile, err := profileRepo.FindProfileByID(permaID)
		if err != nil {
			return nil, err
		}
		if profile != nil {
			return profile, nil
		}
		time.Sleep(delay)
	}
	return nil, nil
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

	profile, _ := profileRepo.FindProfileByID(permaID)
	if profile != nil {
		if profile.ProfileHierarchy.IsMaster {
			return profileRepo.GetPersonalityProfileData(permaID)
		} else {
			return profileRepo.GetPersonalityProfileData(profile.ProfileHierarchy.ParentProfileID)
		}
	}
	return nil, errors.New("profile not found")
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
		if mergedProfile.Identity == nil {
			mergedProfile.Identity = newProfile.Identity
		} else {
			mergeStructFields(mergedProfile.Identity, newProfile.Identity)
		}
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

// mergeStructFields merges non-zero fields from `src` into `dest`
func mergeStructFields(dest interface{}, src interface{}) {
	destVal := reflect.ValueOf(dest).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		field := srcVal.Type().Field(i)
		srcField := srcVal.Field(i)
		destField := destVal.FieldByName(field.Name)

		// Skip if not settable or zero value
		if !destField.CanSet() || isZeroValue(srcField) {
			continue
		}

		// Handle slices: combine with deduplication
		if srcField.Kind() == reflect.Slice {
			merged := mergeSlices(destField.Interface(), srcField.Interface())
			destField.Set(reflect.ValueOf(merged))
			continue
		}

		// Simple overwrite
		destField.Set(srcField)
	}
}

// isZeroValue checks if a field is zero value (e.g. "", nil, 0, false)
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// mergeSlices merges two slices and removes duplicates
func mergeSlices(a, b interface{}) interface{} {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	existing := make(map[interface{}]bool)
	result := reflect.MakeSlice(aVal.Type(), 0, aVal.Len()+bVal.Len())

	// Helper to append unique values
	appendUnique := func(val reflect.Value) {
		if !existing[val.Interface()] {
			existing[val.Interface()] = true
			result = reflect.Append(result, val)
		}
	}

	for i := 0; i < aVal.Len(); i++ {
		appendUnique(aVal.Index(i))
	}
	for i := 0; i < bVal.Len(); i++ {
		appendUnique(bVal.Index(i))
	}

	return result.Interface()
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
