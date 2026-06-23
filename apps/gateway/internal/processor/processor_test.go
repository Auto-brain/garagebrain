package processor

import (
	"testing"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/backend"
)

func TestSplitCommand(t *testing.T) {
	cmd, args := splitCommand("/add@GarageBot Toyota RAV4 2020")
	if cmd != "/add" {
		t.Errorf("cmd = %q, want /add", cmd)
	}
	if len(args) != 3 || args[0] != "Toyota" {
		t.Errorf("args = %v", args)
	}

	cmd, args = splitCommand("/STATUS")
	if cmd != "/status" || len(args) != 0 {
		t.Errorf("cmd=%q args=%v", cmd, args)
	}
}

func TestStripMarkers(t *testing.T) {
	in := "---ЗАПИСЬ---\nТип: fuel\n---КОНЕЦ---\nЗаправил полный бак."
	if got := StripMarkers(in); got != "Заправил полный бак." {
		t.Errorf("StripMarkers = %q", got)
	}
	if got := StripMarkers("---СТАТУС---\nвсё ок\n---КОНЕЦ---"); got != "✅ Готово." {
		t.Errorf("empty-after-strip = %q, want fallback", got)
	}
}

func TestFormatChatResultRecordCard(t *testing.T) {
	res := &backend.ChatResult{
		ParsedType: "record",
		Record:     &backend.ParsedRecord{Type: "service", Title: "Замена масла", Mileage: 87500, Cost: 3800},
		NextAction: "через 10000 км",
	}
	got := FormatChatResult(res)
	for _, want := range []string{"Замена масла", "🔧 Обслуживание", "87500", "3800", "через 10000 км"} {
		if !contains(got, want) {
			t.Errorf("formatted card missing %q:\n%s", want, got)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
