package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type PurchaseService struct {
}

func NewPurchaseService() *PurchaseService {
	return &PurchaseService{}
}

func (s *PurchaseService) IsTourPurchasedByMe(r *http.Request, tourId int) (bool, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, errors.New("authorization header is required")
	}

	validationURL := os.Getenv("PURCHASE_SERVICE_URL") + "/check-is-purchased/" + strconv.Itoa(tourId)
	fmt.Println("Validation URL:", validationURL) // Debugging line
	req, err := http.NewRequest("GET", validationURL, nil)
	if err != nil {
		return false, errors.New("failed to create check purchase request")
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.New("failed to contact purchase service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return false, fmt.Errorf("bad request: %s", errorBody.String())
	}

	var respData struct {
		Purchased bool `json:"purchased"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return false, errors.New("failed to decode validation response")
	}

	return respData.Purchased, nil
}
