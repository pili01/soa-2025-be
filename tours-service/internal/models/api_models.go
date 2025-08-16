package models

type ValidationResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsValid  bool   `json:"isValid"`
}

type TourWithKeypointsRequest struct {
	Tour      Tour       `json:"tour"`
	Keypoints []*Keypoint `json:"keypoints"`
}

type CreateKeypointRequest struct {
	TourID      int     `json:"tourId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"imageUrl"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Ordinal     int     `json:"ordinal"`
}