// Package backend — тонкий HTTP-клиент к GarageBrain API.
// Шлюз не дублирует бизнес-логику (Claude, парсинг, сохранение записей),
// а вызывает /api/chat бэкенда, подписывая запрос JWT с общим секретом
// (единый JWT экосистемы). Бэкенд сам вызывает Claude, парсит ответ,
// сохраняет запись и обновляет пробег.
package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	jwtSecret  string
	httpClient *http.Client
}

func New() *Client {
	baseURL := os.Getenv("BACKEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3002"
	}
	return &Client{
		baseURL:    baseURL,
		jwtSecret:  os.Getenv("JWT_SECRET"),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// ParsedRecord повторяет поля service.ParsedRecord бэкенда (без json-тегов,
// поэтому ключи совпадают с именами полей Go).
type ParsedRecord struct {
	Type    string `json:"Type"`
	Title   string `json:"Title"`
	Date    string `json:"Date"`
	Mileage int    `json:"Mileage"`
	Cost    int    `json:"Cost"`
}

type ChatResult struct {
	Reply      string        `json:"reply"`
	ParsedType string        `json:"parsed_type"`
	Record     *ParsedRecord `json:"parsed_record"`
	NextAction string        `json:"next_action"`
}

// mintToken подписывает короткоживущий JWT для пользователя (sub=userID),
// который принимает middleware.Auth бэкенда (HS256, общий JWT_SECRET).
func (c *Client) mintToken(userID uuid.UUID) (string, error) {
	if c.jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.jwtSecret))
}

// Chat отправляет сообщение пользователя в /api/chat от его имени.
func (c *Client) Chat(ctx context.Context, userID uuid.UUID, carID, message string) (*ChatResult, error) {
	token, err := c.mintToken(userID)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(map[string]string{
		"car_id":  carID,
		"message": message,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	var result ChatResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type UploadResult struct {
	URL      string   `json:"url"`
	RecordID string   `json:"record_id"`
	Photos   []string `json:"photos"`
}

// UploadPhoto загружает фото (чек/документ) в бэкенд от имени пользователя.
// recordID может быть "latest" — бэкенд прикрепит фото к последней записи авто.
// Единый источник хранения: шлюз не пишет файлы сам, а проксирует в /api/upload.
func (c *Client) UploadPhoto(ctx context.Context, userID uuid.UUID, carID, recordID, filename string, data []byte) (*UploadResult, error) {
	token, err := c.mintToken(userID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("car_id", carID)
	if recordID != "" {
		mw.WriteField("record_id", recordID)
	}
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(data); err != nil {
		return nil, err
	}
	mw.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/upload", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend upload returned %d: %s", resp.StatusCode, string(body))
	}

	var result UploadResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
