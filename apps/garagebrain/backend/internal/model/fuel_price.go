package model

import (
	"time"

	"github.com/google/uuid"
)

type FuelPrice struct {
	ID            uuid.UUID `json:"id"`
	Region        string    `json:"region"`
	Currency      string    `json:"currency"`
	PricePerLiter float64   `json:"price_per_liter"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type FuelPricesResponse struct {
	Region  string      `json:"region"`
	Latest  *FuelPrice  `json:"latest"`
	History []FuelPrice `json:"history"`
}
