package model

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID              uuid.UUID  `json:"id"`
	CarID           uuid.UUID  `json:"car_id"`
	Title           string     `json:"title"`
	Type            string     `json:"type"`
	TriggerMileage  *int       `json:"trigger_mileage"`
	TriggerDate     *time.Time `json:"trigger_date"`
	IntervalKm      *int       `json:"interval_km"`
	IntervalDays    *int       `json:"interval_days"`
	IsActive        bool       `json:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateReminderRequest struct {
	CarID          uuid.UUID `json:"car_id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	TriggerMileage *int      `json:"trigger_mileage"`
	TriggerDate    *string   `json:"trigger_date"`
	IntervalKm     *int      `json:"interval_km"`
	IntervalDays   *int      `json:"interval_days"`
}
