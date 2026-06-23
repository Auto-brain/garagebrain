package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StatsResponse struct {
	TotalCost     int            `json:"total_cost"`
	RecordsByType map[string]int `json:"records_by_type"`
	MonthlyCosts  map[string]int `json:"monthly_costs"`
	RecordCount   int            `json:"record_count"`
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	records, err := db.GetExpensesByCar(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	stats := StatsResponse{
		TotalCost:     0,
		RecordsByType: make(map[string]int),
		MonthlyCosts:  make(map[string]int),
		RecordCount:   len(records),
	}

	for _, r := range records {
		if r.Cost != nil {
			stats.TotalCost += *r.Cost
			stats.RecordsByType[r.Type] += *r.Cost
			month := r.Date.Format("2006-01")
			stats.MonthlyCosts[month] += *r.Cost
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
