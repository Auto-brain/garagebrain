package handler

import (
	"encoding/json"
	"errors"
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

// CreateLinkCode (auth) выдаёт 6-значный код для текущего аккаунта. Бот вызывает
// это от имени Telegram-пользователя (Вариант B): код показывается в чате, затем
// вводится на вебе.
func CreateLinkCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	code, err := db.CreateLinkCode(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

type ConfirmLinkRequest struct {
	Code string `json:"code"`
}

// ConfirmTelegramLink (auth, веб-пользователь) принимает код из бота и сливает
// Telegram-аккаунт (на который выписан код) в текущий веб-аккаунт.
func ConfirmTelegramLink(w http.ResponseWriter, r *http.Request) {
	webUserID := middleware.GetUserID(r.Context())

	var req ConfirmLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, `{"error":"code required"}`, http.StatusBadRequest)
		return
	}

	fromUserID, err := db.ConsumeLinkCode(r.Context(), req.Code)
	if err != nil {
		if errors.Is(err, db.ErrLinkCodeInvalid) {
			http.Error(w, `{"error":"код неверный или устарел"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	if err := db.MergeUsers(r.Context(), fromUserID, webUserID); err != nil {
		http.Error(w, `{"error":"merge failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "linked"})
}
