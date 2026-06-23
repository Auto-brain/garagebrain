package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ListFuel(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	records, err := db.GetFuelRecordsByCar(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func GetFuelStats(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	records, err := db.GetFuelRecordsByCar(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(db.ComputeFuelStats(records))
}

func CreateFuel(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFuelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Date == "" || req.Mileage <= 0 {
		http.Error(w, `{"error":"date and mileage required"}`, http.StatusBadRequest)
		return
	}

	car, ok := authorizeCar(w, r, req.CarID)
	if !ok {
		return
	}

	record, err := db.CreateFuelRecord(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	// Заправка двигает текущий пробег вперёд, только если он больше известного
	// (бэкфилл старой заправки не должен откатывать одометр назад).
	if req.Mileage > car.Mileage {
		db.UpdateCarMileage(r.Context(), req.CarID, req.Mileage)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}
