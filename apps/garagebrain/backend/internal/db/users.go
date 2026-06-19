package db

import (
	"context"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := Pool.QueryRow(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateUser(ctx context.Context, email, passwordHash, name string) (*model.User, error) {
	var u model.User
	err := Pool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id, email, password_hash, name, created_at",
		email, passwordHash, name,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := Pool.QueryRow(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
