package repositories

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tours-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TourExecutionRepository struct {
	TourExCollection *mongo.Collection
}

func NewTourExecutionRepository(db *mongo.Database) *TourExecutionRepository {
	return &TourExecutionRepository{
		TourExCollection: db.Collection("tourExecution"),
	}
}

func (r *TourExecutionRepository) FindByUserAndTourId(userId, tourId int) (*models.TourExecution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tour_id": tourId, "user_id": userId}
	var tourEx models.TourExecution
	err := r.TourExCollection.FindOne(ctx, filter).Decode(&tourEx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tour execution not found")
		}
		return nil, errors.New("failed to find tour execution")
	}
	return &tourEx, nil
}

func (r *TourExecutionRepository) CreateExecution(tourExecution *models.TourExecution) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// generate an auto-increment integer id for the execution and set it
	seq, err := r.getNextSequence("tourExecution")
	if err != nil {
		fmt.Printf("Error generating tour execution id: %v\n", err)
		return fmt.Errorf("failed to generate id for tour execution: %w", err)
	}
	tourExecution.ID = seq

	_, err = r.TourExCollection.InsertOne(ctx, tourExecution)
	if err != nil {
		fmt.Printf("Error creating tour execution: %v\n", err)
		return fmt.Errorf("failed to create tour execution: %w", err)
	}
	return nil
}

// getNextSequence atomically increments and returns a sequence number stored in a counters collection
func (r *TourExecutionRepository) getNextSequence(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	countersColl := r.TourExCollection.Database().Collection("counters")
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err := countersColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// shouldn't happen because of upsert, but handle defensively
			return 0, nil
		}
		return 0, err
	}
	return result.Seq, nil
}

func (r *TourExecutionRepository) AbortExecution(tourExecution *models.TourExecution) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": tourExecution.ID}
	update := bson.M{
		"status":        tourExecution.Status,
		"ended_at":      tourExecution.EndedAt,
		"last_activity": tourExecution.LastActivity,
	}
	_, err := r.TourExCollection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		fmt.Printf("Error aborting tour execution: %v\n", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (r *TourExecutionRepository) CompleateKeyPoint(executionId int, keyPoints []models.FinishedKeyPoint, now *time.Time) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": executionId}
	update := bson.M{
		"finished_keypoints": keyPoints,
		"last_activity":      now,
	}

	_, err := r.TourExCollection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		fmt.Printf("error adding key point %v\n", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (r *TourExecutionRepository) CompleateTour(executionId int, now *time.Time) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": executionId}
	update := bson.M{
		"ended_at":      now,
		"last_activity": now,
		"status":        models.ExecutionStatusCompleted,
	}

	_, err := r.TourExCollection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		fmt.Printf("error completing tour %v\n", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
