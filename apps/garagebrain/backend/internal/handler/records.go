package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ListRecords(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	recType := r.URL.Query().Get("type")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		json.Unmarshal([]byte(l), &limit)
	}

	records, err := db.GetRecordsByCar(r.Context(), carID, recType, limit)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.Title == "" || req.Date == "" {
		http.Error(w, `{"error":"type, title, and date required"}`, http.StatusBadRequest)
		return
	}

	record, err := db.CreateRecord(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}
