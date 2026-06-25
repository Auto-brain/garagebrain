package db

import "context"

// SaveRates сохраняет курсы (база USD) в кэш-таблицу (upsert по валюте).
func SaveRates(ctx context.Context, rates map[string]float64) error {
	for quote, rate := range rates {
		_, err := Pool.Exec(ctx,
			`INSERT INTO currency_rates (quote, rate, fetched_at) VALUES ($1, $2, now())
			 ON CONFLICT (quote) DO UPDATE SET rate = $2, fetched_at = now()`,
			quote, rate,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRates загружает закэшированные курсы (для фолбэка при недоступности API).
func GetRates(ctx context.Context) (map[string]float64, error) {
	rows, err := Pool.Query(ctx, "SELECT quote, rate FROM currency_rates")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var q string
		var r float64
		if err := rows.Scan(&q, &r); err != nil {
			return nil, err
		}
		out[q] = r
	}
	return out, nil
}
