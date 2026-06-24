package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const defaultFreeModelsURL = "https://shir-man.com/api/free-llm/top-models"

// hardcodedFreeModels — резерв на случай, если эндпоинт со списком недоступен.
// Это бесплатные (`:free`) модели OpenRouter, не требующие баланса.
var hardcodedFreeModels = []string{
	"nvidia/nemotron-3-super-120b-a12b:free",
	"openai/gpt-oss-20b:free",
	"google/gemma-4-31b-it:free",
	"openrouter/free",
}

type freeModelsResp struct {
	Fallback struct {
		ID string `json:"id"`
	} `json:"fallback"`
	Models []struct {
		ID string `json:"id"`
	} `json:"models"`
}

// freeModelCache подтягивает упорядоченный по рангу список бесплатных моделей
// (https://shir-man.com/api/free-llm/top-models) и кэширует его на TTL.
type freeModelCache struct {
	url        string
	httpClient *http.Client
	ttl        time.Duration

	mu        sync.Mutex
	models    []string
	fetchedAt time.Time
}

func newFreeModelCache() *freeModelCache {
	url := os.Getenv("FREE_MODELS_URL")
	if url == "" {
		url = defaultFreeModelsURL
	}
	return &freeModelCache{
		url:        url,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		ttl:        time.Hour,
	}
}

// list возвращает бесплатные модели по убыванию ранга + управляемый OpenRouter
// fallback (`openrouter/free`) в конце. При ошибке сети отдаёт прошлый удачный
// результат либо захардкоженный дефолт — чтобы фолбэк работал всегда.
func (c *freeModelCache) list(ctx context.Context) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.models) > 0 && time.Since(c.fetchedAt) < c.ttl {
		return c.models
	}

	models, err := c.fetch(ctx)
	if err != nil || len(models) == 0 {
		if len(c.models) > 0 {
			return c.models
		}
		return hardcodedFreeModels
	}
	c.models = models
	c.fetchedAt = time.Now()
	return models
}

func (c *freeModelCache) fetch(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("free-models endpoint returned %d", resp.StatusCode)
	}

	var fr freeModelsResp
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(fr.Models)+1)
	for _, m := range fr.Models {
		if m.ID != "" {
			out = append(out, m.ID)
		}
	}
	if fr.Fallback.ID != "" {
		out = append(out, fr.Fallback.ID)
	}
	return out, nil
}
