package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
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
	Reply      string                `json:"reply"`
	ParsedType string                `json:"parsed_type,omitempty"`
	Record     *service.ParsedRecord `json:"parsed_record,omitempty"`
	NextAction string                `json:"next_action,omitempty"`
}

var claudeSvc *service.ClaudeService

func InitChatHandler() {
	claudeSvc = service.NewClaudeService()
}

func Chat(w http.ResponseWriter, r *http.Request) {
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

	car, ok := authorizeCar(w, r, carID)
	if !ok {
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
				// Пробег двигаем только вперёд (запись может быть задним числом).
				if parsed.Record.Mileage != nil && *parsed.Record.Mileage > car.Mileage {
					db.UpdateCarMileage(r.Context(), carID, *parsed.Record.Mileage)
				}
				// Заправка с указанным объёмом → отдельная fuel-запись для расхода л/100км.
				if parsed.Record.Type == "fuel" && parsed.Record.Liters != nil && parsed.Record.Mileage != nil {
					db.CreateFuelRecord(r.Context(), model.CreateFuelRequest{
						CarID:   carID,
						Date:    parsed.Record.Date.Format("2006-01-02"),
						Mileage: *parsed.Record.Mileage,
						Liters:  parsed.Record.Liters,
						Cost:    parsed.Record.Cost,
					})
				}
				// AI указал «Следующее: …» → создаём напоминание (без дублей).
				createReminderFromNextAction(r.Context(), carID, car.Mileage, parsed.Record.NextAction)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func createRecordFromParsed(carID uuid.UUID, rec *service.ParsedRecord) model.CreateRecordRequest {
	return model.CreateRecordRequest{
		CarID:   carID,
		Type:    rec.Type,
		Title:   rec.Title,
		Date:    rec.Date.Format("2006-01-02"),
		Mileage: rec.Mileage,
		Cost:    rec.Cost,
	}
}

// createReminderFromNextAction разбирает поле «Следующее: …» и заводит
// напоминание (пробеговое или по дате), если интервал распознан.
func createReminderFromNextAction(ctx context.Context, carID uuid.UUID, currentMileage int, nextAction string) {
	pr := service.ParseNextAction(nextAction, currentMileage)
	if pr == nil {
		return
	}

	req := model.CreateReminderRequest{
		CarID:          carID,
		Title:          pr.Title,
		Type:           pr.Type,
		TriggerMileage: pr.TriggerMileage,
	}
	if pr.TriggerDate != nil {
		d := pr.TriggerDate.Format("2006-01-02")
		req.TriggerDate = &d
	}

	db.CreateReminderIfAbsent(ctx, req)
}
