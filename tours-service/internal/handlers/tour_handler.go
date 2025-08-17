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

type TourHandler struct {
	tourService *services.TourService
	authService *services.AuthService
}

func NewTourHandler(tourService *services.TourService, authService *services.AuthService) *TourHandler {
	return &TourHandler{
		tourService: tourService,
		authService: authService,
	}
}

func (h *TourHandler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var requestBody models.TourWithKeypointsRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tour := &requestBody.Tour
	keypoints := requestBody.Keypoints

	tour.AuthorID = userID
	tour.Status = models.StatusDraft
	tour.Price = 0.0

	err = h.tourService.CreateTour(tour, keypoints)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "at least two keypoints") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create tour: " + err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tour)
}

func (h *TourHandler) GetPublishedToursWithFirstKeypoint(w http.ResponseWriter, r *http.Request) {
	toursWithKeypoints, err := h.tourService.GetPublishedToursWithFirstKeypoint()
	if err != nil {
		http.Error(w, "Failed to retrieve published tours", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toursWithKeypoints)
}

func (h *TourHandler) SetTourPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}
	
	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var priceRequest struct {
		Price float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&priceRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.tourService.SetTourPrice(tourID, priceRequest.Price, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found or you are not the author") {
			http.Error(w, "Tour not found or you are not the author", http.StatusForbidden)
		} else {
			http.Error(w, "Failed to update tour price: " + err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tour price updated successfully",
	})
}

func (h *TourHandler) GetToursByAuthor(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tours, err := h.tourService.GetToursByAuthorID(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve tours from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tours)
}

func (h *TourHandler) GetTourByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	tour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tour not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve tour", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tour)
}

func (h *TourHandler) UpdateTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	var tour models.Tour
	if err := json.NewDecoder(r.Body).Decode(&tour); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingTour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}
	
	if existingTour.AuthorID != userID {
		http.Error(w, "Only tour author can modify a tour", http.StatusForbidden)
		return
	}

	tour.ID = tourID
	tour.AuthorID = existingTour.AuthorID
	tour.Status = existingTour.Status
	tour.Price = existingTour.Price

	err = h.tourService.UpdateTour(&tour)
	if err != nil {
		http.Error(w, "Failed to update tour", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tour)
}

func (h *TourHandler) DeleteTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingTour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}
	
	if existingTour.AuthorID != userID {
		http.Error(w, "Only tour author can delete a tour", http.StatusForbidden)
		return
	}

	err = h.tourService.DeleteTour(tourID)
	if err != nil {
		http.Error(w, "Failed to delete tour", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tour and its keypoints deleted successfully",
	})
}

func (h *TourHandler) PublishTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingTour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tour not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve tour", http.StatusInternalServerError)
		}
		return
	}

	if existingTour.AuthorID != userID {
		http.Error(w, "Only the tour author can publish a tour", http.StatusForbidden)
		return
	}

	err = h.tourService.PublishTour(tourID)
	if err != nil {
		http.Error(w, "Failed to publish tour: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tour published successfully",
	})
}

func (h *TourHandler) ArchiveTour(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	userID, err := h.authService.ValidateAndGetUserID(r, "Guide")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	existingTour, err := h.tourService.GetTourByID(tourID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tour not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve tour", http.StatusInternalServerError)
		}
		return
	}

	if existingTour.AuthorID != userID {
		http.Error(w, "Only the tour author can archive a tour", http.StatusForbidden)
		return
	}

	err = h.tourService.ArchiveTour(tourID)
	if err != nil {
		if strings.Contains(err.Error(), "only be archived if") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to archive tour: " + err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tour archived successfully",
	})
}