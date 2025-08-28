package models

import (
	"time"
)

type OrderItem struct {
	ID       int     `json:"id" db:"id"`
	CartID   int     `json:"cart_id" db:"cart_id"`
	TourID   int     `json:"tour_id" db:"tour_id"`
	TourName string  `json:"tour_name" db:"tour_name"`
	Price    float64 `json:"price" db:"price"`
	Quantity int     `json:"quantity" db:"quantity"`
	AddedAt  time.Time `json:"added_at" db:"added_at"`
}

type AddToCartRequest struct {
	TourID    int     `json:"tour_id" validate:"required"`
	TourName  string  `json:"tour_name" validate:"required"`
	Price     float64 `json:"price" validate:"required,min=0"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
}

type RemoveFromCartRequest struct {
	ItemID int `json:"item_id" validate:"required"`
}
