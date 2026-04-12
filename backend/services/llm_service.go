package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"app-backend/config"
	"app-backend/models"
)

type LLMService struct {
	baseURL string
	model   string
	ctxSize int
	client  *http.Client
}

var (
	llmInstance *LLMService
	llmOnce     sync.Once
)

func GetLLMService(cfg config.Config) *LLMService {
	llmOnce.Do(func() {
		llmInstance = &LLMService{
			baseURL: strings.TrimSuffix(cfg.LLMBaseURL, "/"),
			model:   cfg.LLMModel,
			ctxSize: cfg.LLMCtxSize,
			client: &http.Client{
				Timeout: cfg.LLMTimeout,
			},
		}
	})

	return llmInstance
}

type chatCompletionRequest struct {
	Model       string              `json:"model"`
	Messages    []chatMessage       `json:"messages"`
	Temperature float64             `json:"temperature"`
	Stream      bool                `json:"stream"`
	MaxTokens   int                 `json:"max_tokens"`
	ExtraBody   map[string]any      `json:"extra_body,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
	TopP        float64             `json:"top_p,omitempty"`
	PresenceP   float64             `json:"presence_penalty,omitempty"`
	FrequencyP  float64             `json:"frequency_penalty,omitempty"`
	LogitBias   map[string]float64  `json:"logit_bias,omitempty"`
	Tools       []map[string]any    `json:"tools,omitempty"`
	ToolChoice  map[string]string   `json:"tool_choice,omitempty"`
	Seed        *int                `json:"seed,omitempty"`
	User        string              `json:"user,omitempty"`
	Functions   []map[string]string `json:"functions,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (s *LLMService) GenerateReply(ctx context.Context, history []models.Message, userPrompt string) (string, error) {
	requestMessages := []chatMessage{
		{
			Role: "system",
			Content: "You are a concise, helpful assistant. " +
				"Keep answers practical and grounded.",
		},
	}

	for _, msg := range history {
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		requestMessages = append(requestMessages, chatMessage{Role: msg.Role, Content: msg.Content})
	}

	shouldAppendPrompt := true
	if len(history) > 0 {
		last := history[len(history)-1]
		if last.Role == "user" && strings.TrimSpace(last.Content) == strings.TrimSpace(userPrompt) {
			shouldAppendPrompt = false
		}
	}

	if shouldAppendPrompt {
		requestMessages = append(requestMessages, chatMessage{Role: "user", Content: userPrompt})
	}

	requestMessages = limitMessagesByContext(requestMessages, s.ctxSize)

	payload := chatCompletionRequest{
		Model:       s.model,
		Messages:    requestMessages,
		Temperature: 0.7,
		Stream:      false,
		MaxTokens:   512,
		TopP:        0.9,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := s.baseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("llm request failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var completion chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", errors.New("llm returned no choices")
	}

	return strings.TrimSpace(completion.Choices[0].Message.Content), nil
}

func limitMessagesByContext(messages []chatMessage, ctxSize int) []chatMessage {
	if len(messages) <= 1 || ctxSize <= 0 {
		return messages
	}

	inputBudget := ctxSize - 512
	if inputBudget < 512 {
		inputBudget = ctxSize
	}

	system := messages[0]
	usedTokens := estimateTokens(system.Content) + 8

	recentReverse := make([]chatMessage, 0, len(messages)-1)
	for i := len(messages) - 1; i >= 1; i-- {
		msg := messages[i]
		msgTokens := estimateTokens(msg.Content) + 8

		if usedTokens+msgTokens > inputBudget {
			if len(recentReverse) == 0 {
				recentReverse = append(recentReverse, msg)
			}
			continue
		}

		recentReverse = append(recentReverse, msg)
		usedTokens += msgTokens
	}

	trimmed := make([]chatMessage, 0, len(recentReverse)+1)
	trimmed = append(trimmed, system)
	for i := len(recentReverse) - 1; i >= 0; i-- {
		trimmed = append(trimmed, recentReverse[i])
	}

	return trimmed
}

func estimateTokens(content string) int {
	charCount := len([]rune(content))
	if charCount == 0 {
		return 0
	}

	return (charCount / 4) + 1
}

func (s *LLMService) HealthCheck(ctx context.Context) error {
	healthURL := s.baseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err == nil {
		resp, reqErr := s.client.Do(req)
		if reqErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}

	fallbackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	modelsURL := s.baseURL + "/v1/models"
	modelsReq, err := http.NewRequestWithContext(fallbackCtx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(modelsReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("llm health failed with status %d", resp.StatusCode)
	}

	return nil
}
