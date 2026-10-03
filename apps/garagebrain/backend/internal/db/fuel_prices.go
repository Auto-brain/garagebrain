package db

import (
	"context"
	"math"

	"github.com/auto-brain/garagebrain/internal/model"
)

// RecordFuelPrice saves a new price point only if it differs >2% from the latest known price.
func RecordFuelPrice(ctx context.Context, region, currency string, pricePerLiter float64) error {
	var latest float64
	err := Pool.QueryRow(ctx,
		`SELECT price_per_liter FROM fuel_prices WHERE region = $1 ORDER BY recorded_at DESC LIMIT 1`,
		region,
	).Scan(&latest)
	if err == nil {
		// Skip if change is under 2%
		if math.Abs(pricePerLiter-latest)/latest < 0.02 {
			return nil
		}
	}
	_, err = Pool.Exec(ctx,
		`INSERT INTO fuel_prices (region, currency, price_per_liter) VALUES ($1, $2, $3)`,
		region, currency, pricePerLiter,
	)
	return err
}

func GetFuelPricesResponse(ctx context.Context, region string) (*model.FuelPricesResponse, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, region, currency, price_per_liter, recorded_at
		 FROM fuel_prices WHERE region = $1
		 ORDER BY recorded_at DESC LIMIT 30`,
		region,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.FuelPrice
	for rows.Next() {
		var p model.FuelPrice
		if err := rows.Scan(&p.ID, &p.Region, &p.Currency, &p.PricePerLiter, &p.RecordedAt); err != nil {
			return nil, err
		}
		history = append(history, p)
	}

	resp := &model.FuelPricesResponse{Region: region, History: history}
	if len(history) > 0 {
		resp.Latest = &history[0]
	}
	return resp, nil
}
