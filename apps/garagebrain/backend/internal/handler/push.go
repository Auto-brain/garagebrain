package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SherClockHolmes/webpush"
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
