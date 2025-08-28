package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

// UploadKeypointImage handles image upload for a specific keypoint
func (h *KeypointHandler) UploadKeypointImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	keypointID, err := strconv.Atoi(r.FormValue("keypointId"))
	if err != nil {
		http.Error(w, "Invalid or missing keypointId", http.StatusBadRequest)
		return
	}

	// Get keypoint to verify it exists and get tour ID
	keypoint, err := h.keypointService.GetKeypointByID(keypointID)
	if err != nil {
		http.Error(w, "Keypoint not found", http.StatusNotFound)
		return
	}

	// Verify user is the tour author
	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tour, err := h.tourService.GetTourByID(keypoint.TourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	if tour.AuthorID != userID {
		http.Error(w, "Only tour author can upload keypoint images", http.StatusForbidden)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	photoURL, uploadErr := h.uploadKeypointImageToService(file, handler.Filename, keypointID, keypoint.TourID)
	if uploadErr != nil {
		http.Error(w, "Failed to upload image: "+uploadErr.Error(), http.StatusInternalServerError)
		return
	}

	// Update keypoint with new image URL
	keypoint.ImageURL = photoURL
	err = h.keypointService.UpdateKeypoint(keypoint)
	if err != nil {
		http.Error(w, "Failed to update keypoint with image URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Keypoint image uploaded successfully",
		"photoURL": photoURL,
		"keypointId": keypointID,
	})
}

func (h *KeypointHandler) uploadKeypointImageToService(file io.Reader, filename string, keypointId int, tourId int) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("image", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	writer.WriteField("keypointId", strconv.Itoa(keypointId))
	writer.WriteField("tourId", strconv.Itoa(tourId))
	writer.Close()

	req, err := http.NewRequest(
		"POST",
		os.Getenv("IMAGE_SERVICE_URL")+"/api/saveKeypointPhoto",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to image service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("image service responded with error: %s", string(respBody))
	}

	var result struct {
		PhotoURL string `json:"photoURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response from image service: %w", err)
	}

	return result.PhotoURL, nil
}
