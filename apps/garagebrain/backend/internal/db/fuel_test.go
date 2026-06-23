package db

import (
	"math"
	"testing"

	"github.com/auto-brain/garagebrain/internal/model"
)

func ptrF(f float64) *float64 { return &f }
func ptrI(i int) *int         { return &i }

func TestComputeFuelStats(t *testing.T) {
	// Полные баки на 50000/50500/51000, по 40 л на каждой дозаправке.
	// Расстояние между полными: 1000 км, залито (на втором и третьем) 80 л.
	records := []model.FuelRecord{
		{Mileage: 50000, Liters: ptrF(40), Cost: ptrI(2000), FullTank: true},
		{Mileage: 50500, Liters: ptrF(40), Cost: ptrI(2000), FullTank: true},
		{Mileage: 51000, Liters: ptrF(40), Cost: ptrI(2000), FullTank: true},
	}

	stats := ComputeFuelStats(records)

	if stats.FillCount != 3 {
		t.Errorf("fill count = %d, want 3", stats.FillCount)
	}
	if stats.TotalDistance != 1000 {
		t.Errorf("distance = %d, want 1000", stats.TotalDistance)
	}
	if stats.TotalCost != 6000 {
		t.Errorf("total cost = %d, want 6000", stats.TotalCost)
	}
	// 80 л / 1000 км * 100 = 8.0 л/100км
	if math.Abs(stats.AvgConsumption-8.0) > 0.001 {
		t.Errorf("avg consumption = %f, want 8.0", stats.AvgConsumption)
	}
}

func TestComputeFuelStatsSkipsPartialTanks(t *testing.T) {
	// Долив (full_tank=false) не должен начинать/закрывать интервал.
	records := []model.FuelRecord{
		{Mileage: 50000, Liters: ptrF(40), FullTank: true},
		{Mileage: 50300, Liters: ptrF(20), FullTank: false},
		{Mileage: 50600, Liters: ptrF(48), FullTank: true},
	}
	stats := ComputeFuelStats(records)
	if stats.TotalDistance != 600 {
		t.Errorf("distance = %d, want 600", stats.TotalDistance)
	}
	// 48 л / 600 км * 100 = 8.0
	if math.Abs(stats.AvgConsumption-8.0) > 0.001 {
		t.Errorf("avg consumption = %f, want 8.0", stats.AvgConsumption)
	}
}

func TestComputeFuelStatsSingleFillNoConsumption(t *testing.T) {
	records := []model.FuelRecord{{Mileage: 50000, Liters: ptrF(40), FullTank: true}}
	stats := ComputeFuelStats(records)
	if stats.AvgConsumption != 0 {
		t.Errorf("avg consumption = %f, want 0 (not enough data)", stats.AvgConsumption)
	}
}
