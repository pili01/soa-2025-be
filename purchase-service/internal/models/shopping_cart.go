package models

import (
	"time"
)

type ShoppingCart struct {
	ID         int       `json:"id" db:"id"`
	TouristID  int       `json:"tourist_id" db:"tourist_id"`
	TotalPrice float64   `json:"total_price" db:"total_price"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type ShoppingCartResponse struct {
	ID         int         `json:"id"`
	TouristID  int         `json:"tourist_id"`
	TotalPrice float64     `json:"total_price"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
