package services

import (
	"fmt"
	"math"
	"net/http"
	"time"
	"tours-service/internal/models"
	"tours-service/internal/repositories"

	"go.mongodb.org/mongo-driver/mongo"
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

func (tes *TourExecutionService) CheckIsKeyPointReached(tourId, userId int, long, lat float64) (int, *models.Keypoint, bool, error) {
	tourExecution, err := tes.TourExecutionRepository.FindByUserAndTourId(userId, tourId)
	if err != nil {
		return http.StatusInternalServerError, nil, false, fmt.Errorf("database error: %w", err)
	}
	if tourExecution == nil {
		return http.StatusNotFound, nil, false, fmt.Errorf("no tour execution found")
	}
	if tourExecution.Status != models.ExecutionStatusInProgress {
		return http.StatusNotFound, nil, false, fmt.Errorf("tour execution not in progress")
	}
	if userId != tourExecution.UserID {
		return http.StatusUnauthorized, nil, false, fmt.Errorf("user not authorized to change this tour execution")
	}
	var completedKeyPointsIds []int
	for _, kp := range tourExecution.FinishedKeypoints {
		completedKeyPointsIds = append(completedKeyPointsIds, kp.KeypointID)
	}
	keyPoint, err := tes.KeyPointsService.KeypointRepo.GetUncompletedKeyPointsByTourId(tourId, completedKeyPointsIds)
	if err != nil {
		return http.StatusNotFound, nil, false, err
	}
	if keyPoint == nil {
		return http.StatusNotFound, nil, false, fmt.Errorf("no key point found")
	}
	fmt.Println("Next uncompleted key point is " + keyPoint.Name)

	if !tes.checkDistance(long, lat, keyPoint.Longitude, keyPoint.Latitude, 20.0) {
		return http.StatusOK, nil, false, fmt.Errorf("you are not close enough to complete key point")
	}

	now := time.Now()
	newFinishedKp := &models.FinishedKeyPoint{KeypointID: keyPoint.ID, CompletedAt: &now}

	completedKeyPoints := append(tourExecution.FinishedKeypoints, *newFinishedKp)
	_, err = tes.TourExecutionRepository.CompleateKeyPoint(tourExecution.ID, completedKeyPoints, &now)
	if err != nil {
		return http.StatusInternalServerError, nil, false, fmt.Errorf("unable to update execution")
	}

	tourExecution, err = tes.TourExecutionRepository.FindByUserAndTourId(userId, tourId)
	if err != nil || tourExecution == nil {
		return http.StatusInternalServerError, nil, false, fmt.Errorf("database error: %w", err)
	}

	completedKeyPointsIds = []int{}
	for _, kp := range tourExecution.FinishedKeypoints {
		completedKeyPointsIds = append(completedKeyPointsIds, kp.KeypointID)
	}
	_, err = tes.KeyPointsService.KeypointRepo.GetUncompletedKeyPointsByTourId(tourId, completedKeyPointsIds)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return http.StatusNotFound, nil, false, err
		} else {
			fmt.Println("Tour completed successfully")
			_, err = tes.TourExecutionRepository.CompleateTour(tourExecution.ID, &now)
			if err != nil {
				return http.StatusInternalServerError, nil, false, err
			}
			return http.StatusOK, keyPoint, true, nil
		}
	}

	return http.StatusOK, keyPoint, false, nil
}

func (tes *TourExecutionService) checkDistance(first_lon, first_lat, second_lon, second_lat, radius float64) bool {
	const EarthRadius = 6371000

	lat1Rad := first_lat * math.Pi / 180
	lon1Rad := first_lon * math.Pi / 180
	lat2Rad := second_lat * math.Pi / 180
	lon2Rad := second_lon * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := EarthRadius * c // u metrima

	return distance <= radius
}
