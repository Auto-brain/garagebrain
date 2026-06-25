package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// ErrLinkCodeInvalid — код не найден, использован или просрочен.
var ErrLinkCodeInvalid = errors.New("link code invalid or expired")

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

// CreateLinkCode генерирует 6-значный код, привязанный к userID (Вариант B:
// /link в боте → пользователь вводит код на вебе). Ретраи на коллизию кода.
func CreateLinkCode(ctx context.Context, userID uuid.UUID) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		n, err := rand.Int(rand.Reader, big.NewInt(1000000))
		if err != nil {
			return "", err
		}
		code := fmt.Sprintf("%06d", n.Int64())
		_, err = Pool.Exec(ctx,
			`INSERT INTO account_link_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`,
			code, userID, time.Now().Add(linkTokenTTL),
		)
		if err == nil {
			return code, nil
		}
		// конфликт PK (код занят активным токеном) — пробуем ещё.
	}
	return "", fmt.Errorf("could not allocate link code")
}

// ConsumeLinkCode гасит код и возвращает user_id, на который он выписан
// (это Telegram-аккаунт, который нужно слить в текущий веб-аккаунт).
func ConsumeLinkCode(ctx context.Context, code string) (uuid.UUID, error) {
	var fromUserID uuid.UUID
	err := Pool.QueryRow(ctx,
		`UPDATE account_link_tokens SET used_at = now()
		 WHERE token = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING user_id`,
		code,
	).Scan(&fromUserID)
	if err != nil {
		return uuid.Nil, ErrLinkCodeInvalid
	}
	return fromUserID, nil
}

// MergeUsers переносит все данные пользователя fromID на toID и удаляет fromID
// (в транзакции). service_records/fuel/reminders едут вместе с cars (по car_id).
func MergeUsers(ctx context.Context, fromID, toID uuid.UUID) error {
	if fromID == toID {
		return nil
	}
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	stmts := []string{
		`UPDATE cars SET user_id = $2 WHERE user_id = $1`,
		`UPDATE sessions SET user_id = $2 WHERE user_id = $1`,
		`UPDATE push_subscriptions SET user_id = $2 WHERE user_id = $1`,
		`UPDATE user_identities SET user_id = $2 WHERE user_id = $1`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(ctx, s, fromID, toID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, fromID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
