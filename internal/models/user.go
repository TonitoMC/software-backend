package models

// User model
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}
