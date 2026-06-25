package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StatsResponse struct {
	TotalCost     float64            `json:"total_cost"`
	RecordsByType map[string]float64 `json:"records_by_type"`
	MonthlyCosts  map[string]float64 `json:"monthly_costs"`
	RecordCount   int                `json:"record_count"`
	Currency      string             `json:"currency"`
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

	// Валюта отображения = валюта пользователя; все суммы конвертируем в неё,
	// чтобы агрегаты были корректны при разных валютах записей.
	displayCur := ""
	if u, err := db.GetUserByID(r.Context(), middleware.GetUserID(r.Context())); err == nil {
		displayCur = u.Currency
	}

	stats := StatsResponse{
		TotalCost:     0,
		RecordsByType: make(map[string]float64),
		MonthlyCosts:  make(map[string]float64),
		RecordCount:   len(records),
		Currency:      displayCur,
	}

	for _, r := range records {
		amount := 0.0
		if r.Cost != nil {
			amount += convertAmount(*r.Cost, r.Currency, displayCur)
		}
		if r.PartsCost != nil {
			amount += convertAmount(*r.PartsCost, r.PartsCurrency, displayCur)
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

// convertAmount переводит сумму из валюты записи (или валюты пользователя по
// умолчанию, если у записи валюта не указана) в валюту отображения.
func convertAmount(amount float64, recordCur, displayCur string) float64 {
	if currencySvc == nil || displayCur == "" {
		return amount
	}
	from := recordCur
	if from == "" {
		from = displayCur
	}
	return currencySvc.Convert(amount, from, displayCur)
}
