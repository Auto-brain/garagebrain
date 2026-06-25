package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ListMembers (любой участник авто) — список участников.
func ListMembers(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}
	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	members, err := db.ListCarMembers(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

type inviteRequest struct {
	Role string `json:"role"`
	// ExpiresInDays — срок аренды для role=renter (опционально; иначе дефолт в БД).
	ExpiresInDays *int `json:"expires_in_days,omitempty"`
}

// InviteMember (только owner) — создаёт одноразовый код приглашения на роль.
func InviteMember(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}
	if _, ok := authorizeCarOwner(w, r, carID); !ok {
		return
	}

	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = db.RoleDriver
	}
	// Owner через инвайт не выдаётся — только driver/renter/viewer.
	if req.Role == db.RoleOwner || !db.ValidMemberRole(req.Role) {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	var memberExpiresAt *time.Time
	if req.Role == db.RoleRenter && req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresInDays) * 24 * time.Hour)
		memberExpiresAt = &t
	}

	userID := middleware.GetUserID(r.Context())
	code, err := db.CreateCarInvite(r.Context(), carID, req.Role, userID, memberExpiresAt)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"code": code, "role": req.Role})
}

type acceptInviteRequest struct {
	Code string `json:"code"`
}

// AcceptInvite (любой авторизованный) — принимает код приглашения и добавляет
// пользователя в участники авто с ролью из инвайта.
func AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, `{"error":"code required"}`, http.StatusBadRequest)
		return
	}

	carID, err := db.AcceptCarInvite(r.Context(), req.Code, userID)
	if err != nil {
		if errors.Is(err, db.ErrInviteInvalid) {
			http.Error(w, `{"error":"код неверный или устарел"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "joined", "car_id": carID.String()})
}

// RemoveMember (только owner) — удаляет участника авто. Нельзя удалить
// последнего owner.
func RemoveMember(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}
	if _, ok := authorizeCarOwner(w, r, carID); !ok {
		return
	}

	if !guardLastOwner(w, r, carID, memberID) {
		return
	}

	if err := db.RemoveCarMember(r.Context(), carID, memberID); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

// ChangeMemberRole (только owner) — меняет роль участника.
func ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}
	if _, ok := authorizeCarOwner(w, r, carID); !ok {
		return
	}

	var req changeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !db.ValidMemberRole(req.Role) {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	// Понижение последнего owner оставило бы авто без владельца.
	if req.Role != db.RoleOwner && !guardLastOwner(w, r, carID, memberID) {
		return
	}

	if err := db.UpdateCarMemberRole(r.Context(), carID, memberID, req.Role); err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "role": req.Role})
}

// guardLastOwner возвращает false (и пишет ответ) если операция убрала бы
// последнего owner у авто. memberID — затрагиваемый участник.
func guardLastOwner(w http.ResponseWriter, r *http.Request, carID, memberID uuid.UUID) bool {
	role, ok := db.GetCarRole(r.Context(), carID, memberID)
	if !ok || role != db.RoleOwner {
		return true // затрагиваемый участник не owner — ограничение неактуально
	}
	owners, err := db.CountCarOwners(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return false
	}
	if owners <= 1 {
		http.Error(w, `{"error":"нельзя удалить последнего владельца"}`, http.StatusBadRequest)
		return false
	}
	return true
}
