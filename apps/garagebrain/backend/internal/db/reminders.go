package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

func CreateReminder(ctx context.Context, req model.CreateReminderRequest) (*model.Reminder, error) {
	var r model.Reminder
	err := Pool.QueryRow(ctx,
		`INSERT INTO reminders (car_id, title, type, trigger_mileage, trigger_date, interval_km, interval_days)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 RETURNING id, car_id, title, type, trigger_mileage, trigger_date, interval_km, interval_days, is_active, last_triggered_at, created_at`,
		req.CarID, req.Title, req.Type, req.TriggerMileage, req.TriggerDate, req.IntervalKm, req.IntervalDays,
	).Scan(&r.ID, &r.CarID, &r.Title, &r.Type, &r.TriggerMileage, &r.TriggerDate, &r.IntervalKm, &r.IntervalDays,
		&r.IsActive, &r.LastTriggeredAt, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func GetRemindersByCar(ctx context.Context, carID uuid.UUID) ([]model.Reminder, error) {
	rows, err := Pool.Query(ctx,
		"SELECT id, car_id, title, type, trigger_mileage, trigger_date, interval_km, interval_days, is_active, last_triggered_at, created_at FROM reminders WHERE car_id = $1 AND is_active = true ORDER BY trigger_date ASC NULLS LAST",
		carID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.Reminder
	for rows.Next() {
		var r model.Reminder
		err := rows.Scan(&r.ID, &r.CarID, &r.Title, &r.Type, &r.TriggerMileage, &r.TriggerDate, &r.IntervalKm,
			&r.IntervalDays, &r.IsActive, &r.LastTriggeredAt, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

func GetDueDateReminders(ctx context.Context, now interface{}) ([]model.Reminder, error) {
	rows, err := Pool.Query(ctx,
		`SELECT r.id, r.car_id, r.title, r.type, r.trigger_mileage, r.trigger_date, r.interval_km, r.interval_days, r.is_active, r.last_triggered_at, r.created_at
		 FROM reminders r
		 WHERE r.is_active = true AND r.type = 'date' AND r.trigger_date <= CURRENT_DATE`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.Reminder
	for rows.Next() {
		var r model.Reminder
		err := rows.Scan(&r.ID, &r.CarID, &r.Title, &r.Type, &r.TriggerMileage, &r.TriggerDate, &r.IntervalKm,
			&r.IntervalDays, &r.IsActive, &r.LastTriggeredAt, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

func GetDueMileageReminders(ctx context.Context, carID uuid.UUID, currentMileage int) ([]model.Reminder, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, car_id, title, type, trigger_mileage, trigger_date, interval_km, interval_days, is_active, last_triggered_at, created_at
		 FROM reminders
		 WHERE car_id = $1 AND is_active = true AND type = 'mileage' AND trigger_mileage <= $2`,
		carID, currentMileage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.Reminder
	for rows.Next() {
		var r model.Reminder
		err := rows.Scan(&r.ID, &r.CarID, &r.Title, &r.Type, &r.TriggerMileage, &r.TriggerDate, &r.IntervalKm,
			&r.IntervalDays, &r.IsActive, &r.LastTriggeredAt, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

func MarkReminderTriggered(ctx context.Context, reminderID uuid.UUID) error {
	_, err := Pool.Exec(ctx, "UPDATE reminders SET last_triggered_at = now() WHERE id = $1", reminderID)
	return err
}

func GetUserByCarID(ctx context.Context, carID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := Pool.QueryRow(ctx, "SELECT user_id FROM cars WHERE id = $1", carID).Scan(&userID)
	return userID, err
}
