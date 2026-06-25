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
	RegNumber   *string    `json:"reg_number"`
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
	RegNumber   *string `json:"reg_number"`
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

// CarMember — участник авто (для списка «Участники авто»).
type CarMember struct {
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateCarRequest struct {
	Brand     string  `json:"brand"`
	Model     string  `json:"model"`
	Year      *int    `json:"year"`
	Mileage   int     `json:"mileage"`
	VIN       *string `json:"vin"`
	RegNumber *string `json:"reg_number"`
	Color     *string `json:"color"`
	Engine    *string `json:"engine"`
	Drive     *string `json:"drive"`
}
