package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/model"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/processor"
)

var proc *processor.Processor

func InitIncomingHandler(p *processor.Processor) {
	proc = p
}

// IncomingWebhook — нормализованная точка входа для мессенджеров без своего
// long-polling клиента (WhatsApp, Viber, Discord). Их адаптер (или внешний
// коннектор) переводит платформенный webhook в model.IncomingMessage, POST'ит
// сюда и получает model.OutgoingMessage[] для доставки своими средствами.
// Это и есть «единый Message Processor» из plan_ecosystem.md — один код на все
// платформы; Telegram использует тот же Processor напрямую через long-polling.
func IncomingWebhook(w http.ResponseWriter, r *http.Request) {
	if proc == nil {
		http.Error(w, `{"error":"processor not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var in model.IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"error":"invalid message"}`, http.StatusBadRequest)
		return
	}
	if in.Platform == "" || in.UserID == "" || in.ChatID == "" {
		http.Error(w, `{"error":"platform, user_id and chat_id required"}`, http.StatusBadRequest)
		return
	}
	if in.ReceivedAt.IsZero() {
		in.ReceivedAt = time.Now()
	}

	out := proc.Process(r.Context(), in)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"messages": out})
}
