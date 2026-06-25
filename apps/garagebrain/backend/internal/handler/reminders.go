package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ListReminders(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	reminders, err := db.GetRemindersByCar(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}

func CreateReminder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Type == "" {
		http.Error(w, `{"error":"title and type required"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCarWrite(w, r, req.CarID); !ok {
		return
	}

	reminder, err := db.CreateReminder(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reminder)
}
