package service

import (
	"custodian/internal/models"
	"custodian/internal/pkg"
	"custodian/internal/repository"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// UserInput represents input data for creating a profile
type UserInput struct {
	OriginCountry string                  `json:"origin_country" binding:"required"`
	UserIds       []string                `json:"user_ids,omitempty"`
	Identity      *models.IdentityData    `json:"identity,omitempty"`
	Personality   *models.PersonalityData `json:"personality,omitempty"`
	AppContext    []models.AppContext     `json:"app_context,omitempty"`
}

// CreateProfile handles the business logic of creating a new user profile
func CreateProfile(input UserInput) (*models.Profile, error) {

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

	newProfile := models.Profile{
		PermaID:       uuid.NewString(),
		OriginCountry: input.OriginCountry,
		UserIds:       input.UserIds,
		Identity:      input.Identity,
		Personality:   input.Personality,
		AppContext:    input.AppContext,
	}

	//logger.LogMessage("AUDIT", "Created new profile with PermaID: "+newProfile.PermaID)

	// validate input data

	// save to database
	_, err := profileRepo.InsertProfile(newProfile)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to save profile to MongoDB")
		return nil, err
	}
	return &newProfile, nil
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
