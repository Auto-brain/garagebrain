package db

import (
	"context"
	"fmt"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// recordCols / scanRecord — единый список колонок и разбор строки записи,
// чтобы не дублировать порядок полей по нескольким запросам.
const recordCols = "id, car_id, type, title, description, date, mileage, cost, parts_cost, COALESCE(currency,''), parts, workshop, photos, raw_input, created_at"

func scanRecord(row interface {
	Scan(dest ...any) error
}) (model.ServiceRecord, error) {
	var r model.ServiceRecord
	err := row.Scan(&r.ID, &r.CarID, &r.Type, &r.Title, &r.Description, &r.Date, &r.Mileage, &r.Cost,
		&r.PartsCost, &r.Currency, &r.Parts, &r.Workshop, &r.Photos, &r.RawInput, &r.CreatedAt)
	return r, err
}

func CreateRecord(ctx context.Context, req model.CreateRecordRequest) (*model.ServiceRecord, error) {
	r, err := scanRecord(Pool.QueryRow(ctx,
		`INSERT INTO service_records (car_id, type, title, description, date, mileage, cost, parts_cost, currency, parts, workshop)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),$10,$11)
		 RETURNING `+recordCols,
		req.CarID, req.Type, req.Title, req.Description, req.Date, req.Mileage, req.Cost, req.PartsCost,
		req.Currency, req.Parts, req.Workshop,
	))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func GetRecordsByCar(ctx context.Context, carID uuid.UUID, recType string, limit int) ([]model.ServiceRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	query := "SELECT " + recordCols + " FROM service_records WHERE car_id = $1"
	args := []any{carID}

	if recType != "" {
		query += " AND type = $2"
		args = append(args, recType)
	}
	// Плейсхолдер LIMIT зависит от того, добавили ли мы фильтр по типу
	// (иначе при пустом type ссылка на $3 указывает на несуществующий параметр).
	query += fmt.Sprintf(" ORDER BY date DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []model.ServiceRecord{}
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// AppendPhotoToRecord добавляет URL фото в массив photos записи и
// проверяет, что запись принадлежит указанному авто (защита от IDOR).
// Возвращает обновлённый список фото.
func AppendPhotoToRecord(ctx context.Context, carID, recordID uuid.UUID, url string) ([]string, error) {
	var photos []string
	err := Pool.QueryRow(ctx,
		`UPDATE service_records
		 SET photos = array_append(coalesce(photos, '{}'), $1)
		 WHERE id = $2 AND car_id = $3
		 RETURNING photos`,
		url, recordID, carID,
	).Scan(&photos)
	if err != nil {
		return nil, err
	}
	return photos, nil
}

// GetRecordCarID возвращает car_id записи (для проверки владельца перед правкой).
func GetRecordCarID(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	var carID uuid.UUID
	err := Pool.QueryRow(ctx, "SELECT car_id FROM service_records WHERE id = $1", recordID).Scan(&carID)
	return carID, err
}

// UpdateRecord обновляет редактируемые поля записи обслуживания.
func UpdateRecord(ctx context.Context, recordID uuid.UUID, req model.UpdateRecordRequest) (*model.ServiceRecord, error) {
	r, err := scanRecord(Pool.QueryRow(ctx,
		`UPDATE service_records
		 SET type = $2, title = $3, description = $4, date = $5, mileage = $6, cost = $7,
		     parts_cost = $8, currency = NULLIF($9,'')
		 WHERE id = $1
		 RETURNING `+recordCols,
		recordID, req.Type, req.Title, req.Description, req.Date, req.Mileage, req.Cost, req.PartsCost, req.Currency,
	))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteRecord удаляет запись обслуживания.
func DeleteRecord(ctx context.Context, recordID uuid.UUID) error {
	_, err := Pool.Exec(ctx, "DELETE FROM service_records WHERE id = $1", recordID)
	return err
}

// GetLatestRecordID возвращает id последней по дате записи авто.
// pgx.ErrNoRows — если записей нет.
func GetLatestRecordID(ctx context.Context, carID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := Pool.QueryRow(ctx,
		"SELECT id FROM service_records WHERE car_id = $1 ORDER BY date DESC, created_at DESC LIMIT 1",
		carID,
	).Scan(&id)
	return id, err
}

func GetExpensesByCar(ctx context.Context, carID uuid.UUID) ([]model.ServiceRecord, error) {
	rows, err := Pool.Query(ctx,
		"SELECT "+recordCols+" FROM service_records WHERE car_id = $1 AND (cost IS NOT NULL OR parts_cost IS NOT NULL) ORDER BY date DESC",
		carID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []model.ServiceRecord{}
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
