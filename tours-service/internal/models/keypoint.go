package models

type Keypoint struct {
	ID          int       `bson:"_id,omitempty" json:"id"`
	TourID      int       `bson:"tourId" json:"tourId"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description" json:"description"`
	ImageURL    string    `bson:"imageUrl" json:"imageUrl"`
	Latitude    float64   `bson:"latitude" json:"latitude"`
	Longitude   float64   `bson:"longitude" json:"longitude"`
	Ordinal     int       `bson:"ordinal" json:"ordinal"`
}



