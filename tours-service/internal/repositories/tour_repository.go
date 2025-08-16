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

type Counter struct {
	ID    string `bson:"_id"`
	Value int    `bson:"value"`
}

type TourRepository struct {
	Collection         *mongo.Collection
	CountersCollection *mongo.Collection
}

func NewTourRepository(db *mongo.Database) *TourRepository {
	return &TourRepository{
		Collection:         db.Collection("tours"),
		CountersCollection: db.Collection("counters"),
	}
}

func (r *TourRepository) getNextSequenceValue(sequenceName string) (int, error) {
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

func (r *TourRepository) CreateTour(tour *models.Tour) error {
	nextID, err := r.getNextSequenceValue("tour_id")
	if err != nil {
		return err
	}
	tour.ID = nextID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = r.Collection.InsertOne(ctx, tour)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("a tour with this name already exists for this author")
		}
		return fmt.Errorf("failed to create tour: %w", err)
	}

	return nil
}

func (r *TourRepository) GetToursByAuthorID(authorID int) ([]models.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"authorId": authorID}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find tours by author ID: %w", err)
	}
	defer cursor.Close(ctx)

	var tours []models.Tour
	if err = cursor.All(ctx, &tours); err != nil {
		return nil, fmt.Errorf("failed to decode tours: %w", err)
	}

	return tours, nil
}

func (r *TourRepository) GetTourByID(tourID int) (*models.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": tourID}

	var tour models.Tour
	err := r.Collection.FindOne(ctx, filter).Decode(&tour)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tour not found")
		}
		return nil, fmt.Errorf("failed to find tour: %w", err)
	}

	return &tour, nil
}

func (r *TourRepository) UpdateTour(tour *models.Tour) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": tour.ID}
	update := bson.M{"$set": tour}

	_, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update tour: %w", err)
	}
	return nil
}

func (r *TourRepository) DeleteTour(tourID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": tourID}
	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete tour: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("tour not found")
	}

	return nil
}