package service

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/auto-brain/garagebrain/internal/db"
)

// defaultRatesURL — открытый бесплатный источник курсов (без ключа), база USD.
const defaultRatesURL = "https://open.er-api.com/v6/latest/USD"

// CurrencyService держит курсы валют (per USD) в памяти, обновляет из открытого
// API и кэширует в БД (фолбэк при недоступности API).
type CurrencyService struct {
	url        string
	httpClient *http.Client
	mu         sync.RWMutex
	rates      map[string]float64
}

func NewCurrencyService() *CurrencyService {
	url := os.Getenv("RATES_URL")
	if url == "" {
		url = defaultRatesURL
	}
	return &CurrencyService{
		url:        url,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		rates:      map[string]float64{},
	}
}

type ratesResp struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// Refresh тянет свежие курсы из API и кэширует их (в память + БД).
func (s *CurrencyService) Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r ratesResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if len(r.Rates) == 0 {
		return nil
	}

	s.mu.Lock()
	s.rates = r.Rates
	s.mu.Unlock()

	return db.SaveRates(ctx, r.Rates)
}

// LoadFromDB подгружает последний кэш курсов на старте (если API ещё не опрошен).
func (s *CurrencyService) LoadFromDB(ctx context.Context) {
	m, err := db.GetRates(ctx)
	if err == nil && len(m) > 0 {
		s.mu.Lock()
		s.rates = m
		s.mu.Unlock()
	}
}

// Rates возвращает копию курсов (база USD).
func (s *CurrencyService) Rates() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]float64, len(s.rates))
	for k, v := range s.rates {
		out[k] = v
	}
	return out
}

// Convert переводит сумму из валюты from в to. Если валюта неизвестна или курсы
// ещё не загружены — возвращает сумму как есть (без конверсии).
func (s *CurrencyService) Convert(amount float64, from, to string) float64 {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	if from == "" || to == "" || from == to {
		return amount
	}
	s.mu.RLock()
	rf, rt := s.rates[from], s.rates[to]
	s.mu.RUnlock()
	if rf == 0 || rt == 0 {
		return amount
	}
	return amount / rf * rt
}
