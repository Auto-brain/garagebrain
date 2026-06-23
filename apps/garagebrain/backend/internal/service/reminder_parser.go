package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParsedReminder — намерение напоминания, извлечённое из поля
// "Следующее: …" ответа AI (например «через 10000 км» или «через 6 месяцев»).
type ParsedReminder struct {
	Title          string
	Type           string // mileage | date
	TriggerMileage *int   // для type == "mileage"
	TriggerDate    *time.Time
}

var (
	reKm     = regexp.MustCompile(`(?i)через\s*([\d\s]+)\s*(?:тыс\.?\s*)?км`)
	reKmThou = regexp.MustCompile(`(?i)через\s*([\d.,]+)\s*тыс`)
	reMonths = regexp.MustCompile(`(?i)через\s*(\d+)\s*мес`)
	reDays   = regexp.MustCompile(`(?i)через\s*(\d+)\s*(?:дн|день|дня)`)
	reYears  = regexp.MustCompile(`(?i)через\s*(\d+)\s*(?:год|года|лет)`)
	reDate   = regexp.MustCompile(`(\d{1,2}[./]\d{1,2}[./]\d{4})`)
)

// ParseNextAction превращает строку «Следующее: …» в напоминание.
// currentMileage нужен для относительных пробеговых интервалов («через N км»).
// Возвращает nil, если интервал распознать не удалось.
func ParseNextAction(text string, currentMileage int) *ParsedReminder {
	s := strings.TrimSpace(text)
	if s == "" {
		return nil
	}

	title := s
	if len(title) > 200 {
		title = title[:200]
	}

	// «через 10 тыс км» / «через 1.5 тыс»
	if m := reKmThou.FindStringSubmatch(s); m != nil {
		val := strings.ReplaceAll(strings.TrimSpace(m[1]), ",", ".")
		if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 {
			km := currentMileage + int(f*1000)
			return &ParsedReminder{Title: title, Type: "mileage", TriggerMileage: &km}
		}
	}

	// «через 10000 км»
	if m := reKm.FindStringSubmatch(s); m != nil {
		digits := strings.ReplaceAll(strings.TrimSpace(m[1]), " ", "")
		if n, err := strconv.Atoi(digits); err == nil && n > 0 {
			km := currentMileage + n
			return &ParsedReminder{Title: title, Type: "mileage", TriggerMileage: &km}
		}
	}

	now := time.Now()

	if m := reMonths.FindStringSubmatch(s); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			d := now.AddDate(0, n, 0)
			return &ParsedReminder{Title: title, Type: "date", TriggerDate: &d}
		}
	}

	if m := reYears.FindStringSubmatch(s); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			d := now.AddDate(n, 0, 0)
			return &ParsedReminder{Title: title, Type: "date", TriggerDate: &d}
		}
	}

	if m := reDays.FindStringSubmatch(s); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			d := now.AddDate(0, 0, n)
			return &ParsedReminder{Title: title, Type: "date", TriggerDate: &d}
		}
	}

	// Конкретная дата «15.09.2026»
	if m := reDate.FindStringSubmatch(s); m != nil {
		norm := strings.ReplaceAll(m[1], "/", ".")
		if d, err := time.Parse("02.01.2006", norm); err == nil {
			return &ParsedReminder{Title: title, Type: "date", TriggerDate: &d}
		}
	}

	return nil
}
