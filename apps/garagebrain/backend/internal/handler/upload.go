package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var storageSvc *service.Storage

func InitUploadHandler(s *service.Storage) {
	storageSvc = s
}

type UploadResponse struct {
	URL      string   `json:"url"`
	RecordID string   `json:"record_id,omitempty"`
	Photos   []string `json:"photos,omitempty"`
}

// UploadPhoto принимает multipart-форму с файлом чека/фото и сохраняет его
// в локальное хранилище VPS (UPLOAD_DIR). Поля формы:
//   - file      (обяз.) — само изображение
//   - car_id    (обяз.) — авто, к которому относится фото (проверяется владелец)
//   - record_id (опц.)  — конкретная запись; "latest" — последняя запись авто
//
// Если record_id указан/равен latest, URL фото добавляется в photos записи.
func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	// До 15 МБ в памяти, остальное во временные файлы.
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		http.Error(w, `{"error":"invalid multipart form"}`, http.StatusBadRequest)
		return
	}

	carID, err := uuid.Parse(r.FormValue("car_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid car id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := authorizeCar(w, r, carID); !ok {
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !service.IsAllowedImage(header.Filename) {
		http.Error(w, `{"error":"unsupported file type"}`, http.StatusBadRequest)
		return
	}

	url, err := storageSvc.Save(carID, header.Filename, file)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}

	resp := UploadResponse{URL: url}

	// Опциональная привязка к записи.
	if rid := r.FormValue("record_id"); rid != "" {
		var recordID uuid.UUID
		if rid == "latest" {
			recordID, err = db.GetLatestRecordID(r.Context(), carID)
			if errors.Is(err, pgx.ErrNoRows) {
				// Записей ещё нет — фото сохранено, но прикреплять не к чему.
				writeJSON(w, http.StatusCreated, resp)
				return
			}
		} else {
			recordID, err = uuid.Parse(rid)
		}
		if err != nil {
			http.Error(w, `{"error":"invalid record id"}`, http.StatusBadRequest)
			return
		}

		photos, err := db.AppendPhotoToRecord(r.Context(), carID, recordID, url)
		if err != nil {
			http.Error(w, `{"error":"record not found"}`, http.StatusNotFound)
			return
		}
		resp.RecordID = recordID.String()
		resp.Photos = photos
	}

	writeJSON(w, http.StatusCreated, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
