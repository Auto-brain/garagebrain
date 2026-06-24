package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// linkTokenTTL — сколько живёт одноразовый токен связывания аккаунтов.
const linkTokenTTL = 10 * time.Minute

// CreateLinkToken генерирует одноразовый токен для привязки Telegram к
// веб-аккаунту userID (Вариант A: deep-link ?start=link_<token>).
func CreateLinkToken(ctx context.Context, userID uuid.UUID) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)

	_, err := Pool.Exec(ctx,
		`INSERT INTO account_link_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, time.Now().Add(linkTokenTTL),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
