package repositories

import (
	"context"
	"custodian/internal/models"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (repo *ProfileRepository) InsertProfile(profile models.Profile) (*mongo.InsertOneResult, error) {
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

// FindProfileByID retrieves a profile by `perma_id`
func (repo *ProfileRepository) FindProfileByID(permaID string) (*models.Profile, error) {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"permaid": permaID}).Decode(&profile)
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

// DeleteProfile removes a profile from MongoDB using `perma_id`
func (repo *ProfileRepository) DeleteProfile(permaID string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"permaid": permaID}

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

func (repo *ProfileRepository) AddOrUpdateAppContext(permaID string, appContext models.AppContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Try to update existing app_context by app_id using positional $
	filter := bson.M{
		"perma_id":           permaID,
		"app_context.app_id": appContext.AppID, // Required for positional operator
	}
	update := bson.M{
		"$set": bson.M{
			"app_context.$": appContext, // Update matching array element
		},
	}

	res, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update app_context: %w", err)
	}

	// Step 2: If no match, push the new appContext into the array
	if res.MatchedCount == 0 {
		pushFilter := bson.M{"perma_id": permaID}
		push := bson.M{
			"$push": bson.M{
				"app_context": appContext,
			},
		}
		_, err := repo.Collection.UpdateOne(ctx, pushFilter, push)
		if err != nil {
			return fmt.Errorf("failed to insert new app_context: %w", err)
		}
	}

	return nil
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

	filter := bson.M{"permaid": permaID}
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

	filter := bson.M{"permaid": permaID}
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

	filter := bson.M{"permaid": permaID}
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

	filter := bson.M{"permaid": permaID}
	_, err := repo.Collection.UpdateOne(ctx, filter, updates)
	return err
}

// UpdateIdentityData applies PATCH updates to specific fields of IdentityData
func (repo *ProfileRepository) UpdateIdentityData(permaID string, updates bson.M) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"permaid": permaID}
	update := bson.M{"$set": updates}

	_, err := repo.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to patch personality data: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Personality data patched for user " + permaID)
	return nil
}

// GetPersonalityProfileData retrieves the personality data from a user's profile
func (repo *ProfileRepository) GetPersonalityProfileData(permaID string) (*models.PersonalityData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"permaid": permaID}
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

	// 🔹 Query MongoDB for all profiles
	cursor, err := repo.Collection.Find(ctx, bson.M{})
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

// AddOrUpdateUserIds merges and updates the user_ids list inside the profile
func (repo *ProfileRepository) AddOrUpdateUserIds(permaID string, newUserIds []string) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Fetch the existing user_ids
	var profile models.Profile
	err := repo.Collection.FindOne(ctx, bson.M{"permaid": permaID}).Decode(&profile)
	if err != nil && err != mongo.ErrNoDocuments {
		//logger.LogMessage("ERROR", "Failed to fetch profile for user_ids merge: "+err.Error())
		return err
	}

	// Step 2: Perform the update
	filter := bson.M{"permaid": permaID}
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
