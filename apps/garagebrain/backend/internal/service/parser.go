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
	Type        string
	Title       string
	Date        time.Time
	Mileage     int
	Cost        int
	NextAction  string
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

	cost := 0
	costStr := strings.TrimSpace(fields["Стоимость"])
	costStr = strings.TrimRight(costStr, " ₽рубRUB")
	costStr = strings.ReplaceAll(costStr, " ", "")
	if c, err := strconv.Atoi(costStr); err == nil {
		cost = c
	}

	mileage := 0
	mileStr := strings.TrimSpace(fields["Пробег"])
	mileStr = strings.TrimRight(mileStr, " км")
	mileStr = strings.ReplaceAll(mileStr, " ", "")
	if m, err := strconv.Atoi(mileStr); err == nil {
		mileage = m
	}

	record := &ParsedRecord{
		Type:       normalizeType(fields["Тип"]),
		Title:      fields["Описание"],
		Date:       parseDate(fields["Дата"]),
		Mileage:    mileage,
		Cost:       cost,
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

func normalizeType(t string) string {
	switch strings.ToLower(t) {
	case "service", "service", "ТО", "обслуживание":
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
