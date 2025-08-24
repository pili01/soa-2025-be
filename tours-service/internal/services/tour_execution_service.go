package services

import (
	"fmt"
	"net/http"
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

func (tes *TourExecutionService) StartTour(userId int, tourId int) (*int, int, error) {
	tourExecution, err := tes.TourExecutionRepository.FindByUserAndTourId(userId, tourId)
	if tourExecution != nil || err == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("tour already started")
	}
	tour, err := tes.TourService.GetTourByID(tourId)
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("unable to find tour with id %d", tourId)
	}
	if tour.Status != models.StatusArchived && tour.Status != models.StatusPublished {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid tour status: %s", tour.Status)
	}
	now := time.Now()
	newTourExecution := models.TourExecution{TourID: tourId, UserID: userId, StartedAt: &now, LastActivity: &now, Status: models.ExecutionStatusInProgress}
	error := tes.TourExecutionRepository.CreateExecution(&newTourExecution)
	if error != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create tour execution")
	}
	return &newTourExecution.ID, http.StatusCreated, nil
}

func (tes *TourExecutionService) AbortExecution(tourId, userId int) (int, error) {
	tourExecution, err := tes.TourExecutionRepository.FindByUserAndTourId(userId, tourId)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error finding tour execution: %w", err)
	}
	if tourExecution == nil {
		return http.StatusNotFound, fmt.Errorf("tour execution not found")
	}
	if tourExecution.Status != models.ExecutionStatusInProgress {
		return http.StatusBadRequest, fmt.Errorf("invalid tour execution status: %s", tourExecution.Status)
	}
	if tourExecution.UserID != userId {
		return http.StatusUnauthorized, fmt.Errorf("user not authorized to abort this tour execution")
	}
	endTime := time.Now()
	tourExecution.EndedAt = &endTime
	tourExecution.Status = models.ExecutionStatusAborted
	tourExecution.LastActivity = &endTime
	return tes.TourExecutionRepository.AbortExecution(tourExecution)
}
