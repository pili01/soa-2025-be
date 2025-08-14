package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"tours-service/internal/models"
	"tours-service/internal/repositories"

	"github.com/gorilla/mux"
)

type TourReviewHandler struct {
	reviewRepo *repositories.TourReviewRepository
	tourRepo   *repositories.TourRepository
}

func NewTourReviewHandler(reviewRepo *repositories.TourReviewRepository, tourRepo *repositories.TourRepository) *TourReviewHandler {
	return &TourReviewHandler{
		reviewRepo: reviewRepo,
		tourRepo:   tourRepo,
	}
}

func (h *TourReviewHandler) CreateTourReview(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	tourID, err := strconv.Atoi(r.FormValue("tourId"))
	if err != nil {
		http.Error(w, "Invalid or missing tourId", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil {
		http.Error(w, "Invalid or missing rating", http.StatusBadRequest)
		return
	}
	if rating < 1 || rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}
	
	visitDateStr := r.FormValue("visitDate")
	if visitDateStr == "" {
		http.Error(w, "visitDate is required", http.StatusBadRequest)
		return
	}
	visitDate, err := time.Parse(time.RFC3339, visitDateStr)
	if err != nil {
		http.Error(w, "Invalid visitDate format. Use RFC3339 (e.g., YYYY-MM-DDTHH:MM:SSZ)", http.StatusBadRequest)
		return
	}
	
	comment := r.FormValue("comment")
	if comment == "" {
		http.Error(w, "Comment is required", http.StatusBadRequest)
		return
	}

	_, err = h.tourRepo.GetTourByID(tourID)
	if err != nil {
		http.Error(w, "Tour not found", http.StatusNotFound)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	validationURL := os.Getenv("STAKEHOLDERS_SERVICE_URL") + "/api/validateRole?role=Tourist"
	reqAuth, err := http.NewRequest("POST", validationURL, nil)
	if err != nil {
		http.Error(w, "Failed to create validation request", http.StatusInternalServerError)
		return
	}
	reqAuth.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(reqAuth)
	if err != nil {
		http.Error(w, "Failed to contact authentication service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Only users with role 'Tourist' can leave a review", resp.StatusCode)
		return
	}

	var validationResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		http.Error(w, "Failed to decode validation response", http.StatusInternalServerError)
		return
	}
	touristID := validationResp.UserID

	// TODO: Da li je turista bio na turi

	var imageURLs []string
	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening uploaded file", http.StatusInternalServerError)
			return
		}

		photoURL, uploadErr := h.uploadReviewImageToService(file, fileHeader.Filename, touristID, tourID)
		if uploadErr != nil {
			fmt.Printf("Failed to upload image %s: %v\n", fileHeader.Filename, uploadErr)
			http.Error(w, "Failed to upload image: "+uploadErr.Error(), http.StatusInternalServerError)
			file.Close()
			return
		}
		file.Close()
		imageURLs = append(imageURLs, photoURL)
	}

	review := &models.TourReview{
		TourID:    tourID,
		TouristID: touristID,
		Rating:    rating,
		Comment:   comment,
		VisitDate: visitDate,
		ImageURLs: imageURLs,
	}

	err = h.reviewRepo.CreateTourReview(review)
	if err != nil {
		http.Error(w, "Failed to create tour review in database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}


func (h *TourReviewHandler) uploadReviewImageToService(file io.Reader, filename string, userId int, tourId int) (string, error) {
    var body bytes.Buffer
    writer := multipart.NewWriter(&body)

    part, err := writer.CreateFormFile("image", filepath.Base(filename))
    if err != nil {
        return "", fmt.Errorf("failed to create form file: %w", err)
    }
    
    if _, err = io.Copy(part, file); err != nil {
        return "", fmt.Errorf("failed to copy file: %w", err)
    }

    writer.WriteField("userId", strconv.Itoa(userId))
    writer.WriteField("tourId", strconv.Itoa(tourId))
    writer.Close()

    req, err := http.NewRequest(
        "POST",
        os.Getenv("IMAGE_SERVICE_URL")+"/api/saveReviewPhoto",
        &body,
    )
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request to image service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        respBody, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("image service responded with error: %s", string(respBody))
    }

    var result struct {
        PhotoURL string `json:"photoURL"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response from image service: %w", err)
    }

    return result.PhotoURL, nil
}


func (h *TourReviewHandler) GetReviewsByTourID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil {
		http.Error(w, "Invalid tour ID", http.StatusBadRequest)
		return
	}

	reviews, err := h.reviewRepo.GetReviewsByTourID(tourID)
	if err != nil {
		http.Error(w, "Failed to retrieve reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reviews)
}
