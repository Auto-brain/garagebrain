package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	Region       string    `json:"region"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Currency string `json:"currency"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Currency string `json:"currency"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
