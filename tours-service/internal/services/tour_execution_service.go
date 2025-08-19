package services

import (
	"fmt"
	"time"
	"tours-service/internal/models"
	"tours-service/internal/repositories"
)

type TourExecutionService struct {
	TourExecutionRepository *repositories.TourExecutionRepository
	TourService             *TourService
	KeyPointsService        *KeypointService
}

func NewTourExecutionService(tourExRepository *repositories.TourExecutionRepository, tourService *TourService, keyPointService *KeypointService) *TourExecutionService {
	return &TourExecutionService{
		TourExecutionRepository: tourExRepository,
		TourService:             tourService,
		KeyPointsService:        keyPointService,
	}
}

func (tes *TourExecutionService) StartTour(userId int, tourId int) (*int, error) {
	tourExecution, err := tes.TourExecutionRepository.FindByUserAndTourId(userId, tourId)
	if tourExecution != nil || err == nil {
		return nil, fmt.Errorf("tour already started")
	}
	tour, err := tes.TourService.GetTourByID(tourId)
	if err != nil {
		return nil, fmt.Errorf("unable to find tour with id %d", tourId)
	}
	if tour.Status != models.StatusArchived && tour.Status != models.StatusPublished {
		return nil, fmt.Errorf("invalid tour status: %s", tour.Status)
	}
	now := time.Now()
	newTourExecution := models.TourExecution{TourID: tourId, UserID: userId, StartedAt: &now, LastActivity: &now, Status: models.ExecutionStatusInProgress}
	error := tes.TourExecutionRepository.CreateExecution(&newTourExecution)
	if error != nil {
		return nil, fmt.Errorf("failed to create tour execution")
	}
	return &newTourExecution.ID, nil
}
