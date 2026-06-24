package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

func CreateCar(ctx context.Context, userID uuid.UUID, req model.CreateCarRequest) (*model.Car, error) {
	var c model.Car
	err := Pool.QueryRow(ctx,
		`INSERT INTO cars (user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING id, user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url, is_active, created_at`,
		userID, req.Brand, req.Model, req.Year, req.VIN, req.Color, req.Engine, req.Drive, req.Mileage,
		req.BoughtDate, req.BoughtPrice, req.PhotoURL,
	).Scan(&c.ID, &c.UserID, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.Color, &c.Engine, &c.Drive,
		&c.Mileage, &c.BoughtDate, &c.BoughtPrice, &c.PhotoURL, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetCarsByUser(ctx context.Context, userID uuid.UUID) ([]model.Car, error) {
	rows, err := Pool.Query(ctx,
		"SELECT id, user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url, is_active, created_at FROM cars WHERE user_id = $1 AND is_active = true ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []model.Car
	for rows.Next() {
		var c model.Car
		err := rows.Scan(&c.ID, &c.UserID, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.Color, &c.Engine, &c.Drive,
			&c.Mileage, &c.BoughtDate, &c.BoughtPrice, &c.PhotoURL, &c.IsActive, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		cars = append(cars, c)
	}
	return cars, nil
}

func GetCarByID(ctx context.Context, carID uuid.UUID) (*model.Car, error) {
	var c model.Car
	err := Pool.QueryRow(ctx,
		"SELECT id, user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url, is_active, created_at FROM cars WHERE id = $1",
		carID,
	).Scan(&c.ID, &c.UserID, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.Color, &c.Engine, &c.Drive,
		&c.Mileage, &c.BoughtDate, &c.BoughtPrice, &c.PhotoURL, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCar обновляет основные данные авто (марка/модель/год/пробег/VIN/…).
func UpdateCar(ctx context.Context, carID uuid.UUID, req model.UpdateCarRequest) (*model.Car, error) {
	var c model.Car
	err := Pool.QueryRow(ctx,
		`UPDATE cars
		 SET brand = $2, model = $3, year = $4, mileage = $5, vin = $6, color = $7, engine = $8, drive = $9
		 WHERE id = $1
		 RETURNING id, user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url, is_active, created_at`,
		carID, req.Brand, req.Model, req.Year, req.Mileage, req.VIN, req.Color, req.Engine, req.Drive,
	).Scan(&c.ID, &c.UserID, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.Color, &c.Engine, &c.Drive,
		&c.Mileage, &c.BoughtDate, &c.BoughtPrice, &c.PhotoURL, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func UpdateCarMileage(ctx context.Context, carID uuid.UUID, mileage int) error {
	_, err := Pool.Exec(ctx, "UPDATE cars SET mileage = $1 WHERE id = $2", mileage, carID)
	return err
}

func GetAllActiveCars(ctx context.Context) ([]model.Car, error) {
	rows, err := Pool.Query(ctx,
		"SELECT id, user_id, brand, model, year, vin, color, engine, drive, mileage, bought_date, bought_price, photo_url, is_active, created_at FROM cars WHERE is_active = true",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []model.Car
	for rows.Next() {
		var c model.Car
		err := rows.Scan(&c.ID, &c.UserID, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.Color, &c.Engine, &c.Drive,
			&c.Mileage, &c.BoughtDate, &c.BoughtPrice, &c.PhotoURL, &c.IsActive, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		cars = append(cars, c)
	}
	return cars, nil
}
