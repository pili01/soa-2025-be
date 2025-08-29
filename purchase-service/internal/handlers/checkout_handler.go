package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"purchase-service/internal/models"
	"purchase-service/internal/services"

	"github.com/gorilla/mux"
)

type CheckoutHandler struct {
	checkoutService *services.CheckoutService
	authService     *services.AuthService
}

func NewCheckoutHandler(checkoutService *services.CheckoutService, authService *services.AuthService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
		authService:     authService,
	}
}

func (h *CheckoutHandler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.checkoutService.ProcessCheckout(touristID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CheckoutHandler) GetPurchaseHistory(w http.ResponseWriter, r *http.Request) {
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response, err := h.checkoutService.GetPurchaseHistory(touristID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CheckoutHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tourIDStr := r.URL.Query().Get("tour_id")
	token := r.URL.Query().Get("token")

	if tourIDStr == "" || token == "" {
		http.Error(w, "Missing tour_id or token parameter", http.StatusBadRequest)
		return
	}

	tourID := 0
	if _, err := fmt.Sscanf(tourIDStr, "%d", &tourID); err != nil {
		http.Error(w, "Invalid tour_id parameter", http.StatusBadRequest)
		return
	}

	isValid, err := h.checkoutService.ValidateToken(token, tourID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":      isValid,
		"tour_id":    tourID,
		"tourist_id": touristID,
	})
}

func (h *CheckoutHandler) CheckIsPurchased(w http.ResponseWriter, r *http.Request) {
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tourIDStr := vars["tourId"]

	if tourIDStr == "" {
		http.Error(w, "Missing tourId parameter", http.StatusBadRequest)
		return
	}

	tourID := 0
	if _, err := fmt.Sscanf(tourIDStr, "%d", &tourID); err != nil {
		http.Error(w, "Invalid tourId parameter", http.StatusBadRequest)
		return
	}

	isPurchased, err := h.checkoutService.CheckIsPurchased(touristID, tourID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"purchased":  isPurchased,
		"tour_id":    tourID,
		"tourist_id": touristID,
	})
}
