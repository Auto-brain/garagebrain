package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockChatServer отвечает в зависимости от запрошенной модели:
// paidModel → 402 (нет баланса), freeModel → 200 с заданным контентом.
func mockChatServer(t *testing.T, paidModel, freeModel, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)

		switch req.Model {
		case paidModel:
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write([]byte(`{"error":{"message":"requires more credits","code":402}}`))
		case freeModel:
			resp := chatResponse{}
			resp.Choices = append(resp.Choices, struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{})
			resp.Choices[0].Message.Content = content
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newTestService(srvURL, primary string, free []string) *ClaudeService {
	return &ClaudeService{
		apiKey:       "test",
		primaryModel: primary,
		chatURL:      srvURL,
		httpClient:   &http.Client{},
		freeModels:   &freeModelCache{models: free, fetchedAt: time.Now(), ttl: time.Hour},
	}
}

func TestChatFallbackOn402(t *testing.T) {
	srv := mockChatServer(t, "paid/model", "free/model", "ответ от бесплатной")
	defer srv.Close()

	svc := newTestService(srv.URL, "paid/model", []string{"free/model"})
	got, err := svc.Chat("system", "привет", nil)
	if err != nil {
		t.Fatalf("ожидался успех через фолбэк, got err: %v", err)
	}
	if got != "ответ от бесплатной" {
		t.Errorf("content = %q", got)
	}
}

func TestChatPrimarySucceeds(t *testing.T) {
	// primary == "free/model" → сервер вернёт 200 сразу, фолбэк не нужен.
	srv := mockChatServer(t, "paid/model", "free/model", "ответ основной")
	defer srv.Close()

	svc := newTestService(srv.URL, "free/model", []string{"unused/model"})
	got, err := svc.Chat("system", "привет", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "ответ основной" {
		t.Errorf("content = %q", got)
	}
}

func TestChatNon402ErrorSurfaces(t *testing.T) {
	// Сервер всегда отвечает 500 → это НЕ 402, фолбэка быть не должно,
	// ошибка возвращается наверх.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	svc := newTestService(srv.URL, "paid/model", []string{"free/model"})
	if _, err := svc.Chat("system", "привет", nil); err == nil {
		t.Fatal("ожидалась ошибка (не 402 → без фолбэка)")
	}
}
