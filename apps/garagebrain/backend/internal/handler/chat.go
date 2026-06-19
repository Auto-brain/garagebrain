package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/model"
	"github.com/auto-brain/garagebrain/internal/prompt"
	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/google/uuid"
)

type ChatRequest struct {
	CarID   string   `json:"car_id"`
	Message string   `json:"message"`
	History []string `json:"history,omitempty"`
}

type ChatResponse struct {
	Reply      string                 `json:"reply"`
	ParsedType string                 `json:"parsed_type,omitempty"`
	Record     *service.ParsedRecord  `json:"parsed_record,omitempty"`
	NextAction string                 `json:"next_action,omitempty"`
}

var claudeSvc *service.ClaudeService

func InitChatHandler() {
	claudeSvc = service.NewClaudeService()
}

func Chat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, `{"error":"message required"}`, http.StatusBadRequest)
		return
	}

	carID, err := uuid.Parse(req.CarID)
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	car, err := db.GetCarByID(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"car not found"}`, http.StatusNotFound)
		return
	}

	if car.UserID != userID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	records, _ := db.GetRecordsByCar(r.Context(), carID, "", 10)
	reminders, _ := db.GetRemindersByCar(r.Context(), carID)

	systemPrompt := prompt.BuildSystemPrompt(car, records, reminders)

	reply, err := claudeSvc.Chat(systemPrompt, req.Message, req.History)
	if err != nil {
		http.Error(w, `{"error":"ai error"}`, http.StatusInternalServerError)
		return
	}

	parsed := service.ParseAIResponse(reply)

	resp := ChatResponse{
		Reply:      reply,
		ParsedType: parsed.Type,
		NextAction: parsed.NextAction,
	}

	if parsed.Record != nil {
		resp.Record = parsed.Record

		if parsed.Type == "record" && parsed.Record.Title != "" {
			createReq := createRecordFromParsed(carID, parsed.Record)
			if _, err := db.CreateRecord(r.Context(), createReq); err == nil {
				newMileage := parsed.Record.Mileage
				if newMileage > 0 {
					db.UpdateCarMileage(r.Context(), carID, newMileage)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func createRecordFromParsed(carID uuid.UUID, rec *service.ParsedRecord) model.CreateRecordRequest {
	return model.CreateRecordRequest{
		CarID:    carID,
		Type:     rec.Type,
		Title:    rec.Title,
		Date:     rec.Date.Format("2006-01-02"),
		Mileage:  &rec.Mileage,
		Cost:     &rec.Cost,
	}
}
