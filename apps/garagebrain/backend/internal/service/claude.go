package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ClaudeService struct {
	apiKey     string
	siteURL    string
	httpClient *http.Client
}

func NewClaudeService() *ClaudeService {
	return &ClaudeService{
		apiKey:     os.Getenv("OPENROUTER_API_KEY"),
		siteURL:    os.Getenv("OPENROUTER_SITE_URL"),
		httpClient: &http.Client{},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
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

	reqBody := chatRequest{
		Model:    "anthropic/claude-haiku-4-5",
		Messages: messages,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	if s.siteURL != "" {
		req.Header.Set("HTTP-Referer", s.siteURL)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}
