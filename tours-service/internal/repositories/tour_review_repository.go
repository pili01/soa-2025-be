package repositories

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"tours-service/internal/models"
)

type TourReviewRepository struct {
	Collection         *mongo.Collection
	CountersCollection *mongo.Collection
}

func NewTourReviewRepository(db *mongo.Database) *TourReviewRepository {
	return &TourReviewRepository{
		Collection:         db.Collection("tour_reviews"),
		CountersCollection: db.Collection("counters"),
	}
}

func (r *TourReviewRepository) getNextSequenceValue(sequenceName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var counter Counter
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{"_id": sequenceName}
	update := bson.M{"$inc": bson.M{"value": 1}}

	err := r.CountersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence value for %s: %w", sequenceName, err)
	}

	return counter.Value, nil
}

func (r *TourReviewRepository) CreateTourReview(review *models.TourReview) error {
	nextID, err := r.getNextSequenceValue("tour_review_id")
	if err != nil {
		return err
	}
	review.ID = nextID
	review.CommentDate = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = r.Collection.InsertOne(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to create tour review: %w", err)
	}

	return nil
}

func (r *TourReviewRepository) GetReviewsByTourID(tourID int) ([]models.TourReview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tourId": tourID}

	opts := options.Find().SetSort(bson.M{"commentDate": -1})

	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by tour ID: %w", err)
	}
	defer cursor.Close(ctx)

	var reviews []models.TourReview
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, fmt.Errorf("failed to decode tour reviews: %w", err)
	}

	return reviews, nil
}