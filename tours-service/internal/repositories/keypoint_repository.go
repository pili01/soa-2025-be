package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"tours-service/internal/models" 
)

type KeypointRepository struct {
	Collection        *mongo.Collection
	CountersCollection *mongo.Collection
}

func NewKeypointRepository(db *mongo.Database) *KeypointRepository {
	return &KeypointRepository{
		Collection:        db.Collection("keypoints"),
		CountersCollection: db.Collection("counters"),
	}
}

func (r *KeypointRepository) getNextSequenceValue(sequenceName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var counter Counter
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{"_id": sequenceName}
	update := bson.M{"$inc": bson.M{"value": 1}}

	err := r.CountersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence value: %w", err)
	}

	return counter.Value, nil
}

func (r *KeypointRepository) CreateKeypoint(keypoint *models.Keypoint) error {
	nextID, err := r.getNextSequenceValue("keypoint_id")
	if err != nil {
		return err 
	}
	keypoint.ID = nextID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = r.Collection.InsertOne(ctx, keypoint)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("a keypoint with this name already exists for this tour")
		}
		return fmt.Errorf("failed to create keypoint: %w", err)
	}

	return nil
}

func (r *KeypointRepository) GetKeypointsByTourID(tourID int) ([]models.Keypoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tourId": tourID}
	opts := options.Find().SetSort(bson.M{"ordinal": 1}) 

	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find keypoints by tour ID: %w", err)
	}
	defer cursor.Close(ctx)

	var keypoints []models.Keypoint
	if err = cursor.All(ctx, &keypoints); err != nil {
		return nil, fmt.Errorf("failed to decode keypoints: %w", err)
	}

	return keypoints, nil
}

func (r *KeypointRepository) GetKeypointByID(keypointID int) (*models.Keypoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": keypointID}

	var keypoint models.Keypoint
	err := r.Collection.FindOne(ctx, filter).Decode(&keypoint)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("keypoint not found")
		}
		return nil, fmt.Errorf("failed to find keypoint: %w", err)
	}

	return &keypoint, nil
}

func (r *KeypointRepository) UpdateKeypoint(keypoint *models.Keypoint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": keypoint.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        keypoint.Name,
			"description": keypoint.Description,
			"imageUrl":    keypoint.ImageURL,
			"latitude":    keypoint.Latitude,
			"longitude":   keypoint.Longitude,
			"ordinal":     keypoint.Ordinal,
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update keypoint: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("keypoint not found")
	}

	return nil
}

func (r *KeypointRepository) DeleteKeypoint(keypointID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": keypointID}

	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete keypoint: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("keypoint not found")
	}

	return nil
}

func (r *KeypointRepository) DeleteKeypointsByTourID(tourID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tourId": tourID}

	_, err := r.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete keypoints by tour ID: %w", err)
	}

	return nil
}


