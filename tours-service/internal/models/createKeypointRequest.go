package models

type CreateKeypointRequest struct {
	TourID      int     `json:"tourId" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ImageURL    string  `json:"imageUrl"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	Ordinal     int     `json:"ordinal"`
}
