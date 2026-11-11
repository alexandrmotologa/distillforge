package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/distillforge/internal/config"
)

// TeacherClient defines the interface for querying a teacher model.
type TeacherClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// NewTeacherClient constructs the appropriate teacher client based on configuration.
func NewTeacherClient(cfg *config.TeacherConfig) (TeacherClient, error) {
	switch cfg.Provider {
	case "openai", "custom":
		apiKey := ""
		if cfg.APIKeyEnv != "" {
			apiKey = os.Getenv(cfg.APIKeyEnv)
		}
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return &OpenAICompatibleClient{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			APIKey:      apiKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
			HTTPClient:  &http.Client{Timeout: 60 * time.Second},
		}, nil

	case "anthropic":
		apiKey := ""
		if cfg.APIKeyEnv != "" {
			apiKey = os.Getenv(cfg.APIKeyEnv)
		}
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com/v1"
		}
		return &AnthropicClient{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			APIKey:      apiKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
			HTTPClient:  &http.Client{Timeout: 60 * time.Second},
		}, nil

	case "ollama":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return &OllamaClient{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			HTTPClient:  &http.Client{Timeout: 90 * time.Second},
		}, nil

	case "mock":
		return &MockTeacherClient{ModelName: cfg.Model}, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// OpenAICompatibleClient interacts with OpenAI, vLLM, Together, Groq, or DeepSeek.
type OpenAICompatibleClient struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	HTTPClient  *http.Client
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAICompatibleClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := openAIChatRequest{
		Model:       c.Model,
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse response json: %w", err)
	}

	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("api returned error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response")
	}

	return parsed.Choices[0].Message.Content, nil
}

// AnthropicClient interacts with Claude models.
type AnthropicClient struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	HTTPClient  *http.Client
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	Messages    []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:       c.Model,
		System:      systemPrompt,
		MaxTokens:   c.MaxTokens,
		Temperature: c.Temperature,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/messages", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", err
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", parsed.Error.Message)
	}

	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("empty content in anthropic response")
	}

	return parsed.Content[0].Text, nil
}

// OllamaClient interacts with a local Ollama daemon.
type OllamaClient struct {
	BaseURL     string
	Model       string
	Temperature float64
	HTTPClient  *http.Client
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  map[string]any      `json:"options"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Error string `json:"error,omitempty"`
}

func (c *OllamaClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := ollamaChatRequest{
		Model:  c.Model,
		Stream: false,
		Options: map[string]any{
			"temperature": c.Temperature,
		},
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var parsed ollamaChatResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", err
	}

	if parsed.Error != "" {
		return "", fmt.Errorf("ollama error: %s", parsed.Error)
	}

	return parsed.Message.Content, nil
}

// MockTeacherClient provides realistic synthetic generation for local tests without paid APIs.
type MockTeacherClient struct {
	ModelName string
}

func (m *MockTeacherClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// If it is an evolution prompt, mutate the seed
	if strings.Contains(systemPrompt, "Evol-Instruct") || strings.Contains(userPrompt, "evolve") {
		seed := ""
		if idx := strings.Index(userPrompt, "Input:\n"); idx != -1 {
			seed = strings.TrimSpace(userPrompt[idx+len("Input:\n"):])
		} else if idx := strings.Index(userPrompt, "Original:\n"); idx != -1 {
			seed = strings.TrimSpace(userPrompt[idx+len("Original:\n"):])
		} else {
			seed = strings.TrimSpace(userPrompt)
		}

		if strings.Contains(systemPrompt, "in_depth") {
			return fmt.Sprintf("%s (with high priority constraint and mandatory 4-hour SLA verification)", seed), nil
		} else if strings.Contains(systemPrompt, "in_breadth") {
			return fmt.Sprintf("Regarding service concern: %s. Please review immediately.", seed), nil
		} else if strings.Contains(systemPrompt, "adversarial") {
			return fmt.Sprintf("[URGENT ESCALATION] %s - error code 0x9F41 unexpected fault", seed), nil
		} else if strings.Contains(systemPrompt, "cot") {
			return fmt.Sprintf("<thought>Analyzing ticket requirements and classifying issue.</thought>\n%s", seed), nil
		}
		return fmt.Sprintf("Modified variation: %s", seed), nil
	}

	// If it is a judge/auditor prompt
	if strings.Contains(systemPrompt, "Auditor and Judge") || strings.Contains(systemPrompt, "Dataset Auditor") {
		return `{"instruction_following": 0.95, "format_adherence": 1.0, "groundedness": 0.92, "feedback": "High quality response adhering to instructions."}`, nil
	}

	lower := strings.ToLower(userPrompt)
	if strings.Contains(lower, "invoice") && strings.Contains(lower, "total") {
		return `{"invoice_number":"INV-9021","vendor_name":"Acme Cloud","date":"2024-11-01","total_amount":450.00,"currency":"USD","line_items":[{"description":"Compute","quantity":1,"unit_price":450.00,"amount":450.00}]}`, nil
	}
	if strings.Contains(lower, "sql_query") || strings.Contains(lower, "target_tables") {
		return `{"sql_query":"SELECT * FROM customers WHERE active = true;","target_tables":["customers"],"complexity":"simple","explanation":"Queries active customers"}`, nil
	}

	// Default domain classification response
	cat := "General Inquiry"
	if strings.Contains(lower, "network") || strings.Contains(lower, "outage") || strings.Contains(lower, "internet") || strings.Contains(lower, "fiber") || strings.Contains(lower, "router") {
		cat = "Network Outage"
	} else if strings.Contains(lower, "billing") || strings.Contains(lower, "subscription") || strings.Contains(lower, "invoice") || strings.Contains(lower, "refund") || strings.Contains(lower, "charge") || strings.Contains(lower, "payment") {
		cat = "Billing & Payments"
	} else if strings.Contains(lower, "password") || strings.Contains(lower, "login") || strings.Contains(lower, "auth") || strings.Contains(lower, "2fa") {
		cat = "Authentication & 2FA"
	} else if strings.Contains(lower, "crash") || strings.Contains(lower, "bug") || strings.Contains(lower, "error") || strings.Contains(lower, "defect") {
		cat = "Technical Defect"
	} else if strings.Contains(lower, "post") || strings.Contains(lower, "server") || strings.Contains(lower, "memory") || strings.Contains(lower, "hardware") || strings.Contains(lower, "fan") {
		cat = "Hardware Failure"
	}

	return fmt.Sprintf(`{"category": "%s", "confidence": 0.98, "requires_escalation": false, "sla_hours": 4}`, cat), nil
}
