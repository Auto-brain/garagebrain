package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrLinkTokenInvalid — токен не найден, уже использован или просрочен.
var ErrLinkTokenInvalid = errors.New("link token invalid or expired")

// ConsumeLinkAndAttachTelegram гасит одноразовый токен связывания и привязывает
// Telegram-аккаунт (telegramID) к веб-пользователю, на которого выписан токен.
// Если этот Telegram уже принадлежит отдельному (авто-созданному) аккаунту —
// сливает его данные в веб-аккаунт (фундамент для переноса авто между юзерами).
// Всё в одной транзакции.
func ConsumeLinkAndAttachTelegram(ctx context.Context, token, telegramID, username, displayName string) error {
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1) Забираем и гасим токен (FOR UPDATE, проверка TTL и повторного использования).
	var webUserID uuid.UUID
	err = tx.QueryRow(ctx,
		`UPDATE account_link_tokens
		 SET used_at = now()
		 WHERE token = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING user_id`,
		token,
	).Scan(&webUserID)
	if err != nil {
		return ErrLinkTokenInvalid
	}

	// 2) Есть ли уже identity для этого Telegram?
	var existingUserID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT user_id FROM user_identities WHERE platform = 'telegram' AND platform_id = $1`,
		telegramID,
	).Scan(&existingUserID)

	switch {
	case err != nil:
		// identity нет → просто привязываем Telegram к веб-аккаунту.
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_identities (user_id, platform, platform_id, username, display_name)
			 VALUES ($1, 'telegram', $2, $3, $4)`,
			webUserID, telegramID, username, displayName,
		); err != nil {
			return err
		}
	case existingUserID == webUserID:
		// Уже связаны — идемпотентно ничего не делаем.
	default:
		// Telegram принадлежит другому аккаунту → сливаем его в веб-аккаунт.
		if err := mergeUsersTx(ctx, tx, existingUserID, webUserID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// mergeUsersTx переносит все данные пользователя fromID на toID и удаляет fromID.
// service_records/fuel_records/reminders привязаны к cars (по car_id), поэтому
// переноса cars достаточно — записи уезжают вместе с авто.
func mergeUsersTx(ctx context.Context, tx pgx.Tx, fromID, toID uuid.UUID) error {
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
	return nil
}

func GetUserByPlatformID(ctx context.Context, platform, platformID string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := Pool.QueryRow(ctx,
		`SELECT user_id FROM user_identities WHERE platform = $1 AND platform_id = $2`,
		platform, platformID,
	).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func CreateMessengerUser(ctx context.Context) (uuid.UUID, error) {
	var userID uuid.UUID
	err := Pool.QueryRow(ctx,
		"INSERT INTO users DEFAULT VALUES RETURNING id",
	).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func LinkPlatformIdentity(ctx context.Context, userID uuid.UUID, platform, platformID, username, displayName string) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO user_identities (user_id, platform, platform_id, username, display_name)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (platform, platform_id) DO UPDATE SET username = $4, display_name = $5`,
		userID, platform, platformID, username, displayName,
	)
	return err
}

func GetOrCreateUser(ctx context.Context, platform, platformID, username, displayName string) (uuid.UUID, error) {
	userID, err := GetUserByPlatformID(ctx, platform, platformID)
	if err == nil {
		return userID, nil
	}

	userID, err = CreateMessengerUser(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	err = LinkPlatformIdentity(ctx, userID, platform, platformID, username, displayName)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetActiveCar(ctx context.Context, userID uuid.UUID) (*ActiveCar, error) {
	var car ActiveCar
	err := Pool.QueryRow(ctx,
		`SELECT id, brand, model, year, mileage FROM cars WHERE user_id = $1 AND is_active = true LIMIT 1`,
		userID,
	).Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Mileage)
	if err != nil {
		return nil, err
	}
	return &car, nil
}

type ActiveCar struct {
	ID      string
	Brand   string
	Model   string
	Year    *int
	Mileage int
}

func GetUserCars(ctx context.Context, userID uuid.UUID) ([]ActiveCar, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, brand, model, year, mileage FROM cars WHERE user_id = $1 AND is_active = true ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []ActiveCar
	for rows.Next() {
		var c ActiveCar
		err := rows.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.Mileage)
		if err != nil {
			return nil, err
		}
		cars = append(cars, c)
	}
	return cars, nil
}

func GetUserCarCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM cars WHERE user_id = $1 AND is_active = true",
		userID,
	).Scan(&count)
	return count, err
}

func CreateCarFromBot(ctx context.Context, userID uuid.UUID, brand, model string, year *int, mileage int) (string, error) {
	var carID string
	err := Pool.QueryRow(ctx,
		`INSERT INTO cars (user_id, brand, model, year, mileage) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, brand, model, year, mileage,
	).Scan(&carID)
	if err != nil {
		return "", err
	}

	_, err = Pool.Exec(ctx,
		"UPDATE cars SET is_active = false WHERE user_id = $1 AND id != $2",
		userID, carID,
	)
	return carID, err
}

func GetLatestRecords(ctx context.Context, carID string, limit int) ([]RecordRow, error) {
	rows, err := Pool.Query(ctx,
		`SELECT type, title, to_char(date, 'DD.MM.YYYY'), mileage, COALESCE(cost, 0) + COALESCE(parts_cost, 0)
		 FROM service_records WHERE car_id = $1 ORDER BY date DESC LIMIT $2`,
		carID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RecordRow
	for rows.Next() {
		var r RecordRow
		err := rows.Scan(&r.Type, &r.Title, &r.Date, &r.Mileage, &r.Cost)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

type RecordRow struct {
	Type    string
	Title   string
	Date    string
	Mileage *int
	Cost    *float64 // service_records.cost — NUMERIC(14,2)
}

// GetDueTelegramReminders возвращает сработавшие напоминания только для
// пользователей с Telegram-идентичностью (push-уведомления им шлёт не бэкенд,
// а этот шлюз). Уже отправленные (last_triggered_at IS NOT NULL) отсекаются —
// иначе ежечасный опрос слал бы их повторно.
func GetDueTelegramReminders(ctx context.Context) ([]ReminderRow, error) {
	rows, err := Pool.Query(ctx,
		`SELECT r.id, r.car_id, r.title, r.type, c.brand, c.model, c.user_id, ui.platform_id
		 FROM reminders r
		 JOIN cars c ON r.car_id = c.id
		 JOIN user_identities ui ON ui.user_id = c.user_id AND ui.platform = 'telegram'
		 WHERE r.is_active = true
		   AND r.last_triggered_at IS NULL
		   AND (
		     (r.type = 'date' AND r.trigger_date <= CURRENT_DATE)
		     OR (r.type = 'mileage' AND r.trigger_mileage <= c.mileage)
		   )`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []ReminderRow
	for rows.Next() {
		var r ReminderRow
		err := rows.Scan(&r.ID, &r.CarID, &r.Title, &r.Type, &r.CarBrand, &r.CarModel, &r.UserID, &r.TelegramChatID)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

type ReminderRow struct {
	ID             string
	CarID          string
	Title          string
	Type           string
	CarBrand       string
	CarModel       string
	UserID         string
	TelegramChatID string
}

func MarkReminderTriggered(ctx context.Context, reminderID string) error {
	_, err := Pool.Exec(ctx, "UPDATE reminders SET last_triggered_at = now() WHERE id = $1", reminderID)
	return err
}

func GetUserIdentityChatID(ctx context.Context, userID string, platform string) (string, error) {
	var platformID string
	err := Pool.QueryRow(ctx,
		"SELECT platform_id FROM user_identities WHERE user_id = $1 AND platform = $2",
		userID, platform,
	).Scan(&platformID)
	return platformID, err
}
