package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tours-service/internal/models"
	"tours-service/internal/repositories"
	"github.com/gorilla/mux"
)

type KeypointHandler struct {
	keypointRepo *repositories.KeypointRepository
	tourRepo     *repositories.TourRepository
}

func NewKeypointHandler(keypointRepo *repositories.KeypointRepository, tourRepo *repositories.TourRepository) *KeypointHandler {
	return &KeypointHandler{
		keypointRepo: keypointRepo,
		tourRepo:     tourRepo,
	}
}

func (h *KeypointHandler) CreateKeypoint(w http.ResponseWriter, r *http.Request) {
	var req models.CreateKeypointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Guide"
	reqAuth, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	reqAuth.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(reqAuth)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Unauthorized or insufficient privileges", resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}

	tour, err := h.tourRepo.GetTourByID(req.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != validationResp.UserID {
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

	err = h.keypointRepo.CreateKeypoint(keypoint)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create keypoint", http.StatusInternalServerError)
		}
		return
	}

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

	keypoints, err := h.keypointRepo.GetKeypointsByTourID(tourID)
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

	keypoint, err := h.keypointRepo.GetKeypointByID(keypointID)
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

	var req models.CreateKeypointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Guide"
	reqAuth, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	reqAuth.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(reqAuth)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Unauthorized or insufficient privileges", resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}

	
	existingKeypoint, err := h.keypointRepo.GetKeypointByID(keypointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve keypoint", http.StatusInternalServerError)
		}
		return
	}

	
	tour, err := h.tourRepo.GetTourByID(existingKeypoint.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != validationResp.UserID {
		http.Error(w, "Only tour author can modify keypoints", http.StatusForbidden)
		return
	}


	keypoint := &models.Keypoint{
		ID:          keypointID,
		TourID:      existingKeypoint.TourID, 
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Ordinal:     req.Ordinal,
	}

	err = h.keypointRepo.UpdateKeypoint(keypoint)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update keypoint", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keypoint)
}


func (h *KeypointHandler) DeleteKeypoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keypointIDStr := vars["keypointId"]
	
	keypointID, err := strconv.Atoi(keypointIDStr)
	if err != nil {
		http.Error(w, "Invalid keypoint ID", http.StatusBadRequest)
		return
	}

	
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Guide"
	reqAuth, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	reqAuth.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(reqAuth)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Unauthorized or insufficient privileges", resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}

	
	existingKeypoint, err := h.keypointRepo.GetKeypointByID(keypointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve keypoint", http.StatusInternalServerError)
		}
		return
	}

	
	tour, err := h.tourRepo.GetTourByID(existingKeypoint.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != validationResp.UserID {
		http.Error(w, "Only tour author can delete keypoints", http.StatusForbidden)
		return
	}

	err = h.keypointRepo.DeleteKeypoint(keypointID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Keypoint not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete keypoint", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Keypoint deleted successfully",
	})
}

