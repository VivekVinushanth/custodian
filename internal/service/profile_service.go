package service

import (
	"custodian/internal/constants"
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"custodian/internal/utils"
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

//// UserInput represents input data for creating a profile
//type UserInput struct {
//	OriginCountry string                  `json:"origin_country" binding:"required"`
//	UserIds       []string                `json:"user_ids,omitempty"`
//	Identity      *models.IdentityData    `json:"identity,omitempty"`
//	Personality   *models.PersonalityData `json:"personality,omitempty"`
//	AppContext    []models.AppContext     `json:"app_context,omitempty"`
//}

func CreateOrUpdateProfile(event models.Event) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	lock := pkg.GetDistributedLock()
	lockKey := "lock:profile:" + event.PermaId

	// 🔁 Retry logic for acquiring the lock
	maxAttempts := 10
	var acquired bool
	var err error
	for i := 0; i < maxAttempts; i++ {
		acquired, err = lock.Acquire(lockKey, 1*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire lock: %v", err)
		}
		if acquired {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !acquired {
		return nil, fmt.Errorf("could not acquire lock for profile %s after retries", event.PermaId)
	}
	defer lock.Release(lockKey)

	// todo: if we gather login events from IS then we can use tbat userid - check if it is attached to any profile already-if so append if not then create a new profile
	// Safe insert if not exists (upsert)
	profile := models.Profile{
		PermaId: event.PermaId,
		ProfileHierarchy: &models.ProfileHierarchy{
			IsParent:    true,
			ListProfile: true,
		},
	}
	log.Println("before profile added", event.PermaId)

	if err := profileRepo.InsertProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to insert or ensure profile: %v", err)
	}

	// Wait for consistency
	profileFetched, err := waitForProfile(event.PermaId, 10, 100*time.Millisecond)
	if err != nil || profileFetched == nil {
		return nil, fmt.Errorf("profile not visible after insert: %v", err)
	}

	log.Println("profile added succesfully", profileFetched.PermaId)
	return profileFetched, nil
}

func unifyProfiles(newProfile models.Profile) (*models.Profile, error) {
	mongoDB := pkg.GetMongoDBInstance()

	lock := pkg.GetDistributedLock()
	lockKey := "lock:unify:" + newProfile.PermaId

	// Try to acquire the lock before doing unification
	acquired, err := lock.Acquire(lockKey, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for unification: %v", err)
	}
	if !acquired {
		log.Println("Unification already in progress for:", newProfile.PermaId)
		return nil, nil // Or retry logic if needed
	}
	defer lock.Release(lockKey) // Always release

	unificationRepo := repositories.NewResolutionRuleRepository(mongoDB.Database, constants.ResolutionRulesCollection)
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, constants.ProfileCollection)

	// Step 1: Fetch all unification rules
	unificationRules, err := unificationRepo.GetResolutionRules()
	if err != nil {
		return nil, errors.New("failed to fetch unification rules")
	}

	// 🔹 Step 2: Fetch all existing profiles from DB
	existingMasterProfiles, err := profileRepo.GetAllMasterProfilesExceptForCurrent(newProfile)
	if err != nil {
		return nil, errors.New("failed to fetch existing profiles")
	}

	sortRulesByPriority(unificationRules)
	// 🔹 Step 3: Loop through unification rules and compare profiles
	for _, rule := range unificationRules {

		for _, existingProfile := range existingMasterProfiles {

			if doesProfileMatch(existingProfile, newProfile, rule) {
				log.Println("Unifying profiles:", existingProfile.PermaId, "and", newProfile.PermaId, "on ", rule.RuleName)

				// 🔄 Merge the existing master to the old master of current
				mongoDB := pkg.GetMongoDBInstance()
				schemaRepo := repositories.NewProfileSchemaRepository(mongoDB.Database, "profile_schema")
				traitRules, _ := schemaRepo.GetSchemaRules()
				newMasterProfile := MergeProfileFields(existingProfile, newProfile, traitRules)

				if len(existingProfile.ProfileHierarchy.ChildProfiles) == 0 {
					newMasterProfile.PermaId = uuid.New().String()
					childProfile1 := models.ChildProfile{
						ChildProfileId: newProfile.PermaId,
						RuleName:       rule.RuleName,
					}
					childProfile2 := models.ChildProfile{
						ChildProfileId: existingProfile.PermaId,
						RuleName:       rule.RuleName,
					}
					newMasterProfile.ProfileHierarchy = &models.ProfileHierarchy{
						IsParent:      true,
						ListProfile:   false,
						ChildProfiles: []models.ChildProfile{childProfile1, childProfile2},
					}
					// creating and inserting the new master profile
					err := profileRepo.InsertProfile(newMasterProfile)
					if err != nil {
						return nil, err
					}

					// Attaching peer profiles for each of the child profiles of old master profile
					//profileRepo.LinkPeers(newProfile.PermaId, existingProfile.PermaId, rule.RuleName)
					err = profileRepo.UpdateParent(newMasterProfile, newProfile)
					err = profileRepo.UpdateParent(newMasterProfile, existingProfile)
					if err != nil {
						return nil, err
					}

				} else if (len(existingProfile.ProfileHierarchy.ChildProfiles) > 0) && existingProfile.ProfileHierarchy.IsParent {
					newChild := models.ChildProfile{
						ChildProfileId: newProfile.PermaId,
						RuleName:       rule.RuleName,
					}
					err = profileRepo.AddChildProfile(newMasterProfile, newChild)
					err = profileRepo.UpdateParent(newMasterProfile, newProfile)
					if err != nil {
						return nil, err
					}
				}

				// Update AppContext
				for _, appCtx := range newMasterProfile.AppContext {
					err := profileRepo.AddOrUpdateAppContext(newMasterProfile.PermaId, appCtx)
					if err != nil {
						log.Println("Failed to update AppContext for:", appCtx.AppID, "Error:", err)
					}
				}

				// Update Personality
				if newMasterProfile.Personality != nil {
					err := profileRepo.AddOrUpdatePersonalityData(newMasterProfile.PermaId, newMasterProfile.Personality)
					if err != nil {
						log.Println("Failed to update PersonalityData:", err)
					}
				}

				// Update Identity
				if newMasterProfile.Identity != nil {
					err := profileRepo.UpsertIdentityData(newMasterProfile.PermaId, newMasterProfile.Identity)
					if err != nil {
						log.Println("Failed to update IdentityData:", err)
					}
				}

				return &newMasterProfile, nil

			}
		}
	}

	// No unification match found, return newProfile as-is
	return &newProfile, nil
}

// doesProfileMatch checks if two profiles have matching attributes based on a unification rule
func doesProfileMatch(existingProfile models.Profile, newProfile models.Profile, rule models.ResolutionRule) bool {
	// Convert Profiles to JSON bytes (`[]byte`)
	existingJSON, _ := json.Marshal(existingProfile)
	//log.Print(string(existingJSON))
	newJSON, _ := json.Marshal(newProfile)

	// Iterate over all rule attributes
	//for _, attrRule := range rule. {
	existingValues := extractFieldFromJSON(existingJSON, rule.Attribute)
	newValues := extractFieldFromJSON(newJSON, rule.Attribute)
	if checkForMatch(existingValues, newValues) {
		return true // ✅ Match found
	}
	//}

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
		return nil, errors.New("profile not found12122")
	}

	if profile.ProfileHierarchy.IsParent {
		return profile, nil
	} else {
		// fetching merged master profile
		masterProfile, err := profileRepo.FindProfileByID(profile.ProfileHierarchy.ParentProfileID)

		// only session data of the relevant perma id is returned
		masterProfile.Session = profile.Session
		// todo: app context should be restricted for apps that is requesting these
		// setting the current profile hierarchy to the master profile
		masterProfile.ProfileHierarchy = buildProfileHierarchy(profile, masterProfile)
		masterProfile.PermaId = profile.PermaId
		if err != nil {
			//logger.LogMessage("ERROR", "Error retrieving profile for PermaId: "+permaID)
			return nil, err
		}
		if masterProfile == nil {
			//logger.LogMessage("INFO", "Profile not found for PermaId: "+permaID)
			return nil, nil
		}
		return masterProfile, nil
	}
}

func buildProfileHierarchy(profile *models.Profile, masterProfile *models.Profile) *models.ProfileHierarchy {

	profileHierarchy := &models.ProfileHierarchy{
		IsParent:        false,
		ListProfile:     true,
		ParentProfileID: masterProfile.PermaId,
		ChildProfiles:   []models.ChildProfile{},
	}
	if len(masterProfile.ProfileHierarchy.ChildProfiles) > 0 {
		profileHierarchy.ChildProfiles = masterProfile.ProfileHierarchy.ChildProfiles
	}
	return profileHierarchy
}

// GetProfileWithToken retrieves a profile from MongoDB by `perma_id`
func GetProfileWithToken(permaID string, token string) (*models.Profile, error) {
	//mongoDB := pkg.GetMongoDBInstance()
	//profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	profile, _ := GetProfile(permaID)
	if profile == nil {
		return nil, errors.New("profile not found")
	}

	if profile.Identity == nil {
		return profile, nil
		// identity not found in profile
	}
	// Safely fetch userId from profile.Identity
	userIDRaw, ok := profile.Identity["user_id"]
	log.Println("user id raw ===", userIDRaw)
	log.Println("uokkkk=", ok)
	if !ok {
		return profile, nil
	}
	log.Println("user id raw ===", userIDRaw)
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		return nil, errors.New("invalid user_id in profile.identity")
	}

	// Fetch SCIM user data
	userData, err := utils.GetUserDataFromSCIM(token, userID)
	log.Println("userData===", userData)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user data from SCIM: %v", err)
	}
	if userData == nil {
		return nil, errors.New("user data not found from SCIM")
	}

	// Merge enriched SCIM data into identity (in-memory only)
	enrichedIdentity := utils.ExtractIdentityFromSCIM(userData)
	for k, v := range enrichedIdentity {
		profile.Identity[k] = v
	}

	return profile, nil
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
		log.Println("Failed to delete events for PermaId:", permaID)
	}

	if profile.ProfileHierarchy.IsParent && len(profile.ProfileHierarchy.ChildProfiles) == 0 {
		// Delete the master with no children
		err = profileRepo.DeleteProfile(permaID)
		if err != nil {
			return nil, errors.New("failed to delete profile")
		}
	}

	if profile.ProfileHierarchy.IsParent && len(profile.ProfileHierarchy.ChildProfiles) >= 0 {
		//get all child profiles and delete
		for _, childProfile := range profile.ProfileHierarchy.ChildProfiles {
			_, err := profileRepo.FindProfileByID(childProfile.ChildProfileId)
			if err != nil {
				return nil, errors.New("failed to delete child profile")
			}
			err = profileRepo.DeleteProfile(childProfile.ChildProfileId)
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

	if !(profile.ProfileHierarchy.IsParent) {
		err = profileRepo.DeleteProfile(permaID)
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

func waitForProfileWithUserName(userId string, maxRetries int, delay time.Duration) (*models.Profile, error) {
	profileRepo := repositories.NewProfileRepository(pkg.GetMongoDBInstance().Database, "profiles")

	for i := 0; i < maxRetries; i++ {
		profile, err := profileRepo.FindProfileByUserName(userId)
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
		//logger.LogMessage("ERROR", "Error retrieving profile for PermaId: "+permaID)
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
func AddOrUpdatePersonalityData(permaID string, personalityData map[string]interface{}) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.AddOrUpdatePersonalityData(permaID, personalityData)
}

// UpdatePersonalityData_0 applies PATCH updates to specific fields of PersonalityData
func UpdatePersonalityData(permaID string, updates bson.M) error {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")
	return profileRepo.UpsertPersonalityData(permaID, updates)
}

// GetPersonalityProfileData fetches personality data from a profile
func GetPersonalityProfileData(permaID string) (map[string]interface{}, error) {
	mongoDB := pkg.GetMongoDBInstance()
	profileRepo := repositories.NewProfileRepository(mongoDB.Database, "profiles")

	profile, _ := profileRepo.FindProfileByID(permaID)
	if profile != nil {
		if profile.ProfileHierarchy.IsParent {
			return profileRepo.GetPersonalityProfileData(permaID)
		} else {
			return profileRepo.GetPersonalityProfileData(profile.ProfileHierarchy.ParentProfileID)
		}
	}
	return nil, errors.New("profile not found")
}

// sortRulesByPriority sorts unification rule attributes by priority (lowest first)
func sortRulesByPriority(rules []models.ResolutionRule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})
}

// mergeProfiles merges two profiles using `unify` or `combine` strategies
func mergeProfiles(existing models.Profile, newProfile models.Profile) models.Profile {
	mergedProfile := existing

	// 🔹 Merge `identity`
	if newProfile.Identity != nil {
		if mergedProfile.Identity == nil {
			mergedProfile.Identity = newProfile.Identity
		} else {
			mergeStructFields(mergedProfile.Identity, newProfile.Identity)
		}
	}

	// 🔹 Merge `personality`
	//if newProfile.Personality != nil {
	//	if mergedProfile.Personality == nil {
	//		mergedProfile.Personality = newProfile.Personality
	//	} else {
	//		mergedProfile.Personality.Interests = mergeLists(existing.Personality.Interests, newProfile.Personality.Interests)
	//		//mergedProfile.Personality.CommunicationPreferences = mergeCommunicationPreferences(existing.Personality.CommunicationPreferences, newProfile.Personality.CommunicationPreferences)
	//	}
	//}

	// 🔹 Merge `app_context` grouped by `app_id`
	if newProfile.AppContext != nil {
		mergedProfile.AppContext = mergeAppContexts(existing.AppContext, newProfile.AppContext)
	}

	return mergedProfile
}

func MergeProfileFields(existing, incoming models.Profile, traitRules []models.ProfileEnrichmentRule) models.Profile {
	merged := existing // start with existing

	for _, rule := range traitRules {
		traitPath := strings.Split(rule.TraitName, ".")
		if len(traitPath) < 2 {
			continue
		}

		traitNamespace := traitPath[0] // e.g., "personality"
		traitKey := traitPath[1]       // e.g., "interests"

		var existingVal, newVal interface{}
		switch traitNamespace {
		case "personality":
			if existing.Personality != nil {
				existingVal = existing.Personality[traitKey]
			}
			if incoming.Personality != nil {
				newVal = incoming.Personality[traitKey]
			}
		case "identity":
			if existing.Identity != nil {
				existingVal = existing.Identity[traitKey]
			}
			if incoming.Identity != nil {
				newVal = incoming.Identity[traitKey]
			}
		case "app_context":
			if incoming.AppContext != nil {
				incoming.AppContext = mergeAppContexts(existing.AppContext, incoming.AppContext)
			}
		}

		// todo: Note that merging with less info and then later profiel with max info leads to issue or rather without merge strat they are left out
		// Perform merge based on strategy
		mergedVal := MergeTraitValue(existingVal, newVal, rule.MergeStrategy, rule.ValueType)

		// Apply merged result
		switch traitNamespace {
		case "personality":
			if merged.Personality == nil {
				merged.Personality = map[string]interface{}{}
			}
			merged.Personality[traitKey] = mergedVal
		case "identity":

			//if merged.Identity == nil {
			//	merged.Identity = map[string]interface{}{}
			//}
			//merged.Identity[traitKey] = mergedVal
		}
	}

	return merged
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

// mergeAppContexts merges app contexts, ensuring grouping by `app_id`
func mergeAppContexts(existing []models.AppContext, newContexts []models.AppContext) []models.AppContext {
	appContextMap := make(map[string]models.AppContext)

	// 🔹 AddEventSchema existing app contexts to the map
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
			// 🔹 AddEventSchema new app context if `app_id` doesn't exist
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
