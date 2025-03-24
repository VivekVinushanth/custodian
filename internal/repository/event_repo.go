package repositories

import (
	"context"
	"custodian/internal/models"
	//"custodian/internal/pkg"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
)

// EventRepository handles MongoDB operations for user events
type EventRepository struct {
	Collection *mongo.Collection
}

// NewEventRepository initializes a repository for `events` collection
func NewEventRepository(db *mongo.Database, collectionName string) *EventRepository {
	return &EventRepository{
		Collection: db.Collection(collectionName),
	}
}

// AddEvent inserts a single event into MongoDB
func (repo *EventRepository) AddEvent(event models.Event) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.Collection.InsertOne(ctx, event)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to insert event: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Event inserted successfully for user "+event.PermaID)
	return nil
}

// AddEvents inserts multiple events in bulk
func (repo *EventRepository) AddEvents(events []models.Event) error {
	//logger := pkg.GetLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var docs []interface{}
	for _, event := range events {
		docs = append(docs, event)
	}

	_, err := repo.Collection.InsertMany(ctx, docs)
	if err != nil {
		//logger.LogMessage("ERROR", "Failed to insert multiple events: "+err.Error())
		return err
	}

	//logger.LogMessage("INFO", "Batch events inserted successfully")
	return nil
}

// GetUserEvent fetches a specific event by `event_id`
func (repo *EventRepository) GetUserEvent(permaID, eventID string) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID, "event_id": eventID}
	var event models.Event

	err := repo.Collection.FindOne(ctx, filter).Decode(&event)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetUserEvents fetches all events for a user
func (repo *EventRepository) GetUserEvents(permaID string) ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"perma_id": permaID}
	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

func (repo *EventRepository) DeleteEvent(permaID, eventID string) error {
	filter := bson.M{"perma_id": permaID, "event_id": eventID}
	_, err := repo.Collection.DeleteOne(context.TODO(), filter)
	return err
}

func (repo *EventRepository) DeleteEventsByPermaID(permaID string) error {
	_, err := repo.Collection.DeleteMany(context.TODO(), bson.M{"perma_id": permaID})
	return err
}

func (repo *EventRepository) DeleteEventsByAppID(permaID, appID string) error {
	filter := bson.M{"perma_id": permaID, "app_id": appID}
	_, err := repo.Collection.DeleteMany(context.TODO(), filter)
	return err
}
