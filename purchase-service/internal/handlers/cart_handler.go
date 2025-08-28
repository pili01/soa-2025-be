package handlers

import (
	"encoding/json"
	"net/http"

	"purchase-service/internal/models"
	"purchase-service/internal/services"
)

type CartHandler struct {
	cartService *services.CartService
	authService *services.AuthService
}

func NewCartHandler(cartService *services.CartService, authService *services.AuthService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		authService: authService,
	}
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request models.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.cartService.AddToCart(touristID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item added to cart successfully"})
}

func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	
	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request models.RemoveFromCartRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.cartService.RemoveFromCart(touristID, request.ItemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item removed from cart successfully"})
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {

	touristID, err := h.authService.ValidateAndGetUserID(r, "Tourist")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	cart, err := h.cartService.GetCart(touristID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}
