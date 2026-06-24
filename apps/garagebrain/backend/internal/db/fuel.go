package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

func CreateFuelRecord(ctx context.Context, req model.CreateFuelRequest) (*model.FuelRecord, error) {
	fullTank := true
	if req.FullTank != nil {
		fullTank = *req.FullTank
	}

	var f model.FuelRecord
	err := Pool.QueryRow(ctx,
		`INSERT INTO fuel_records (car_id, date, mileage, liters, cost, station, full_tank)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 RETURNING id, car_id, date, mileage, liters, cost, station, full_tank`,
		req.CarID, req.Date, req.Mileage, req.Liters, req.Cost, req.Station, fullTank,
	).Scan(&f.ID, &f.CarID, &f.Date, &f.Mileage, &f.Liters, &f.Cost, &f.Station, &f.FullTank)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetFuelRecordsByCar возвращает заправки авто, упорядоченные по пробегу
// (для корректного расчёта расхода между полными баками).
func GetFuelRecordsByCar(ctx context.Context, carID uuid.UUID) ([]model.FuelRecord, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, car_id, date, mileage, liters, cost, station, full_tank
		 FROM fuel_records WHERE car_id = $1 ORDER BY mileage ASC`,
		carID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []model.FuelRecord{}
	for rows.Next() {
		var f model.FuelRecord
		if err := rows.Scan(&f.ID, &f.CarID, &f.Date, &f.Mileage, &f.Liters, &f.Cost,
			&f.Station, &f.FullTank); err != nil {
			return nil, err
		}
		records = append(records, f)
	}
	return records, nil
}

// ComputeFuelStats считает средний расход по методу полного бака:
// между двумя последовательными заправками «до полного» залитый на второй
// объём топлива покрывает пройденное расстояние.
func ComputeFuelStats(records []model.FuelRecord) model.FuelStats {
	stats := model.FuelStats{FillCount: len(records)}

	var (
		consumedLiters  float64
		consumedDistGap int
		prevFullMileage = -1
	)

	for _, f := range records {
		if f.Liters != nil {
			stats.TotalLiters += *f.Liters
		}
		if f.Cost != nil {
			stats.TotalCost += *f.Cost
		}

		if !f.FullTank || f.Liters == nil {
			continue
		}
		if prevFullMileage >= 0 && f.Mileage > prevFullMileage {
			consumedDistGap += f.Mileage - prevFullMileage
			consumedLiters += *f.Liters
		}
		prevFullMileage = f.Mileage
	}

	stats.TotalDistance = consumedDistGap
	if consumedDistGap > 0 {
		stats.AvgConsumption = consumedLiters / float64(consumedDistGap) * 100
	}
	return stats
}
