package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFreeModelCacheFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"fallback": {"id": "openrouter/free"},
			"models": [
				{"id": "nvidia/nemotron:free"},
				{"id": "openai/gpt-oss-20b:free"}
			]
		}`))
	}))
	defer srv.Close()

	c := &freeModelCache{url: srv.URL, httpClient: srv.Client(), ttl: time.Hour}
	got := c.list(context.Background())

	want := []string{"nvidia/nemotron:free", "openai/gpt-oss-20b:free", "openrouter/free"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFreeModelCacheFallbackOnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &freeModelCache{url: srv.URL, httpClient: srv.Client(), ttl: time.Hour}
	got := c.list(context.Background())

	if len(got) != len(hardcodedFreeModels) {
		t.Fatalf("при ошибке эндпоинта ожидался захардкоженный список, got %v", got)
	}
}

func TestFreeModelCacheCaches(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"models":[{"id":"a:free"}]}`))
	}))
	defer srv.Close()

	c := &freeModelCache{url: srv.URL, httpClient: srv.Client(), ttl: time.Hour}
	c.list(context.Background())
	c.list(context.Background())

	if calls != 1 {
		t.Errorf("ожидался 1 запрос к эндпоинту (кэш на TTL), было %d", calls)
	}
}
