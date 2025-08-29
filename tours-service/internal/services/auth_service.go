package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"tours-service/internal/models"
)

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) ValidateAndGetUserID(r *http.Request, role string) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header is required")
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=" + role
	req, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		return 0, errors.New("failed to create validation request")
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.New("failed to contact authentication service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return 0, fmt.Errorf("unauthorized: %s", errorBody.String())
	}

	var validationResp models.ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		return 0, errors.New("failed to decode validation response")
	}

	return validationResp.UserID, nil
}

func (s *AuthService) GetMyPosition(r *http.Request, userId int) (float64, float64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, 0, errors.New("authorization header is required")
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/position/" + strconv.Itoa(userId)
	req, err := http.NewRequest("GET", validationURL, nil)
	if err != nil {
		return 0, 0, errors.New("failed to create position request")
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, errors.New("failed to contact authentication service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return 0, 0, fmt.Errorf("unauthorized: %s", errorBody.String())
	}

	var validationResp = struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		return 0, 0, errors.New("failed to decode validation response")
	}

	return validationResp.Longitude, validationResp.Latitude, nil
}
