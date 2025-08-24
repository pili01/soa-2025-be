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

	_, err := r.TourExCollection.InsertOne(ctx, tourExecution)
	if err != nil {
		fmt.Printf("Error creating tour execution: %v\n", err)
		return fmt.Errorf("failed to create tour execution: %w", err)
	}
	return nil
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
