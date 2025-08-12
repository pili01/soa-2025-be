package models

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` 
	Email     string    `json:"email"`
	Role      string    `json:"role"` // Admin, Guide, Tourist
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Biography string    `json:"biography"`
	Moto      string    `json:"moto"`
	PhotoURL  string    `json:"photo_url"`
	IsBlocked bool      `json:"is_blocked"`
}