package models

import "time"

type TourReview struct {
	ID          int       `bson:"_id,omitempty" json:"id"`
	TourID      int       `bson:"tourId" json:"tourId"`
	TouristID   int       `bson:"touristId" json:"touristId"`
	Rating      int       `bson:"rating" json:"rating"`
	Comment     string    `bson:"comment" json:"comment"`
	VisitDate   time.Time `bson:"visitDate" json:"visitDate"`
	CommentDate time.Time `bson:"commentDate" json:"commentDate"`
	ImageURLs   []string  `bson:"imageUrls" json:"imageUrls"`
}

type CreateTourReviewRequest struct {
	TourID    int       `json:"tourId" validate:"required"`
	Rating    int       `json:"rating" validate:"required,min=1,max=5"`
	Comment   string    `json:"comment" validate:"required"`
	VisitDate time.Time `json:"visitDate" validate:"required"`
	ImageURLs []string  `json:"imageUrls"`
}