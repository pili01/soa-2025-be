package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"os"
	"tours-service/internal/models"
	"tours-service/internal/repositories"
)

type TourHandler struct {
	tourRepo *repositories.TourRepository
}

func NewTourHandler(repo *repositories.TourRepository) *TourHandler {
	return &TourHandler{tourRepo: repo}
}

type ValidationResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsValid  bool   `json:"isValid"`
}

func (h *TourHandler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour models.Tour
	if err := json.NewDecoder(r.Body).Decode(&tour); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Guide"
	req, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		http.Error(w, errorBody.String(), resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}
	
	tour.AuthorID = validationResp.UserID
	tour.Status = models.StatusDraft
	tour.Price = 0.0

	err = h.tourRepo.CreateTour(&tour)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create tour in database", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tour)
}

func (h *TourHandler) GetToursByAuthor(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Guide"
	req, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		http.Error(w, errorBody.String(), resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}

	tours, err := h.tourRepo.GetToursByAuthorID(validationResp.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve tours from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tours)
}