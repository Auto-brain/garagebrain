package handler

import (
	"encoding/json"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/service"
)

type PushSubscribeRequest struct {
	Endpoint        string                 `json:"endpoint"`
	Keys            map[string]string      `json:"keys"`
	ExpirationTime  *int64                 `json:"expirationTime,omitempty"`
}

var pushSvc *service.PushService

func InitPushHandler(svc *service.PushService) {
	pushSvc = svc
}

// VapidKey отдаёт публичный VAPID-ключ, нужный фронтенду для подписки на push.
func VapidKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"public_key": pushSvc.PublicKey()})
}

func SubscribePush(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req webpush.Subscription
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if err := pushSvc.Subscribe(r.Context(), userID, req); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}
