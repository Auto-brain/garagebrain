package model

import (
	"time"

	"github.com/google/uuid"
)

type Car struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Brand       string     `json:"brand"`
	Model       string     `json:"model"`
	Year        *int       `json:"year"`
	VIN         *string    `json:"vin"`
	Color       *string    `json:"color"`
	Engine      *string    `json:"engine"`
	Drive       *string    `json:"drive"`
	Mileage     int        `json:"mileage"`
	BoughtDate  *time.Time `json:"bought_date"`
	BoughtPrice *int       `json:"bought_price"`
	PhotoURL    *string    `json:"photo_url"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateCarRequest struct {
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Year        *int    `json:"year"`
	VIN         *string `json:"vin"`
	Color       *string `json:"color"`
	Engine      *string `json:"engine"`
	Drive       *string `json:"drive"`
	Mileage     int     `json:"mileage"`
	BoughtDate  *string `json:"bought_date"`
	BoughtPrice *int    `json:"bought_price"`
	PhotoURL    *string `json:"photo_url"`
}

type UpdateMileageRequest struct {
	Mileage int `json:"mileage"`
}
