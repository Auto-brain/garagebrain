package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StatsResponse struct {
	TotalCost     float64            `json:"total_cost"`
	RecordsByType map[string]float64 `json:"records_by_type"`
	MonthlyCosts  map[string]float64 `json:"monthly_costs"`
	RecordCount   int                `json:"record_count"`
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
		RecordsByType: make(map[string]float64),
		MonthlyCosts:  make(map[string]float64),
		RecordCount:   len(records),
	}

	for _, r := range records {
		amount := 0.0
		if r.Cost != nil {
			amount += *r.Cost
		}
		if r.PartsCost != nil {
			amount += *r.PartsCost
		}
		if amount == 0 {
			continue
		}
		stats.TotalCost += amount
		stats.RecordsByType[r.Type] += amount
		month := r.Date.Format("2006-01")
		stats.MonthlyCosts[month] += amount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
