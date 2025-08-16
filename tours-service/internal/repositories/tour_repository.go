// tours-service/internal/repositories/tour_repository.go
package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"
	"net/http"
	"encoding/json"
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

func (r *TourRepository) CreateTourWithKeypoints(tour *models.Tour, keypoints []*models.Keypoint, keypointRepo *KeypointRepository) error {
	if len(keypoints) < 2 {
		return errors.New("a tour must have at least two keypoints")
	}

	totalStats := map[string]models.DistanceAndDuration{
		"driving-car": {},
		"foot-walking": {},
		"cycling-regular": {},
	}

	for i := 0; i < len(keypoints)-1; i++ {
		origin := *keypoints[i]
		dest := *keypoints[i+1]
		
		segmentStats, err := r.getDistanceBetweenTwoKeypoints(context.Background(), origin, dest)
		if err != nil {
			fmt.Printf("Warning: Failed to get segment distance for tour creation: %v\n", err)
			break 
		}

		for profile, stats := range segmentStats {
			currentStats := totalStats[profile]
			currentStats.Distance += stats.Distance
			currentStats.Duration += stats.Duration
			totalStats[profile] = currentStats
		}
	}

	tour.DrivingStats = totalStats["driving-car"]
	tour.WalkingStats = totalStats["foot-walking"]
	tour.CyclingStats = totalStats["cycling-regular"]
	
	err := r.CreateTour(tour)
	if err != nil {
		return fmt.Errorf("failed to create tour: %w", err)
	}

	for _, keypoint := range keypoints {
		keypoint.TourID = tour.ID
	}

	for _, keypoint := range keypoints {
		err := keypointRepo.CreateKeypoint(keypoint)
		if err != nil {
			r.DeleteTour(tour.ID)
			keypointRepo.DeleteKeypointsByTourID(tour.ID)
			return fmt.Errorf("failed to create keypoints, rolling back: %w", err)
		}
	}

	return nil
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
		fmt.Printf("MongoDB InsertOne error: %v\n", err)

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

func (r *TourRepository) getDistanceBetweenTwoKeypoints(ctx context.Context, origin, dest models.Keypoint) (map[string]models.DistanceAndDuration, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("http://map-service:3000/api/getdistances?originLat=%f&originLng=%f&destLat=%f&destLng=%f",
		origin.Latitude, origin.Longitude, dest.Latitude, dest.Longitude)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call map-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("map-service returned non-OK status: %d", resp.StatusCode)
	}

	var results map[string]models.DistanceAndDuration
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return results, nil
}