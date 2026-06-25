package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// userCols — общий список колонок. У мессенджер-пользователей email/name/
// password_hash могут быть NULL, поэтому всё текстовое заворачиваем в COALESCE
// (иначе scan NULL→string падает и юзер «не находится»).
const userCols = "id, COALESCE(email,''), COALESCE(password_hash,''), COALESCE(name,''), COALESCE(country,''), COALESCE(region,''), COALESCE(currency,''), COALESCE(language,''), created_at"

func scanUser(row interface {
	Scan(dest ...any) error
}) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Country, &u.Region, &u.Currency, &u.Language, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"SELECT "+userCols+" FROM users WHERE email = $1", email))
}

func CreateUser(ctx context.Context, email, passwordHash, name, country, region, currency, language string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, name, country, region, currency, language) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING "+userCols,
		email, passwordHash, name, country, region, currency, language))
}

func GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"SELECT "+userCols+" FROM users WHERE id = $1", id))
}

// UpdateProfile обновляет имя, страну, регион, валюту и язык пользователя.
func UpdateProfile(ctx context.Context, id uuid.UUID, name, country, region, currency, language string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"UPDATE users SET name = $2, country = $3, region = $4, currency = $5, language = $6 WHERE id = $1 RETURNING "+userCols,
		id, name, country, region, currency, language))
}
