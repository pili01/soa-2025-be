package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tours-service/internal/models"
	"tours-service/internal/services"

	"github.com/gorilla/mux"
)

type TourExecutionHandler struct {
	tourExecutionService *services.TourExecutionService
	authService          *services.AuthService
}

func NewTourExecutionHandler(tourExecutionService *services.TourExecutionService, authService *services.AuthService) *TourExecutionHandler {
	return &TourExecutionHandler{
		tourExecutionService: tourExecutionService,
		authService:          authService,
	}
}

func (h *TourExecutionHandler) StartTourExecution(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	tourIDStr := vars["tour_id"]
	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid or missing tour_id", http.StatusBadRequest)
		return
	}
	tourId, httpStatus, err := h.tourExecutionService.StartTour(userID, tourID)
	if err != nil {
		http.Error(w, err.Error(), httpStatus)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	response := map[string]string{"message": fmt.Sprintf("Tour successfully started: %d", tourId)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *TourExecutionHandler) AbortExecution(w http.ResponseWriter, r *http.Request) {
	userId, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	tourIdStr := vars["tour_id"]
	tourId, err := strconv.Atoi(tourIdStr)
	if err != nil {
		http.Error(w, "Invalid or missing tour_id", http.StatusBadRequest)
		return
	}
	httpStatus, err := h.tourExecutionService.AbortExecution(tourId, userId)
	if err != nil {
		http.Error(w, err.Error(), httpStatus)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Tour execution successfully aborted"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *TourExecutionHandler) CheckIsKeyPointReached(w http.ResponseWriter, r *http.Request) {
	userId, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	tourIdStr := vars["tour_id"]
	tourId, err := strconv.Atoi(tourIdStr)
	if err != nil {
		http.Error(w, "Invalid or missing tour_id", http.StatusBadRequest)
		return
	}
	var keyPoint *models.Keypoint
	var long, lat float64 = 15, 15
	httpStatus, keyPoint, finished, err := h.tourExecutionService.CheckIsKeyPointReached(tourId, userId, long, lat)
	if err != nil {
		http.Error(w, err.Error(), httpStatus)
		return
	}
	if keyPoint == nil {
		http.Error(w, "key point not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	resp := map[string]string{
		"message":      "Key point with name: " + keyPoint.Name + " reached successfully",
		"keyPointId":   strconv.Itoa(keyPoint.ID),
		"tourFinished": strconv.FormatBool(finished),
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
