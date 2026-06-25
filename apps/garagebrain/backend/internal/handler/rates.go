package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/service"
)

var currencySvc *service.CurrencyService

func InitCurrencyHandler(s *service.CurrencyService) {
	currencySvc = s
}

// GetRates отдаёт текущие курсы (база USD) — для конверсии на фронте при желании.
func GetRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"base": "USD", "rates": currencySvc.Rates()})
}
