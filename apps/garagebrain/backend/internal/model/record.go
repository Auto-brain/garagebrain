package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ServiceRecord struct {
	ID         uuid.UUID       `json:"id"`
	CarID      uuid.UUID       `json:"car_id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Description string         `json:"description,omitempty"`
	Date       time.Time       `json:"date"`
	Mileage    *int            `json:"mileage"`
	Cost       *int            `json:"cost"`
	Parts      json.RawMessage `json:"parts"`
	Workshop   *string         `json:"workshop"`
	Photos     []string        `json:"photos,omitempty"`
	RawInput   *string         `json:"raw_input,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type CreateRecordRequest struct {
	CarID       uuid.UUID `json:"car_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Date        string    `json:"date"`
	Mileage     *int      `json:"mileage"`
	Cost        *int      `json:"cost"`
	Parts       json.RawMessage `json:"parts"`
	Workshop    *string   `json:"workshop"`
}

type FuelRecord struct {
	ID       uuid.UUID `json:"id"`
	CarID    uuid.UUID `json:"car_id"`
	Date     time.Time `json:"date"`
	Mileage  int       `json:"mileage"`
	Liters   *float64  `json:"liters"`
	Cost     *int      `json:"cost"`
	Station  *string   `json:"station"`
	FullTank bool      `json:"full_tank"`
}
