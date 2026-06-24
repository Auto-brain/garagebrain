package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
)

type LinkStartResponse struct {
	Token    string `json:"token"`
	DeepLink string `json:"deep_link,omitempty"`
}

// StartTelegramLink (auth) создаёт одноразовый токен и возвращает deep-link на
// бота. Пользователь открывает ссылку → бот получает /start link_<token> и
// привязывает свой Telegram к этому веб-аккаунту (Вариант A).
func StartTelegramLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	token, err := db.CreateLinkToken(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	resp := LinkStartResponse{Token: token}
	if bot := os.Getenv("TELEGRAM_BOT_USERNAME"); bot != "" {
		resp.DeepLink = "https://t.me/" + bot + "?start=link_" + token
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
