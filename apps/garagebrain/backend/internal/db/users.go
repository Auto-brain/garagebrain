package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// userCols — общий список колонок; country/region могут быть NULL, поэтому COALESCE.
const userCols = "id, email, password_hash, name, COALESCE(country,''), COALESCE(region,''), created_at"

func scanUser(row interface {
	Scan(dest ...any) error
}) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Country, &u.Region, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"SELECT "+userCols+" FROM users WHERE email = $1", email))
}

func CreateUser(ctx context.Context, email, passwordHash, name, country, region string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, name, country, region) VALUES ($1, $2, $3, $4, $5) RETURNING "+userCols,
		email, passwordHash, name, country, region))
}

func GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"SELECT "+userCols+" FROM users WHERE id = $1", id))
}

// UpdateProfile обновляет имя, страну и регион пользователя.
func UpdateProfile(ctx context.Context, id uuid.UUID, name, country, region string) (*model.User, error) {
	return scanUser(Pool.QueryRow(ctx,
		"UPDATE users SET name = $2, country = $3, region = $4 WHERE id = $1 RETURNING "+userCols,
		id, name, country, region))
}
