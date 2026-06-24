package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// defaultPrimaryModel — основная (платная) модель по умолчанию. Переопределяется
// переменной CLAUDE_MODEL (напр. на бесплатную модель на время тестирования).
// При нехватке баланса (402) сервис всё равно уходит в фолбэк (см. freeModelCache).
const defaultPrimaryModel = "anthropic/claude-haiku-4-5"

// defaultMaxTokens ограничивает ответ модели. Без него OpenRouter резервирует
// максимум модели (у claude-haiku-4-5 — 64000), из-за чего на бесплатном/малом
// балансе запрос отклоняется с 402 (не хватает кредитов на бронь токенов).
const defaultMaxTokens = 2000

type ClaudeService struct {
	apiKey       string
	siteURL      string
	primaryModel string
	httpClient   *http.Client
	freeModels   *freeModelCache
}

func NewClaudeService() *ClaudeService {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = defaultPrimaryModel
	}
	return &ClaudeService{
		apiKey:       os.Getenv("OPENROUTER_API_KEY"),
		siteURL:      os.Getenv("OPENROUTER_SITE_URL"),
		primaryModel: model,
		httpClient:   &http.Client{},
		freeModels:   newFreeModelCache(),
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *ClaudeService) Chat(systemPrompt string, userMessage string, conversationHistory []string) (string, error) {
	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
	}
	for _, msg := range conversationHistory {
		messages = append(messages, chatMessage{Role: "user", Content: msg})
	}
	messages = append(messages, chatMessage{Role: "user", Content: userMessage})

	// 1) Основная модель.
	content, status, err := s.callModel(s.primaryModel, messages)
	if err == nil {
		return content, nil
	}
	// Переключаемся на бесплатные ТОЛЬКО при нехватке баланса (402). Любую другую
	// ошибку основной модели возвращаем как есть.
	if status != http.StatusPaymentRequired {
		return "", err
	}
	log.Printf("claude: основная модель %s недоступна по балансу (402), переключаюсь на бесплатные", s.primaryModel)

	// 2) Бесплатные модели по рангу, затем openrouter/free.
	for _, m := range s.freeModels.list(context.Background()) {
		if m == s.primaryModel {
			continue
		}
		content, status, err = s.callModel(m, messages)
		if err == nil {
			log.Printf("claude: ответ получен через бесплатную модель %s", m)
			return content, nil
		}
		log.Printf("claude: бесплатная модель %s не сработала (status=%d): %v", m, status, err)
	}

	return "", fmt.Errorf("все модели недоступны, последняя ошибка: %w", err)
}

// callModel выполняет один запрос к OpenRouter указанной моделью и возвращает
// (ответ, HTTP-статус, ошибка). Статус нужен вызывающему, чтобы отличить 402
// (нехватка баланса → пробуем другую модель) от прочих ошибок.
func (s *ClaudeService) callModel(model string, messages []chatMessage) (string, int, error) {
	reqBody := chatRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: defaultMaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	if s.siteURL != "" {
		req.Header.Set("HTTP-Referer", s.siteURL)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("openrouter returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", resp.StatusCode, err
	}
	if len(chatResp.Choices) == 0 {
		return "", resp.StatusCode, fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, resp.StatusCode, nil
}
