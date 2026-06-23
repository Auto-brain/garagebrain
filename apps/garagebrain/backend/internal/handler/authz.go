package handler

import (
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/google/uuid"
)

// authorizeCar проверяет, что автомобиль существует и принадлежит текущему
// пользователю. При нарушении пишет HTTP-ответ и возвращает ok=false —
// вызывающему достаточно сделать return.
func authorizeCar(w http.ResponseWriter, r *http.Request, carID uuid.UUID) (*model.Car, bool) {
	userID := middleware.GetUserID(r.Context())

	car, err := db.GetCarByID(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"car not found"}`, http.StatusNotFound)
		return nil, false
	}

	if car.UserID != userID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return nil, false
	}

	return car, true
}
