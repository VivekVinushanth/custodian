package repositories

import (
	"context"
	"custodian/internal/models"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	filter := bson.M{"perma_id": profile.PermaID}
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
	//logger.LogMessage("INFO", "Profile inserted with ID: "+result.InsertedID.(string))
	return result, nil
}

// GetProfile retrieves a profile by `perma_id`
func (repo *ProfileRepository) GetProfile(permaID string) (*models.Profile, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"perma_id": permaID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//logger.LogMessage("INFO", "Profile not found for PermaID: "+permaID)
			return nil, nil // Return `nil` instead of error
		}
		//logger.LogMessage("ERROR", "Error finding profile: "+err.Error())
		return nil, err
	}

	//logger.LogMessage("INFO", "Profile retrieved for PermaID: "+permaID)
	return &profile, nil
}

// FindProfileByID retrieves a profile by `perma_id`
func (repo *ProfileRepository) FindProfileByID(permaID string) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"perma_id": permaID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Profile not found is not an error
		}
		log.Print("Error finding profile==123: " + err.Error())
		return nil, err
	}
	return &profile, nil
}

// DeleteProfile removes a profile from MongoDB using `perma_id`
func (repo *ProfileRepository) DeleteProfile(permaID string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}

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
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": parentID}, update)
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
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": profileID}, update)
	return err
}

func (repo *ProfileRepository) AddOrUpdateAppContext(permaID string, newAppCtx models.AppContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Fetch full profile
	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"perma_id": permaID}).Decode(&profile)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	// Step 2: Check if app_id already exists
	found := false
	var updatedAppContexts []models.AppContext

	for _, existingApp := range profile.AppContext {
		if existingApp.AppID == newAppCtx.AppID {
			// Merge devices
			existingApp.Devices = mergeDeviceLists(existingApp.Devices, newAppCtx.Devices)
			updatedAppContexts = append(updatedAppContexts, existingApp)
			found = true
		} else {
			updatedAppContexts = append(updatedAppContexts, existingApp)
		}
	}

	if !found {
		updatedAppContexts = append(updatedAppContexts, newAppCtx)
	}

	// Step 3: Replace entire app_context array
	update := bson.M{"$set": bson.M{"app_context": updatedAppContexts}}
	_, err = repo.Collection.UpdateOne(ctx, bson.M{"perma_id": permaID}, update)
	if err != nil {
		return fmt.Errorf("failed to update app_context: %w", err)
	}

	return nil
}

func mergeDeviceLists(existing, incoming []models.AppContextDevices) []models.AppContextDevices {
	deviceMap := make(map[string]models.AppContextDevices)

	for _, dev := range existing {
		deviceMap[dev.DeviceID] = dev
	}
	for _, dev := range incoming {
		deviceMap[dev.DeviceID] = dev // overwrite if already exists
	}

	var merged []models.AppContextDevices
	for _, dev := range deviceMap {
		merged = append(merged, dev)
	}
	return merged
}

// PatchAppContext updates only specific fields inside the app context
func (repo *ProfileRepository) PatchAppContext(permaID, appID string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID, "app_context.app_id": appID}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch app context: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "App context patched for user " + permaID)
	return nil
}

// GetAppContext retrieves a specific app context inside a profile
func (repo *ProfileRepository) GetAppContext(permaID, appID string) (*models.AppContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID, "app_context.app_id": appID}
	projection := bson.M{"app_context.$": 1}

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&profile)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(profile.AppContext) > 0 {
		return &profile.AppContext[0], nil
	}

	return nil, nil
}

// GetListOfAppContext retrieves all app contexts for a user
func (repo *ProfileRepository) GetListOfAppContext(permaID string) ([]models.AppContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"permaid": permaID}
	projection := bson.M{"appcontext": 1}

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&profile)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return profile.AppContext, nil
}

// AddOrUpdatePersonalityData replaces (PUT) the personality data inside Profile
func (repo *ProfileRepository) AddOrUpdatePersonalityData(permaID string, personalityData models.PersonalityData) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	update := bson.M{"$set": bson.M{"personality": personalityData}}

	opts := options.Update().SetUpsert(true) // Insert if not found
	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data updated for user "+permaID)
	return nil
}

// AddOrUpdateIdentityData replaces (PUT) the personality data inside Profile
func (repo *ProfileRepository) AddOrUpdateIdentityData(permaID string, personalityData models.IdentityData) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	update := bson.M{"$set": bson.M{"identity": personalityData}}

	opts := options.Update().SetUpsert(true) // Insert if not found
	_, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data updated for user "+permaID)
	return nil
}

// UpdatePersonalityData applies PATCH updates to specific fields of PersonalityData
func (repo *ProfileRepository) UpdatePersonalityData(permaID string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data patched for user " + permaID)
	return nil
}

// UpdatePreferenceData applies PATCH updates to specific fields of PersonalityData
func (repo *ProfileRepository) UpdatePreferenceData(permaID string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	_, err := repo.Collection.UpdateOne(ctx, filter, updates)
	return err
}

// UpdateIdentityData adds or updates fields in IdentityData without overwriting existing non-empty values
func (repo *ProfileRepository) UpdateIdentityData(permaID string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Ensure identity exists
	identityInit := bson.M{
		"$setOnInsert": bson.M{
			"identity": bson.M{},
		},
	}
	_, _ = repo.Collection.UpdateOne(ctx, bson.M{"perma_id": permaID, "identity": bson.M{"$exists": false}}, identityInit)

	// Step 2: Prepare full update
	setUpdates := bson.M{}
	for k, v := range updates {
		setUpdates["identity."+k] = v
	}

	update := bson.M{
		"$set": setUpdates,
	}

	// Step 3: Apply
	_, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": permaID}, update)
	return err
}

// GetPersonalityProfileData retrieves the personality data from a user's profile
func (repo *ProfileRepository) GetPersonalityProfileData(permaID string) (*models.PersonalityData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	projection := bson.M{"personality": 1}

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&profile)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return profile.Personality, nil
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
	excludedIDs = append(excludedIDs, currentProfile.PermaID)

	// Fetch only master profiles excluding the parent of the current profile
	filter := bson.M{
		"profile_hierarchy.is_master": true,
		"perma_id": bson.M{
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
func (repo *ProfileRepository) AddOrUpdateUserIds(permaID string, newUserIds []string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Fetch the existing user_ids
	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"perma_id": permaID}).Decode(&profile)
	if err != nil && err != mongo.ErrNoDocuments {
		//logger.LogMessage("ERROR", "Failed to fetch profile for user_ids merge: "+err.Error())
		return err
	}

	// Step 2: Perform the update
	filter := bson.M{"perma_id": permaID}
	update := bson.M{"$set": bson.M{"userids": newUserIds}}

	opts := options.Update().SetUpsert(true)
	_, err = repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update user_ids: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "User IDs updated for user "+permaID)
	return nil
}

func (repo *ProfileRepository) UpdateParent(master models.Profile, newProfile models.Profile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateProfile := bson.M{
		"$set": bson.M{
			"profile_hierarchy.parent_profile_id": master.PermaID,
			"profile_hierarchy.is_master":         false,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": newProfile.PermaID}, updateProfile); err != nil {
		return fmt.Errorf("failed to update profile %s: %w", newProfile.PermaID, err)
	}

	return nil
}

// LinkPeers creates a bidirectional link between two peer profiles
func (repo *ProfileRepository) LinkPeers(peerPermaId1 string, peerPermaId2 string, ruleName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add peer2 to peer1
	peer2 := models.PeerProfile{
		PeerProfileId: peerPermaId2,
		RuleName:      ruleName,
	}
	updateProfile1 := bson.M{
		"$addToSet": bson.M{
			"profile_hierarchy.peer_profile_ids": peer2,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": peerPermaId1}, updateProfile1); err != nil {
		return fmt.Errorf("failed to update peer profile for %s: %w", peerPermaId1, err)
	}

	// Add peer1 to peer2
	peer1 := models.PeerProfile{
		PeerProfileId: peerPermaId1, // ✅ Corrected
		RuleName:      ruleName,
	}
	updateProfile2 := bson.M{
		"$addToSet": bson.M{
			"profile_hierarchy.peer_profile_ids": peer1,
		},
	}
	if _, err := repo.Collection.UpdateOne(ctx, bson.M{"perma_id": peerPermaId2}, updateProfile2); err != nil {
		return fmt.Errorf("failed to update peer profile for %s: %w", peerPermaId2, err)
	}

	return nil
}
