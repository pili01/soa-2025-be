package models

type RegisterUserRequest struct {
    Username  string `json:"username"`
    Password  string `json:"password"`
    Email     string `json:"email"`
    Role      string `json:"role"`
    Name      string `json:"name"`
    Surname   string `json:"surname"`
    Biography string `json:"biography"`
    Moto      string `json:"moto"`
    PhotoURL  string `json:"photo_url"`
}