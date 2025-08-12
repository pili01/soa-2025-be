package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Password should not be exposed in JSON
	Email     string    `json:"email"`
	Role      string    `json:"role"` 
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Biography string    `json:"biography"`
	Moto      string    `json:"moto"`
	PhotoURL  string    `json:"photo_url"`
	IsBlocked bool      `json:"is_blocked"`
}