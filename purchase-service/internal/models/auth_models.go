package models

type ValidationResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsValid  bool   `json:"isValid"`
}
