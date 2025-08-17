package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"tours-service/internal/models"
)

type MapService struct {
	Client  *http.Client
	BaseURL string
}

func NewMapService(baseURL string) *MapService {
	return &MapService{
		Client:  &http.Client{Timeout: 10 * time.Second},
		BaseURL: baseURL,
	}
}

func (s *MapService) GetDistanceBetweenTwoKeypoints(ctx context.Context, origin, dest models.Keypoint) (map[string]models.DistanceAndDuration, error) {
	url := fmt.Sprintf("%s/api/getdistances?originLat=%f&originLng=%f&destLat=%f&destLng=%f",
		s.BaseURL, origin.Latitude, origin.Longitude, dest.Latitude, dest.Longitude)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call map-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse map[string]string
		json.NewDecoder(resp.Body).Decode(&errResponse)
		return nil, fmt.Errorf("map-service returned non-OK status %d: %s", resp.StatusCode, errResponse["error"])
	}

	var results map[string]models.DistanceAndDuration
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response from map-service: %w", err)
	}

	return results, nil
}