package models

import "time"

type TourCapacity struct {
	TourID         int       `bson:"tourId" json:"tourId"`
	Capacity       int       `bson:"capacity" json:"capacity"`
	AvailableSeats int       `bson:"availableSeats" json:"availableSeats"`
	UpdatedAt      time.Time `bson:"updatedAt" json:"updatedAt"`
}