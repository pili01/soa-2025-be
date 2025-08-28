package models

import (
	"time"
)

type TourPurchaseToken struct {
	ID          int       `json:"id" db:"id"`
	TouristID   int       `json:"tourist_id" db:"tourist_id"`
	TourID      int       `json:"tour_id" db:"tour_id"`
	Token       string    `json:"token" db:"token"`
	PurchasedAt time.Time `json:"purchased_at" db:"purchased_at"`
}

type CheckoutRequest struct {
	CartID int `json:"cart_id" validate:"required"`
}

type CheckoutResponse struct {
	Success bool                    `json:"success"`
	Tokens  []TourPurchaseToken    `json:"tokens"`
	Message string                  `json:"message"`
}

type PurchaseHistoryResponse struct {
	TouristID int                    `json:"tourist_id"`
	Purchases []TourPurchaseToken    `json:"purchases"`
}
