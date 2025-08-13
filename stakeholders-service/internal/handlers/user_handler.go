package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"stakeholders-service/internal/models"
	"stakeholders-service/internal/repository"
	"stakeholders-service/internal/utils"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newUser := models.User{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Role:      req.Role,
		Name:      req.Name,
		Surname:   req.Surname,
		Biography: req.Biography,
		Moto:      req.Moto,
		PhotoURL:  req.PhotoURL,
		IsBlocked: false,
	}

	if newUser.Role != "Guide" && newUser.Role != "Tourist" {
		http.Error(w, "Invalid role. Role must be 'Guide' or 'Tourist'.", http.StatusBadRequest)
		return
	}

	err = h.userRepo.CreateUser(&newUser)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully!"})
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials models.LoginCredentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByUsername(credentials.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if user.IsBlocked {
		http.Error(w, "Account is blocked. Please contact administrator.", http.StatusForbidden)
		return
	}

	tokenString, err := utils.GenerateToken(user.Username, user.Role)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
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

	var requestData struct {
		Name      string `json:"name"`
		Surname   string `json:"surname"`
		Biography string `json:"biography"`
		Moto      string `json:"moto"`
		PhotoURL  string `json:"photo_url"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.Name = requestData.Name
	user.Surname = requestData.Surname
	user.Biography = requestData.Biography
	user.Moto = requestData.Moto
	user.PhotoURL = requestData.PhotoURL

	err = h.userRepo.UpdateProfile(user)
	if err != nil {
		http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully!"})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {

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

	if claims.Role != "Admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
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

	if claims.Role != "Admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	var req struct {
		UserID    int  `json:"user_id"`
		IsBlocked bool `json:"is_blocked"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByID(req.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.Role == "Admin" {
		http.Error(w, "Cannot block admin users", http.StatusForbidden)
		return
	}

	err = h.userRepo.UpdateUserBlockStatus(req.UserID, req.IsBlocked)
	if err != nil {
		http.Error(w, "Failed to update user block status", http.StatusInternalServerError)
		return
	}

	action := "blocked"
	if !req.IsBlocked {
		action = "unblocked"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User " + action + " successfully",
		"user_id": string(rune(req.UserID)),
	})
}
