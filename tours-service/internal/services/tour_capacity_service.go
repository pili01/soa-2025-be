package services

import (
	"fmt"
	"tours-service/internal/repositories"
)

type TourCapacityService struct {
	Repo *repositories.TourCapacityRepository
}

func NewTourCapacityService(repo *repositories.TourCapacityRepository) *TourCapacityService {
	return &TourCapacityService{Repo: repo}
}

func (s *TourCapacityService) InitCapacity(tourID, capacity int) error {
	if capacity < 0 {
		return fmt.Errorf("capacity must be >= 0")
	}
	return s.Repo.InitCapacity(tourID, capacity)
}

func (s *TourCapacityService) GetCapacity(tourID int) (int, int, error) {
	cap, err := s.Repo.GetCapacity(tourID)
	if err != nil {
		return 0, 0, err
	}
	return cap.Capacity, cap.AvailableSeats, nil
}

func (s *TourCapacityService) Consume(tourID, qty int) error {
	return s.Repo.ConsumeSeats(tourID, qty)
}

func (s *TourCapacityService) Release(tourID, qty int) error {
	return s.Repo.ReleaseSeats(tourID, qty)
}
