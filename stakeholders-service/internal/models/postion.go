package models

type Position struct {
	ID        int     `json:"id"`
	UserId    int     `json:"user_id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
