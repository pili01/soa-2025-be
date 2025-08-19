package handlers

import (
	"fmt"
	"net/http"
	"strconv"
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
	tourId, err := h.tourExecutionService.StartTour(userID, tourID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Tour successfully started: %d", tourId)))
}
