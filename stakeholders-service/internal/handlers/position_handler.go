package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"stakeholders-service/internal/models"
	repository "stakeholders-service/internal/repositories"
)

type PositionHandler struct {
	positionRepo *repository.PositionRepository
}

func NewPositionHandler(positionRepo *repository.PositionRepository) *PositionHandler {
	return &PositionHandler{positionRepo: positionRepo}
}

func (h *PositionHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req models.Position
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.positionRepo.CreatePosition(&req)
	if err != nil {
		http.Error(w, "Failed to create/update position: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Position saved successfully!"})
}

func (h *PositionHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["userId"]
	if userIdStr == "" {
		http.Error(w, "Missing userId in path", http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid userId format", http.StatusBadRequest)
		return
	}

	position, err := h.positionRepo.GetPositionByUserID(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}
