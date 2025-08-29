package handlers

import (
	"encoding/json"
	"net/http"
	"stakeholders-service/internal/models"
	repository "stakeholders-service/internal/repositories"
	"stakeholders-service/internal/utils"
	"database/sql"
	"strings"
)

type PositionHandler struct {
	positionRepo *repository.PositionRepository
	userRepo *repository.UserRepository
}

func NewPositionHandler(positionRepo *repository.PositionRepository, userRepo *repository.UserRepository) *PositionHandler {
	return &PositionHandler{positionRepo: positionRepo, userRepo: userRepo}
}

func (h *PositionHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req models.Position

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Authorization token is missing", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserByUsername(claims.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.UserId = user.ID

	createdPosition, err := h.positionRepo.CreatePosition(&req)
	if err != nil {
		http.Error(w, "Failed to create/update position: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
    "message": "Position saved successfully!",
    "position": createdPosition,
	})
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
