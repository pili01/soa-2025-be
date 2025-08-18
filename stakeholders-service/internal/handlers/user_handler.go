package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"stakeholders-service/internal/models"
	repository "stakeholders-service/internal/repositories"
	"stakeholders-service/internal/utils"
	proto "stakeholders-service/proto"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
)

type UserHandler struct {
	userRepo *repository.UserRepository
	proto.UnimplementedStakeholdersServiceServer
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	var req models.RegisterUserRequest
	req.Username = r.FormValue("username")
	req.Password = r.FormValue("password")
	req.Email = r.FormValue("email")
	req.Role = r.FormValue("role")
	req.Name = r.FormValue("name")
	req.Surname = r.FormValue("surname")
	req.Biography = r.FormValue("biography")
	req.Moto = r.FormValue("moto")

	if req.Role != "Guide" && req.Role != "Tourist" {
		http.Error(w, "Invalid role. Role must be 'Guide' or 'Tourist'.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		photoURL, uploadErr := h.uploadToImageService(file, handler.Filename)
		if uploadErr != nil {
			http.Error(w, "Failed to upload image", http.StatusInternalServerError)
			return
		}
		req.PhotoURL = photoURL
	} else if err != http.ErrMissingFile {
		http.Error(w, "Error reading file", http.StatusBadRequest)
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

func (h *UserHandler) GetUserFromToken(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":   claims.ID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

func (h *UserHandler) GetMyInfo(ctx context.Context, req *proto.GetMyInfoRequest) (*proto.GetMyInfoResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		fmt.Println("failed to retrieve metadata from context")
		return nil, fmt.Errorf("failed to retrieve metadata")
	}

	if authHeaders, exists := md["authorization"]; exists && len(authHeaders) > 0 {
		tokenParts := strings.Split(authHeaders[0], " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return nil, fmt.Errorf("authorization header format must be Bearer {token}")
		}
		tokenString := tokenParts[1]

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return nil, fmt.Errorf("invalid or expired token")
		}

		// Retrieve user information from the database
		user, err := h.userRepo.GetUserByID(claims.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data: %w", err)
		}

		return &proto.GetMyInfoResponse{
			Id:       (int64)(user.ID),
			Username: user.Username,
			Role:     user.Role,
		}, nil
	} else {
		return nil, fmt.Errorf("authorization header is missing")
	}
}

func (h *UserHandler) uploadToImageService(file io.Reader, filename string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("image", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest(
		"POST",
		os.Getenv("IMAGE_SERVICE_URL")+"/api/save-image",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("image service error: %s", string(respBody))
	}

	var result struct {
		PhotoURL string `json:"photoURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.PhotoURL, nil
}
