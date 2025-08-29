package handlers

import (
	"encoding/json"
	"net/http"
	"stakeholders-service/internal/models"
	repository "stakeholders-service/internal/repositories"
	"stakeholders-service/internal/utils"
	"strings"
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
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
		return
	}
	tokenString := tokenParts[1]

	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	userId := claims.ID

	if claims.Role != "Tourist" {
		http.Error(w, "Forbidden", http.StatusForbidden)
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
