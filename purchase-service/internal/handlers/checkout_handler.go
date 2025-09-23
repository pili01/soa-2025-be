package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"purchase-service/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CheckoutHandler struct {
	checkoutService *services.CheckoutService
	authService     *services.AuthService
	orchestrator    *services.TourPurchaseOrchestrator
}

func NewCheckoutHandler(checkoutService *services.CheckoutService, authService *services.AuthService, orch *services.TourPurchaseOrchestrator) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
		authService:     authService,
		orchestrator:    orch,
	}
}

func (h *CheckoutHandler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {

	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	cart, err := h.checkoutService.GetCartForUser(touristID)
	if err != nil {
		log.Printf("Error fetching cart for tourist %d: %v", touristID, err)
		http.Error(w, "Could not retrieve shopping cart", http.StatusInternalServerError)
		return
	}
	if len(cart.Items) == 0 {
		http.Error(w, "Shopping cart is empty", http.StatusBadRequest)
		return
	}

	purchaseID := uuid.New().String()

	err = h.orchestrator.Start(purchaseID, *cart)
	if err != nil {
		log.Printf("CRITICAL: Failed to start saga for purchase %s: %v", purchaseID, err)
		http.Error(w, "Could not initiate checkout process. Please try again later.", http.StatusInternalServerError)
		return
	}

	log.Printf("Saga started successfully for purchase ID: %s", purchaseID)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Checkout process has been initiated.",
		"purchaseId": purchaseID,
	})
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
