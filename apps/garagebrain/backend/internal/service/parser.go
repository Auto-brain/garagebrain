package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ParsedResponse struct {
	Type       string
	Record     *ParsedRecord
	NextAction string
	Comment    string
	RawText    string
}

type ParsedRecord struct {
	Type       string
	Title      string
	Date       time.Time
	Mileage    *int
	Cost       *int
	Liters     *float64
	Currency   string
	NextAction string
}

func ParseAIResponse(text string) ParsedResponse {
	switch {
	case strings.Contains(text, "---ЗАПИСЬ---"):
		return parseRecord(text)
	case strings.Contains(text, "---СТАТУС---"):
		return ParsedResponse{Type: "status", RawText: text}
	case strings.Contains(text, "---РАСХОДЫ---"):
		return ParsedResponse{Type: "expenses", RawText: text}
	case strings.Contains(text, "---ПАСПОРТ---"):
		return ParsedResponse{Type: "passport", RawText: text}
	default:
		return ParsedResponse{Type: "text", RawText: text}
	}
}

func parseRecord(text string) ParsedResponse {
	re := regexp.MustCompile(`---ЗАПИСЬ---([\s\S]*?)---КОНЕЦ---`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ParsedResponse{Type: "text", RawText: text}
	}

	fields := parseFields(match[1])

	// Валюта: из явного поля «Валюта» либо из символа/кода в строке стоимости.
	currency := detectCurrency(fields["Валюта"])
	if currency == "" {
		currency = detectCurrency(fields["Стоимость"])
	}

	// Стоимость/пробег/литры остаются nil, если AI их не указал —
	// иначе 0 искажает статистику и расход топлива.
	var cost *int
	costStr := strings.TrimSpace(fields["Стоимость"])
	costStr = strings.TrimRight(costStr, " ₽рубRUB")
	costStr = strings.ReplaceAll(costStr, " ", "")
	if c, err := strconv.Atoi(costStr); err == nil {
		cost = &c
	}

	var mileage *int
	mileStr := strings.TrimSpace(fields["Пробег"])
	mileStr = strings.TrimRight(mileStr, " км")
	mileStr = strings.ReplaceAll(mileStr, " ", "")
	if m, err := strconv.Atoi(mileStr); err == nil {
		mileage = &m
	}

	var liters *float64
	litStr := strings.TrimSpace(fields["Литры"])
	litStr = strings.TrimRight(litStr, " лl")
	litStr = strings.ReplaceAll(litStr, " ", "")
	litStr = strings.ReplaceAll(litStr, ",", ".")
	if l, err := strconv.ParseFloat(litStr, 64); err == nil && l > 0 {
		liters = &l
	}

	record := &ParsedRecord{
		Type:       normalizeType(fields["Тип"]),
		Title:      fields["Описание"],
		Date:       parseDate(fields["Дата"]),
		Mileage:    mileage,
		Cost:       cost,
		Liters:     liters,
		Currency:   currency,
		NextAction: fields["Следующее"],
	}

	parts := strings.SplitN(text, "---КОНЕЦ---", 2)
	comment := ""
	if len(parts) > 1 {
		comment = strings.TrimSpace(parts[1])
	}

	return ParsedResponse{
		Type:       "record",
		Record:     record,
		NextAction: record.NextAction,
		Comment:    comment,
	}
}

func parseFields(text string) map[string]string {
	fields := make(map[string]string)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return fields
}

// detectCurrency распознаёт код/символ валюты в строке (BYN/$/€/₽/₴/₸/zł/руб…).
func detectCurrency(s string) string {
	u := strings.ToUpper(s)
	codes := []string{"BYN", "USD", "EUR", "RUB", "UAH", "KZT", "PLN"}
	for _, c := range codes {
		if strings.Contains(u, c) {
			return c
		}
	}
	switch {
	case strings.Contains(s, "₽") || strings.Contains(u, "РУБ"):
		return "RUB"
	case strings.Contains(s, "$"):
		return "USD"
	case strings.Contains(s, "€"):
		return "EUR"
	case strings.Contains(s, "₴") || strings.Contains(u, "ГРН"):
		return "UAH"
	case strings.Contains(s, "₸") || strings.Contains(u, "ТГ"):
		return "KZT"
	case strings.Contains(s, "ZŁ") || strings.Contains(u, "ZL"):
		return "PLN"
	}
	return ""
}

func normalizeType(t string) string {
	switch strings.ToLower(t) {
	case "service", "то", "обслуживание":
		return "service"
	case "repair", "ремонт":
		return "repair"
	case "fuel", "заправка":
		return "fuel"
	default:
		return "other"
	}
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now()
	}

	formats := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
		"02.01",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			if t.Year() == 0 {
				t = t.AddDate(time.Now().Year(), 0, 0)
			}
			return t
		}
	}

	return time.Now()
}

func ExtractComment(text string) string {
	parts := strings.SplitN(text, "---КОНЕЦ---", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return text
}
