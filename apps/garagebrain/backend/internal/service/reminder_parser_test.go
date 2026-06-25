package service

import "testing"

func TestParseNextAction(t *testing.T) {
	cur := 50000

	tests := []struct {
		name        string
		in          string
		wantNil     bool
		wantType    string
		wantMileage int // 0 = don't check
	}{
		{"km plain", "через 10000 км", false, "mileage", 60000},
		{"km spaced", "через 10 000 км", false, "mileage", 60000},
		{"km thousands", "через 10 тыс км", false, "mileage", 60000},
		{"months", "через 6 месяцев", false, "date", 0},
		{"days", "через 30 дней", false, "date", 0},
		{"years", "через 1 год", false, "date", 0},
		{"explicit date", "до 15.09.2026", false, "date", 0},
		{"empty", "", true, "", 0},
		{"unparseable", "когда-нибудь потом", true, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseNextAction(tt.in, cur)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected result, got nil")
			}
			if got.Type != tt.wantType {
				t.Errorf("type = %q, want %q", got.Type, tt.wantType)
			}
			if tt.wantMileage != 0 {
				if got.TriggerMileage == nil || *got.TriggerMileage != tt.wantMileage {
					t.Errorf("trigger mileage = %v, want %d", got.TriggerMileage, tt.wantMileage)
				}
			}
		})
	}
}

func TestParseRecordFuelFields(t *testing.T) {
	text := "---ЗАПИСЬ---\nТип: fuel\nОписание: Заправка АИ-95\nДата: 10.06.2026\nПробег: 51000 км\nСтоимость: 3000 ₽\nЛитры: 45,5 л\n---КОНЕЦ---\nЗаправил полный бак."
	res := ParseAIResponse(text)
	if res.Type != "record" || res.Record == nil {
		t.Fatalf("expected record, got %+v", res)
	}
	if res.Record.Liters == nil || *res.Record.Liters != 45.5 {
		t.Errorf("liters = %v, want 45.5", res.Record.Liters)
	}
	if res.Record.Cost == nil || *res.Record.Cost != 3000 {
		t.Errorf("cost = %v, want 3000", res.Record.Cost)
	}
	if res.Record.Mileage == nil || *res.Record.Mileage != 51000 {
		t.Errorf("mileage = %v, want 51000", res.Record.Mileage)
	}
}

func TestParseRecordOmittedFieldsAreNil(t *testing.T) {
	text := "---ЗАПИСЬ---\nТип: repair\nОписание: Подтянул ручник\nДата: 10.06.2026\n---КОНЕЦ---\nГотово."
	res := ParseAIResponse(text)
	if res.Record == nil {
		t.Fatal("expected record")
	}
	if res.Record.Cost != nil {
		t.Errorf("cost should be nil, got %v", *res.Record.Cost)
	}
	if res.Record.Mileage != nil {
		t.Errorf("mileage should be nil, got %v", *res.Record.Mileage)
	}
}

func TestParseRecordCurrency(t *testing.T) {
	cases := map[string]string{
		"Стоимость: 3500 BYN": "BYN",
		"Стоимость: $50":      "USD",
		"Стоимость: 20€":      "EUR",
		"Стоимость: 3000 ₽":   "RUB",
		"Стоимость: 1000":     "",
	}
	for costLine, want := range cases {
		text := "---ЗАПИСЬ---\nТип: service\nОписание: X\nДата: 10.06.2026\n" + costLine + "\n---КОНЕЦ---\nok"
		got := ParseAIResponse(text)
		if got.Record == nil {
			t.Fatalf("no record for %q", costLine)
		}
		if got.Record.Currency != want {
			t.Errorf("%q -> currency %q, want %q", costLine, got.Record.Currency, want)
		}
	}
}
