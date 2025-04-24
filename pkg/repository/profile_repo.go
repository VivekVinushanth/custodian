package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"identity-customer-data-service/pkg/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// ProfileRepository handles MongoDB operations for profiles
type ProfileRepository struct {
	Collection *mongo.Collection
}

// NewProfileRepository creates a new repository instance
func NewProfileRepository(db *mongo.Database, collectionName string) *ProfileRepository {
	return &ProfileRepository{
		Collection: db.Collection(collectionName),
	}
}

// InsertProfile saves a profile in MongoDB
func (repo *ProfileRepository) InsertProfile(profile models.Profile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profile.ProfileId}
	update := bson.M{"$setOnInsert": profile}

	opts := options.Update().SetUpsert(true)

	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpdateProfile saves a profile in MongoDB
func (repo *ProfileRepository) UpdateProfile(profile models.Profile) (*mongo.InsertOneResult, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := repo.Collection.InsertOne(ctx, profile)

	if err != nil {
		//logger.LogMessage("ERROR", "Failed to insert profile: "+err.Error())
		return nil, err
	}
	//logger.LogMessage("INFO", "Profile inserted with TraitId: "+result.InsertedID.(string))
	return result, nil
}

// GetProfile retrieves a profile by `profile_id`
func (repo *ProfileRepository) GetProfile(profileId string) (*models.Profile, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"profile_id": profileId}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//logger.LogMessage("INFO", "Profile not found for profileId: "+profileId)
			return nil, nil // Return `nil` instead of error
		}
		//logger.LogMessage("ERROR", "Error finding profile: "+err.Error())
		return nil, err
	}

	//logger.LogMessage("INFO", "Profile retrieved for profileId: "+profileId)
	return &profile, nil
}

// FindProfileByID retrieves a profile by `profile_id`
func (repo *ProfileRepository) FindProfileByID(profileId string) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"profile_id": profileId}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Profile not found is not an error
		}
		log.Print("Error finding profile==123: " + err.Error())
		return nil, err
	}
	return &profile, nil
}

// DeleteProfile removes a profile from MongoDB using `profile_id`
func (repo *ProfileRepository) DeleteProfile(profileId string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}

	result, err := repo.Collection.DeleteOne(ctx, filter)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to delete profile: "+err.Error())
		return err
	}

	if result.DeletedCount == 0 {
		//logger.LogMessage("INFO", "No profile found to delete")
		return mongo.ErrNoDocuments
	}

	//logger.LogMessage("INFO", "Profile deleted successfully")
	return nil
}

func (repo *ProfileRepository) DetachChildFromParent(parentID, childID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{
			"profile_hierarchy.child_profile_ids": bson.M{
				"child_profile_id": childID,
			},
		},
	}
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": parentID}, update)
	return err
}

func (repo *ProfileRepository) DetachPeer(profileID, peerToRemove string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{
			"profile_hierarchy.peer_profile_ids": peerToRemove,
		},
	}
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": profileID}, update)
	return err
}

func (repo *ProfileRepository) AddOrUpdateAppContext(profileId string, newAppCtx models.ApplicationData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Fetch full profile
	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"profile_id": profileId}).Decode(&profile)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	// Step 2: Prepare or update application data
	updated := false
	var updatedAppData []models.ApplicationData

	for _, existing := range profile.ApplicationData {
		if existing.AppId == newAppCtx.AppId {
			// Merge devices
			existing.Devices = mergeDeviceLists(existing.Devices, newAppCtx.Devices)

			// Merge app-specific fields
			if existing.AppSpecificData == nil {
				existing.AppSpecificData = map[string]interface{}{}
			}
			for k, v := range newAppCtx.AppSpecificData {
				existing.AppSpecificData[k] = v
			}

			updatedAppData = append(updatedAppData, existing)
			updated = true
		} else {
			updatedAppData = append(updatedAppData, existing)
		}
	}

	// If no existing entry matched, add new one
	if !updated {
		updatedAppData = append(updatedAppData, newAppCtx)
	}

	// Step 3: Persist the updated application data
	update := bson.M{
		"$set": bson.M{
			"application_data": updatedAppData,
		},
	}
	_, err = repo.Collection.UpdateOne(ctx, bson.M{"profile_id": profileId}, update)
	if err != nil {
		return fmt.Errorf("failed to update application_data: %w", err)
	}

	return nil
}

func (repo *ProfileRepository) PatchAppContext(profileId, appID string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{fmt.Sprintf("application_data.%s", appID): bson.M{"$exists": true}, "profile_id": profileId}
	update := bson.M{"$set": bson.M{fmt.Sprintf("application_data.%s", appID): updates}}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	return err
}

//func (repo *ProfileRepository) GetAppContext(profileId, appID string) (*models.ApplicationData, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	filter := bson.M{"profile_id": profileId}
//	projection := bson.M{"application_data": 1}
//
//	var profile models.Profile
//	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&profile)
//	if err != nil {
//		if err == mongo.ErrNoDocuments {
//			return nil, nil
//		}
//		return nil, err
//	}
//
//	appData, exists := profile.ApplicationData[appID]
//	if !exists {
//		return nil, nil
//	}
//
//	return &appData, nil
//}

func decodeDevices(raw interface{}) []models.Devices {
	var devices []models.Devices

	if rawList, ok := raw.([]interface{}); ok {
		for _, item := range rawList {
			if deviceMap, ok := item.(map[string]interface{}); ok {
				device := models.Devices{}
				data, _ := json.Marshal(deviceMap)
				_ = json.Unmarshal(data, &device)
				devices = append(devices, device)
			}
		}
	}

	return devices
}

func mergeDeviceLists(existing, incoming []models.Devices) []models.Devices {
	deviceMap := make(map[string]models.Devices)
	for _, d := range existing {
		deviceMap[d.DeviceId] = d
	}
	for _, d := range incoming {
		deviceMap[d.DeviceId] = d
	}
	var merged []models.Devices
	for _, d := range deviceMap {
		merged = append(merged, d)
	}
	return merged
}

//func (repo *ProfileRepository) GetListOfAppContext(profileId string) ([]models.ApplicationData, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	filter := bson.M{"profile_id": profileId}
//	projection := bson.M{"application_data": 1}
//
//	var profile models.Profile
//	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&profile)
//	if err == mongo.ErrNoDocuments {
//		return nil, nil
//	} else if err != nil {
//		return nil, fmt.Errorf("failed to retrieve application data: %w", err)
//	}
//
//	var flattened []map[string]interface{}
//
//	for _, app := range profile.ApplicationData {
//		flat := map[string]interface{}{
//			"app_id":  app.AppId,
//			"devices": app.Devices,
//		}
//
//		// Merge app_specific_data into the flat map
//		for k, v := range app.AppSpecificData {
//			flat[k] = v
//		}
//
//		flattened = append(flattened, flat)
//	}
//
//	return flattened, nil
//}

// AddOrUpdatePersonalityData replaces (PUT) the personality data inside Profile
func (repo *ProfileRepository) AddOrUpdatePersonalityData(profileId string, personalityData map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	update := bson.M{"$set": bson.M{"personality": personalityData}}

	opts := options.Update().SetUpsert(true) // Insert if not found
	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

// UpsertIdentityData replaces (PUT) the personality data inside Profile
func (repo *ProfileRepository) UpsertIdentityData(profileId string, identityData map[string]interface{}) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	updateFields := bson.M{}
	for k, v := range identityData {
		updateFields["identity."+k] = v
	}

	update := bson.M{"$set": updateFields}

	opts := options.Update().SetUpsert(true) // Insert if not found
	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data updated for user "+profileId)
	return nil
}

// UpsertTraits applies PATCH updates to specific fields of PersonalityData
func (repo *ProfileRepository) UpsertTraits(profileId string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data patched for user " + profileId)
	return nil
}

// UpdatePreferenceData applies PATCH updates to specific fields of PersonalityData
func (repo *ProfileRepository) UpdatePreferenceData(profileId string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	_, err := repo.Collection.UpdateOne(ctx, filter, updates)
	return err
}

// UpdateIdentityData adds or updates fields in IdentityData without overwriting existing non-empty values
func (repo *ProfileRepository) UpdateIdentityData(profileId string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Ensure identity exists
	identityInit := bson.M{
		"$setOnInsert": bson.M{
			"identity_attributes": bson.M{},
		},
	}
	_, _ = repo.Collection.UpdateOne(ctx, bson.M{"profile_id": profileId, "identity_attributes": bson.M{"$exists": false}}, identityInit)

	// Step 2: Prepare full update
	setUpdates := bson.M{}
	for k, v := range updates {
		setUpdates["identity_attributes."+k] = v
	}

	update := bson.M{
		"$set": setUpdates,
	}

	// Step 3: Apply
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": profileId}, update)
	return err
}

func (repo *ProfileRepository) GetPersonalityProfileData(profileId string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	projection := bson.M{"personality": 1}

	var result struct {
		Personality map[string]interface{} `bson:"personality"`
	}

	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return result.Personality, nil
}

// GetAllProfiles retrieves all profiles from MongoDB
func (repo *ProfileRepository) GetAllProfiles() ([]models.Profile, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Exclude profiles where profile_hierarchy.is_master == true
	// Only fetch profiles with profile_hierarchy.list_profile == true
	filter := bson.M{
		"profile_hierarchy.list_profile": true,
	}

	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to fetch profiles: "+err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var profiles []models.Profile
	// 🔹 Decode all profiles
	if err = cursor.All(ctx, &profiles); err != nil {
		//logger.LogMessage("ERROR", "Error decoding profiles: "+err.Error())
		return nil, err
	}

	//logger.LogMessage("INFO", "Successfully fetched profiles")
	return profiles, nil
}

// GetAllMasterProfilesExceptForCurrent retrieves all master profiles excluding the current profile's parent
func (repo *ProfileRepository) GetAllMasterProfilesExceptForCurrent(currentProfile models.Profile) ([]models.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	excludedIDs := []string{}
	excludedIDs = append(excludedIDs, currentProfile.ProfileId)

	// Fetch only master profiles excluding the parent of the current profile
	filter := bson.M{
		"profile_hierarchy.is_parent": true,
		"profile_id": bson.M{
			"$nin": excludedIDs,
		},
	}

	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var profiles []models.Profile
	if err = cursor.All(ctx, &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

// AddOrUpdateUserIds merges and updates the user_ids list inside the profile
func (repo *ProfileRepository) AddOrUpdateUserIds(profileId string, newUserIds []string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Fetch the existing user_ids
	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"profile_id": profileId}).Decode(&profile)
	if err != nil && err != mongo.ErrNoDocuments {
		//logger.LogMessage("ERROR", "Failed to fetch profile for user_ids merge: "+err.Error())
		return err
	}

	// Step 2: Perform the update
	filter := bson.M{"profile_id": profileId}
	update := bson.M{"$set": bson.M{"userids": newUserIds}}

	opts := options.Update().SetUpsert(true)
	_, err = repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update user_ids: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "User IDs updated for user "+profileId)
	return nil
}

func (repo *ProfileRepository) UpdateParent(master models.Profile, newProfile models.Profile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateProfile := bson.M{
		"$set": bson.M{
			"profile_hierarchy.parent_profile_id": master.ProfileId,
			"profile_hierarchy.is_parent":         false,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": newProfile.ProfileId}, updateProfile); err != nil {
		return fmt.Errorf("failed to update profile %s: %w", newProfile.ProfileId, err)
	}

	return nil
}

// LinkPeers creates a bidirectional link between two peer profiles
func (repo *ProfileRepository) LinkPeers(peerprofileId1 string, peerprofileId2 string, ruleName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// AddEventSchema peer2 to peer1
	peer2 := models.ChildProfile{
		ChildProfileId: peerprofileId2,
		RuleName:       ruleName,
	}
	updateProfile1 := bson.M{
		"$addToSet": bson.M{
			"profile_hierarchy.peer_profile_ids": peer2,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": peerprofileId1}, updateProfile1); err != nil {
		return fmt.Errorf("failed to update peer profile for %s: %w", peerprofileId1, err)
	}

	// AddEventSchema peer1 to peer2
	peer1 := models.ChildProfile{
		ChildProfileId: peerprofileId1, // ✅ Corrected
		RuleName:       ruleName,
	}
	updateProfile2 := bson.M{
		"$addToSet": bson.M{
			"profile_hierarchy.peer_profile_ids": peer1,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"profile_id": peerprofileId2}, updateProfile2); err != nil {
		return fmt.Errorf("failed to update peer profile for %s: %w", peerprofileId2, err)
	}

	return nil
}

func (repo *ProfileRepository) FindProfileByUserName(userID string) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"identity.user_name": userID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Profile not found is not an error
		}
		return nil, err
	}
	return &profile, nil
}

func (repo *ProfileRepository) AddChildProfile(parentProfile models.Profile, child models.ChildProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": parentProfile.ProfileId}
	update := bson.M{
		"$addToSet": bson.M{
			"profile_hierarchy.child_profile_ids": child,
		},
	}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to add child profile to parent %s: %w", parentProfile.ProfileId, err)
	}
	return nil
}

// UpsertSocialData applies PATCH updates to specific fields of PersonalityData
func (repo *ProfileRepository) UpsertSocialData(profileId string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": profileId}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data patched for user " + profileId)
	return nil
}

func (repo *ProfileRepository) UpsertIdentityAttributes(id string, updates bson.M) interface{} {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"profile_id": id}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data patched for user " + profileId)
	return nil
}
