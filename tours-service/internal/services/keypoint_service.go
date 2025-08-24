package services

import (
	"tours-service/internal/models"
	"tours-service/internal/repositories"
)

type KeypointService struct {
	KeypointRepo *repositories.KeypointRepository
}

func NewKeypointService(keypointRepo *repositories.KeypointRepository) *KeypointService {
	return &KeypointService{KeypointRepo: keypointRepo}
}

func (s *KeypointService) CreateKeypoint(keypoint *models.Keypoint) error {
	return s.KeypointRepo.CreateKeypoint(keypoint)
}

func (s *KeypointService) GetKeypointsByTourID(tourID int) ([]models.Keypoint, error) {
	return s.KeypointRepo.GetKeypointsByTourID(tourID)
}

func (s *KeypointService) GetKeypointByID(keypointID int) (*models.Keypoint, error) {
	return s.KeypointRepo.GetKeypointByID(keypointID)
}

func (s *KeypointService) UpdateKeypoint(keypoint *models.Keypoint) error {
	return s.KeypointRepo.UpdateKeypoint(keypoint)
}

func (s *KeypointService) DeleteKeypoint(keypointID int) error {
	return s.KeypointRepo.DeleteKeypoint(keypointID)
}

func (s *KeypointService) DeleteKeypointsByTourID(tourID int) error {
	return s.KeypointRepo.DeleteKeypointsByTourID(tourID)
}

func (s *KeypointService) GetNextUncompletedKeyPointByTourId(tourID int, completedPoints []int) (*models.Keypoint, error) {
	return s.KeypointRepo.GetUncompletedKeyPointsByTourId(tourID, completedPoints)
}
