package services

import (
	"context"
	"errors"
	"fmt"
	"tours-service/internal/models"
	"tours-service/internal/repositories"
)

type TourService struct {
	TourRepo     *repositories.TourRepository
	KeypointRepo *repositories.KeypointRepository
	MapService   *MapService
}

func NewTourService(tourRepo *repositories.TourRepository, keypointRepo *repositories.KeypointRepository, mapService *MapService) *TourService {
	return &TourService{
		TourRepo:     tourRepo,
		KeypointRepo: keypointRepo,
		MapService:   mapService,
	}
}

func (s *TourService) CreateTour(tour *models.Tour, keypoints []*models.Keypoint) error {
	if len(keypoints) < 2 {
		return errors.New("a tour must have at least two keypoints")
	}

	totalStats := map[string]models.DistanceAndDuration{
		"driving-car":     {},
		"foot-walking":    {},
		"cycling-regular": {},
	}

	for i := 0; i < len(keypoints)-1; i++ {
		origin := *keypoints[i]
		dest := *keypoints[i+1]

		segmentStats, err := s.MapService.GetDistanceBetweenTwoKeypoints(context.Background(), origin, dest)
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

	err := s.TourRepo.CreateTour(tour)
	if err != nil {
		return fmt.Errorf("failed to create tour: %w", err)
	}

	for _, keypoint := range keypoints {
		keypoint.TourID = tour.ID
	}

	for _, keypoint := range keypoints {
		err := s.KeypointRepo.CreateKeypoint(keypoint)
		if err != nil {
			s.TourRepo.DeleteTour(tour.ID)
			s.KeypointRepo.DeleteKeypointsByTourID(tour.ID)
			return fmt.Errorf("failed to create keypoints, rolling back: %w", err)
		}
	}

	return nil
}

func (s *TourService) DeleteTour(tourID int) error {
	err := s.KeypointRepo.DeleteKeypointsByTourID(tourID)
	if err != nil {
		return fmt.Errorf("service failed to delete keypoints for tour %d: %w", tourID, err)
	}

	err = s.TourRepo.DeleteTour(tourID)
	if err != nil {
		return fmt.Errorf("service failed to delete tour %d: %w", tourID, err)
	}

	return nil
}

func (s *TourService) GetToursByAuthorID(authorID int) ([]models.Tour, error) {
	tours, err := s.TourRepo.GetToursByAuthorID(authorID)
	if err != nil {
		return nil, fmt.Errorf("service failed to get tours by author ID: %w", err)
	}
	return tours, nil
}

func (s *TourService) GetTourByID(tourID int) (*models.Tour, error) {
	tour, err := s.TourRepo.GetTourByID(tourID)
	if err != nil {
		return nil, fmt.Errorf("service failed to get tour by ID: %w", err)
	}
	return tour, nil
}

func (s *TourService) UpdateTour(tour *models.Tour) error {
	err := s.TourRepo.UpdateTour(tour)
	if err != nil {
		return fmt.Errorf("service failed to update tour: %w", err)
	}
	return nil
}


