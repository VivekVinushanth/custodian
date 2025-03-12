package repositories

import (
	"context"
	"custodian/internal/models"
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

// AddOrUpdateAppContext replaces (PUT) or inserts a new AppContext inside Profile
func (repo *ProfileRepository) AddOrUpdateAppContext(permaID string, appContext models.AppContext) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID, "app_context.app_id": appContext.AppID}
	update := bson.M{
		"$set": bson.M{
			"app_context.$": appContext, // Replace existing app context
		},
	}

	opts := options.Update().SetUpsert(true) // Insert if not found
	res, err := repo.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to update app context: "+err.Error())
		return err
	}

	// If no documents were modified, insert a new array element
	if res.MatchedCount == 0 {
		update = bson.M{"$push": bson.M{"app_context": appContext}} // Append new app context
		_, err = repo.Collection.UpdateOne(ctx, bson.M{"perma_id": permaID}, update)
		if err != nil {
			//logger.LogMessage("ERROR", "Failed to insert new app context: "+err.Error())
			return err
		}
	}

	//logger.LogMessage("INFO", "App context updated for user "+permaID)
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
