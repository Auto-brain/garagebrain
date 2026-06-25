package db

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// ErrInviteInvalid — инвайт не найден, использован или просрочен.
var ErrInviteInvalid = errors.New("invite code invalid or expired")

// inviteTTL — сколько живёт одноразовый код приглашения в авто.
const inviteTTL = 24 * time.Hour

// Роли участника авто.
const (
	RoleOwner  = "owner"
	RoleDriver = "driver"
	RoleRenter = "renter"
	RoleViewer = "viewer"
)

// ValidMemberRole проверяет, что роль из числа допустимых (owner назначается
// только бэкфиллом/созданием авто, инвайты — driver/renter/viewer).
func ValidMemberRole(role string) bool {
	switch role {
	case RoleOwner, RoleDriver, RoleRenter, RoleViewer:
		return true
	}
	return false
}

// GetCarRole возвращает роль пользователя в авто. Если пользователь не участник
// (или его аренда истекла), возвращает пустую строку и ok=false.
func GetCarRole(ctx context.Context, carID, userID uuid.UUID) (string, bool) {
	var role string
	err := Pool.QueryRow(ctx,
		`SELECT role FROM car_members
		 WHERE car_id = $1 AND user_id = $2
		   AND (expires_at IS NULL OR expires_at > now())`,
		carID, userID,
	).Scan(&role)
	if err != nil {
		return "", false
	}
	return role, true
}

// AddCarMember добавляет (или обновляет роль) участника авто.
func AddCarMember(ctx context.Context, carID, userID uuid.UUID, role string, invitedBy uuid.UUID, expiresAt *time.Time) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO car_members (car_id, user_id, role, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (car_id, user_id)
		 DO UPDATE SET role = $3, expires_at = $5`,
		carID, userID, role, invitedBy, expiresAt,
	)
	return err
}

// memberCols / scanMember — единый список колонок и разбор участника с именем/email.
const memberCols = `cm.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), cm.role, cm.created_at, cm.expires_at`

func scanMember(row interface {
	Scan(dest ...any) error
}) (model.CarMember, error) {
	var m model.CarMember
	err := row.Scan(&m.UserID, &m.Name, &m.Email, &m.Role, &m.CreatedAt, &m.ExpiresAt)
	return m, err
}

// ListCarMembers возвращает участников авто (с именем/email), owner — первым.
func ListCarMembers(ctx context.Context, carID uuid.UUID) ([]model.CarMember, error) {
	rows, err := Pool.Query(ctx,
		`SELECT `+memberCols+`
		 FROM car_members cm
		 LEFT JOIN users u ON u.id = cm.user_id
		 WHERE cm.car_id = $1
		 ORDER BY (cm.role = 'owner') DESC, cm.created_at`,
		carID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []model.CarMember{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

// CountCarOwners — сколько owner-ов у авто (чтобы не дать удалить последнего).
func CountCarOwners(ctx context.Context, carID uuid.UUID) (int, error) {
	var n int
	err := Pool.QueryRow(ctx,
		`SELECT count(*) FROM car_members WHERE car_id = $1 AND role = 'owner'`,
		carID,
	).Scan(&n)
	return n, err
}

// RemoveCarMember удаляет участника авто.
func RemoveCarMember(ctx context.Context, carID, userID uuid.UUID) error {
	_, err := Pool.Exec(ctx, `DELETE FROM car_members WHERE car_id = $1 AND user_id = $2`, carID, userID)
	return err
}

// UpdateCarMemberRole меняет роль участника.
func UpdateCarMemberRole(ctx context.Context, carID, userID uuid.UUID, role string) error {
	_, err := Pool.Exec(ctx, `UPDATE car_members SET role = $3 WHERE car_id = $1 AND user_id = $2`, carID, userID, role)
	return err
}

// defaultRentalWindow — срок аренды по умолчанию, если для renter-инвайта
// явный срок не задан (renter без срока не имел бы смысла).
const defaultRentalWindow = 30 * 24 * time.Hour

// CreateCarInvite генерирует одноразовый 6-значный код приглашения в авто.
// memberExpiresAt — срок доступа участника после принятия (для renter кладётся
// в car_members.expires_at). Для renter без явного срока подставляется
// defaultRentalWindow; для остальных ролей игнорируется (доступ бессрочный).
func CreateCarInvite(ctx context.Context, carID uuid.UUID, role string, invitedBy uuid.UUID, memberExpiresAt *time.Time) (string, error) {
	if role == RoleRenter && memberExpiresAt == nil {
		t := time.Now().Add(defaultRentalWindow)
		memberExpiresAt = &t
	}
	if role != RoleRenter {
		memberExpiresAt = nil
	}
	for attempt := 0; attempt < 5; attempt++ {
		n, err := rand.Int(rand.Reader, big.NewInt(1000000))
		if err != nil {
			return "", err
		}
		code := fmt.Sprintf("%06d", n.Int64())
		_, err = Pool.Exec(ctx,
			`INSERT INTO car_invites (code, car_id, role, invited_by, expires_at, member_expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			code, carID, role, invitedBy, time.Now().Add(inviteTTL), memberExpiresAt,
		)
		if err == nil {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not allocate invite code")
}

// AcceptCarInvite гасит код приглашения и добавляет userID участником авто с
// ролью из инвайта (в одной транзакции). Возвращает car_id для ответа.
func AcceptCarInvite(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var (
		carID           uuid.UUID
		role            string
		invitedBy       *uuid.UUID
		memberExpiresAt *time.Time
	)
	err = tx.QueryRow(ctx,
		`UPDATE car_invites SET used_at = now()
		 WHERE code = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING car_id, role, invited_by, member_expires_at`,
		code,
	).Scan(&carID, &role, &invitedBy, &memberExpiresAt)
	if err != nil {
		return uuid.Nil, ErrInviteInvalid
	}

	var inviter uuid.UUID
	if invitedBy != nil {
		inviter = *invitedBy
	}

	// expires_at участника — из инвайта (для renter); для прочих ролей NULL
	// (бессрочно). Так доступ renter автоматически истекает.
	if _, err := tx.Exec(ctx,
		`INSERT INTO car_members (car_id, user_id, role, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (car_id, user_id) DO UPDATE SET role = $3, expires_at = $5`,
		carID, userID, role, inviter, memberExpiresAt,
	); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return carID, nil
}
