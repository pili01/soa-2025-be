package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tours-service/internal/models"
	"tours-service/internal/services" 
	"github.com/gorilla/mux"
)

type KeypointHandler struct {
	keypointService *services.KeypointService
	tourService     *services.TourService
	authService     *services.AuthService 
}

func NewKeypointHandler(keypointService *services.KeypointService, tourService *services.TourService, authService *services.AuthService) *KeypointHandler {
	return &KeypointHandler{
		keypointService: keypointService,
		tourService:     tourService,
		authService:     authService,
	}
}

func (h *KeypointHandler) CreateKeypoint(w http.ResponseWriter, r *http.Request) {
	var req models.CreateKeypointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tour, err := h.tourService.GetTourByID(req.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != userID {
		http.Error(w, "Only tour author can add keypoints", http.StatusForbidden)
		return
	}

	keypoint := &models.Keypoint{
		TourID:      req.TourID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Ordinal:     req.Ordinal,
	}

	err = h.keypointService.CreateKeypoint(keypoint)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create keypoint", http.StatusInternalServerError)
		}
		return
	}

	_ = h.tourService.RecalculateTourLength(r.Context(), keypoint.TourID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keypoint)
}

func (h *KeypointHandler) GetKeypointsByTourID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	keypoints, err := h.keypointService.GetKeypointsByTourID(tourID)
	if err != nil {
		http.Error(w, "Failed to retrieve keypoints", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keypoints)
}

func (h *KeypointHandler) GetKeypointByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keypointIDStr := vars["keypointId"]

	keypointID, err := strconv.Atoi(keypointIDStr)
	if err != nil {
		http.Error(w, "Invalid keypoint ID", http.StatusBadRequest)
		return
	}

	keypoint, err := h.keypointService.GetKeypointByID(keypointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve keypoint", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keypoint)
}

func (h *KeypointHandler) UpdateKeypoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keypointIDStr := vars["keypointId"]

	keypointID, err := strconv.Atoi(keypointIDStr)
	if err != nil {
		http.Error(w, "Invalid keypoint ID", http.StatusBadRequest)
		return
	}

	var keypointUpdate models.CreateKeypointRequest
	if err := json.NewDecoder(r.Body).Decode(&keypointUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingKeypoint, err := h.keypointService.GetKeypointByID(keypointID)
	if err != nil {
		http.Error(w, "Keypoint not found", http.StatusNotFound)
		return
	}
	
	tour, err := h.tourService.GetTourByID(existingKeypoint.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != userID {
		http.Error(w, "Only tour author can modify keypoints", http.StatusForbidden)
		return
	}

	keypointToUpdate := &models.Keypoint{
		ID:          keypointID,
		TourID:      existingKeypoint.TourID,
		Name:        keypointUpdate.Name,
		Description: keypointUpdate.Description,
		ImageURL:    keypointUpdate.ImageURL,
		Latitude:    keypointUpdate.Latitude,
		Longitude:   keypointUpdate.Longitude,
		Ordinal:     keypointUpdate.Ordinal,
	}

	err = h.keypointService.UpdateKeypoint(keypointToUpdate)
	if err != nil {
		http.Error(w, "Failed to update keypoint", http.StatusInternalServerError)
		return
	}

	_ = h.tourService.RecalculateTourLength(r.Context(), keypointToUpdate.TourID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keypointToUpdate)
}

func (h *KeypointHandler) DeleteKeypoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keypointIDStr := vars["keypointId"]

	keypointID, err := strconv.Atoi(keypointIDStr)
	if err != nil {
		http.Error(w, "Invalid keypoint ID", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingKeypoint, err := h.keypointService.GetKeypointByID(keypointID)
	if err != nil {
		http.Error(w, "Keypoint not found", http.StatusNotFound)
		return
	}

	tour, err := h.tourService.GetTourByID(existingKeypoint.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != userID {
		http.Error(w, "Only tour author can delete keypoints", http.StatusForbidden)
		return
	}

	err = h.keypointService.DeleteKeypoint(keypointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete keypoint", http.StatusInternalServerError)
		}
		return
	}

	_ = h.tourService.RecalculateTourLength(r.Context(), existingKeypoint.TourID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Keypoint deleted successfully",
	})
}
