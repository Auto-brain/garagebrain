package handler

import (
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetPassport(w http.ResponseWriter, r *http.Request) {
	carID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	car, err := db.GetCarByID(r.Context(), carID)
	if err != nil {
		http.Error(w, `{"error":"car not found"}`, http.StatusNotFound)
		return
	}

	records, err := db.GetRecordsByCar(r.Context(), carID, "", 100)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	pdfBytes, err := service.GeneratePassport(*car, records)
	if err != nil {
		http.Error(w, `{"error":"pdf generation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=passport.pdf")
	w.Write(pdfBytes)
}
