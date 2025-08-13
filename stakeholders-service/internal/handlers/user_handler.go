package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt" 
	"net/http"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"stakeholders-service/internal/models"
	"stakeholders-service/internal/repositories"
	"stakeholders-service/internal/utils"
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
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
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
		if err.Error() == "user not found" {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if user.IsBlocked {
		http.Error(w, "Account is blocked. Please contact administrator.", http.StatusForbidden)
		return
	}

	tokenString, err := utils.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
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
		if err == sql.ErrNoRows || err.Error() == "user not found" {
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User " + action + " successfully",
		"user_id": req.UserID,
	})
}

func (h *UserHandler) ValidateRole(w http.ResponseWriter, r *http.Request) {
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

	requiredRoles := r.URL.Query()["role"]
	if len(requiredRoles) > 0 {
		isAuthorized := false
		for _, requiredRole := range requiredRoles {
			if claims.Role == requiredRole {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			errMsg := fmt.Sprintf("Insufficient permissions: User role '%s' does not match required roles: %v", claims.Role, requiredRoles)
			http.Error(w, errMsg, http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":   claims.ID,
		"username": claims.Username,
		"role":     claims.Role,
		"isValid":  true,
	})
}