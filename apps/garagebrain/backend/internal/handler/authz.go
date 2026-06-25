package handler

import (
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// authorizeCar проверяет, что автомобиль существует и текущий пользователь —
// его участник (любая роль: owner/driver/renter/viewer). При нарушении пишет
// HTTP-ответ и возвращает ok=false — вызывающему достаточно сделать return.
//
// Доступ основан на членстве (car_members), а не только на cars.user_id —
// так одним авто могут пользоваться несколько аккаунтов (TODO #6).
func authorizeCar(w http.ResponseWriter, r *http.Request, carID uuid.UUID) (*model.Car, bool) {
	car, _, ok := authorizeCarRole(w, r, carID)
	return car, ok
}

// authorizeCarRole — как authorizeCar, но дополнительно возвращает роль
// пользователя в авто (для операций, требующих конкретной роли).
func authorizeCarRole(w http.ResponseWriter, r *http.Request, carID uuid.UUID) (*model.Car, string, bool) {
	userID := middleware.GetUserID(r.Context())

	car, err := db.GetCarByID(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"car not found"}`, http.StatusNotFound)
		return nil, "", false
	}

	role, ok := db.GetCarRole(r.Context(), carID, userID)
	if !ok {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return nil, "", false
	}

	return car, role, true
}

// authorizeCarWrite разрешает мутирующие операции (записи/заправки/напоминания,
// пробег, чат-ввод) всем участникам, кроме viewer (только чтение по спецификации).
// owner/driver/renter — могут писать; viewer получает 403.
func authorizeCarWrite(w http.ResponseWriter, r *http.Request, carID uuid.UUID) (*model.Car, bool) {
	car, role, ok := authorizeCarRole(w, r, carID)
	if !ok {
		return nil, false
	}
	if role == db.RoleViewer {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return nil, false
	}
	return car, true
}

// authorizeCarOwner разрешает операцию только владельцу авто (удаление,
// управление участниками). Возвращает car и ok=true только для role=owner.
func authorizeCarOwner(w http.ResponseWriter, r *http.Request, carID uuid.UUID) (*model.Car, bool) {
	car, role, ok := authorizeCarRole(w, r, carID)
	if !ok {
		return nil, false
	}
	if role != db.RoleOwner {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return nil, false
	}
	return car, true
}
