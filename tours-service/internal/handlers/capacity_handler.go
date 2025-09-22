package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tours-service/internal/repositories"
	"tours-service/internal/services"

	"github.com/gorilla/mux"
)

type CapacityHandler struct {
	capacityService *services.TourCapacityService
	tourService     *services.TourService
	authService     *services.AuthService
}

func NewCapacityHandler(capSvc *services.TourCapacityService, tourSvc *services.TourService, authSvc *services.AuthService) *CapacityHandler {
	return &CapacityHandler{
		capacityService: capSvc,
		tourService:     tourSvc,
		authService:     authSvc,
	}
}

func (h *CapacityHandler) GetCapacity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourID, err := strconv.Atoi(vars["tourId"])
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	capacity, available, err := h.capacityService.GetCapacity(tourID)
	if err != nil {
		if errors.Is(err, repositories.ErrCapacityNotFound) {
			http.Error(w, "Capacity not found for this tour", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get capacity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tourId":         tourID,
		"capacity":       capacity,
		"availableSeats": available,
	})
}

func (h *CapacityHandler) InitOrUpdateCapacity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourID, err := strconv.Atoi(vars["tourId"])
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}
	if tour.AuthorID != userID {
		http.Error(w, "Only the tour author can set capacity", http.StatusForbidden)
		return
	}

	var payload struct {
		Capacity int `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Capacity < 0 {
		http.Error(w, "Capacity must be >= 0", http.StatusBadRequest)
		return
	}

	if err := h.capacityService.InitCapacity(tourID, payload.Capacity); err != nil {
		http.Error(w, "Failed to set capacity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	capacity, available, err := h.capacityService.GetCapacity(tourID)
	if err != nil {
		http.Error(w, "Capacity updated, but failed to read back: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Capacity updated successfully",
		"tourId":         tourID,
		"capacity":       capacity,
		"availableSeats": available,
	})
}

func (h *CapacityHandler) ConsumeSeats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourID, err := strconv.Atoi(vars["tourId"])
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Qty int `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Qty <= 0 {
		http.Error(w, "qty must be > 0", http.StatusBadRequest)
		return
	}

	err = h.capacityService.Consume(tourID, payload.Qty)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrCapacityNotFound):
			http.Error(w, "Capacity not found for this tour", http.StatusNotFound)
			return
		case errors.Is(err, repositories.ErrNotEnoughSeats):
			http.Error(w, "Not enough available seats", http.StatusConflict)
			return
		default:
			http.Error(w, "Failed to consume seats: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	capacity, available, err := h.capacityService.GetCapacity(tourID)
	if err != nil {
		http.Error(w, "Seats consumed, but failed to read back: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Seats consumed",
		"tourId":         tourID,
		"capacity":       capacity,
		"availableSeats": available,
	})
}

func (h *CapacityHandler) ReleaseSeats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourID, err := strconv.Atoi(vars["tourId"])
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Qty int `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Qty <= 0 {
		http.Error(w, "qty must be > 0", http.StatusBadRequest)
		return
	}

	err = h.capacityService.Release(tourID, payload.Qty)
	if err != nil {
		if errors.Is(err, repositories.ErrCapacityNotFound) {
			http.Error(w, "Capacity not found for this tour", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to release seats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	capacity, available, err := h.capacityService.GetCapacity(tourID)
	if err != nil {
		http.Error(w, "Seats released, but failed to read back: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Seats released",
		"tourId":         tourID,
		"capacity":       capacity,
		"availableSeats": available,
	})
}
