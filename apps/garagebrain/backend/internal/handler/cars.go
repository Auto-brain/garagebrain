package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ListCars(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	cars, err := db.GetCarsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cars)
}

func CreateCar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req model.CreateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Brand == "" || req.Model == "" {
		http.Error(w, `{"error":"brand and model required"}`, http.StatusBadRequest)
		return
	}

	car, err := db.CreateCar(r.Context(), userID, req)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func GetCar(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	car, ok := authorizeCar(w, r, carID)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(car)
}

func UpdateCar(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	var req model.UpdateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Brand == "" || req.Model == "" {
		http.Error(w, `{"error":"brand and model required"}`, http.StatusBadRequest)
		return
	}

	car, err := db.UpdateCar(r.Context(), carID, req)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(car)
}

func UpdateMileage(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	var req model.UpdateMileageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if err := db.UpdateCarMileage(r.Context(), carID, req.Mileage); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
