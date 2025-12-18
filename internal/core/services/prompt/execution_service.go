package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/pkg/errors"
)

// LLMProvider represents the LLM provider type.
type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderAnthropic LLMProvider = "anthropic"
)

// LLMClientConfig holds the configuration for LLM clients.
type LLMClientConfig struct {
	OpenAIAPIKey      string
	OpenAIBaseURL     string
	AnthropicAPIKey   string
	AnthropicBaseURL  string
	DefaultTimeout    time.Duration
}

// executionService implements promptDomain.ExecutionService.
type executionService struct {
	compiler   promptDomain.CompilerService
	config     *LLMClientConfig
	httpClient *http.Client
}

// NewExecutionService creates a new execution service instance.
func NewExecutionService(compiler promptDomain.CompilerService, config *LLMClientConfig) promptDomain.ExecutionService {
	timeout := config.DefaultTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &executionService{
		compiler: compiler,
		config:   config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Execute executes a prompt with the configured LLM.
func (s *executionService) Execute(ctx context.Context, prompt *promptDomain.PromptResponse, variables map[string]string, configOverrides *promptDomain.ModelConfig) (*promptDomain.ExecutePromptResponse, error) {
	startTime := time.Now()

	compiled, err := s.compiler.Compile(prompt.Template, prompt.Type, variables)
	if err != nil {
		return &promptDomain.ExecutePromptResponse{
			CompiledPrompt: nil,
			LatencyMs:      time.Since(startTime).Milliseconds(),
			Error:          fmt.Sprintf("failed to compile template: %v", err),
		}, nil
	}

	effectiveConfig := s.mergeConfig(prompt.Config, configOverrides)
	if effectiveConfig == nil || effectiveConfig.Model == "" {
		return &promptDomain.ExecutePromptResponse{
			CompiledPrompt: compiled,
			LatencyMs:      time.Since(startTime).Milliseconds(),
			Error:          "no model specified in config",
		}, nil
	}

	provider := s.detectProvider(effectiveConfig.Model)

	var llmResp *promptDomain.LLMResponse
	switch provider {
	case ProviderOpenAI:
		llmResp, err = s.executeOpenAI(ctx, prompt.Type, compiled, effectiveConfig)
	case ProviderAnthropic:
		llmResp, err = s.executeAnthropic(ctx, prompt.Type, compiled, effectiveConfig)
	default:
		return &promptDomain.ExecutePromptResponse{
			CompiledPrompt: compiled,
			LatencyMs:      time.Since(startTime).Milliseconds(),
			Error:          fmt.Sprintf("unsupported provider for model: %s", effectiveConfig.Model),
		}, nil
	}

	latencyMs := time.Since(startTime).Milliseconds()

	if err != nil {
		return &promptDomain.ExecutePromptResponse{
			CompiledPrompt: compiled,
			LatencyMs:      latencyMs,
			Error:          err.Error(),
		}, nil
	}

	return &promptDomain.ExecutePromptResponse{
		CompiledPrompt: compiled,
		Response:       llmResp,
		LatencyMs:      latencyMs,
	}, nil
}

// Preview compiles and returns the prompt without executing.
func (s *executionService) Preview(ctx context.Context, prompt *promptDomain.PromptResponse, variables map[string]string) (interface{}, error) {
	return s.compiler.Compile(prompt.Template, prompt.Type, variables)
}

// mergeConfig merges the base config with overrides.
func (s *executionService) mergeConfig(base, overrides *promptDomain.ModelConfig) *promptDomain.ModelConfig {
	if base == nil && overrides == nil {
		return nil
	}
	if base == nil {
		return overrides
	}
	if overrides == nil {
		return base
	}

	result := &promptDomain.ModelConfig{
		Model:            base.Model,
		Temperature:      base.Temperature,
		MaxTokens:        base.MaxTokens,
		TopP:             base.TopP,
		FrequencyPenalty: base.FrequencyPenalty,
		PresencePenalty:  base.PresencePenalty,
		Stop:             base.Stop,
	}

	if overrides.Model != "" {
		result.Model = overrides.Model
	}
	if overrides.Temperature != nil {
		result.Temperature = overrides.Temperature
	}
	if overrides.MaxTokens != nil {
		result.MaxTokens = overrides.MaxTokens
	}
	if overrides.TopP != nil {
		result.TopP = overrides.TopP
	}
	if overrides.FrequencyPenalty != nil {
		result.FrequencyPenalty = overrides.FrequencyPenalty
	}
	if overrides.PresencePenalty != nil {
		result.PresencePenalty = overrides.PresencePenalty
	}
	if len(overrides.Stop) > 0 {
		result.Stop = overrides.Stop
	}

	return result
}

// detectProvider determines the LLM provider from the model name.
func (s *executionService) detectProvider(model string) LLMProvider {
	model = strings.ToLower(model)

	if strings.HasPrefix(model, "claude") {
		return ProviderAnthropic
	}

	// OpenAI: gpt-*, o1-*, text-*, davinci*, chatgpt*
	if strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "text-") || strings.HasPrefix(model, "davinci") ||
		strings.HasPrefix(model, "chatgpt") {
		return ProviderOpenAI
	}

	return ProviderOpenAI // Default fallback
}

// ----------------------------
// OpenAI Integration
// ----------------------------

type openAIRequest struct {
	Model            string           `json:"model"`
	Messages         []openAIMessage  `json:"messages,omitempty"`
	Prompt           string           `json:"prompt,omitempty"`
	Temperature      *float64         `json:"temperature,omitempty"`
	MaxTokens        *int             `json:"max_tokens,omitempty"`
	TopP             *float64         `json:"top_p,omitempty"`
	FrequencyPenalty *float64         `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64         `json:"presence_penalty,omitempty"`
	Stop             []string         `json:"stop,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Text         string `json:"text,omitempty"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (s *executionService) executeOpenAI(ctx context.Context, promptType promptDomain.PromptType, compiled interface{}, config *promptDomain.ModelConfig) (*promptDomain.LLMResponse, error) {
	if s.config.OpenAIAPIKey == "" {
		return nil, errors.NewValidationError("OPENAI_API_KEY not configured", "")
	}

	baseURL := s.config.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req := openAIRequest{
		Model:            config.Model,
		Temperature:      config.Temperature,
		MaxTokens:        config.MaxTokens,
		TopP:             config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
		Stop:             config.Stop,
	}

	var endpoint string
	switch promptType {
	case promptDomain.PromptTypeChat:
		messages, ok := compiled.([]promptDomain.ChatMessage)
		if !ok {
			return nil, errors.NewValidationError("invalid compiled chat messages", "")
		}
		req.Messages = make([]openAIMessage, len(messages))
		for i, msg := range messages {
			req.Messages[i] = openAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		endpoint = baseURL + "/chat/completions"

	case promptDomain.PromptTypeText:
		text, ok := compiled.(string)
		if !ok {
			return nil, errors.NewValidationError("invalid compiled text prompt", "")
		}
		// Text prompts use chat API with user role
		req.Messages = []openAIMessage{
			{Role: "user", Content: text},
		}
		endpoint = baseURL + "/chat/completions"

	default:
		return nil, errors.NewValidationError("unsupported prompt type: "+string(promptType), "")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.OpenAIAPIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (%s)", openAIResp.Error.Message, openAIResp.Error.Type)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	var content string
	if openAIResp.Choices[0].Message.Content != "" {
		content = openAIResp.Choices[0].Message.Content
	} else {
		content = openAIResp.Choices[0].Text
	}

	cost := s.calculateOpenAICost(config.Model, openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens)

	return &promptDomain.LLMResponse{
		Content: content,
		Model:   openAIResp.Model,
		Usage: &promptDomain.LLMUsage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
		Cost: &cost,
	}, nil
}

// ----------------------------
// Anthropic Integration
// ----------------------------

type anthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []anthropicMessage  `json:"messages"`
	System      string              `json:"system,omitempty"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature *float64            `json:"temperature,omitempty"`
	TopP        *float64            `json:"top_p,omitempty"`
	StopSeq     []string            `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *executionService) executeAnthropic(ctx context.Context, promptType promptDomain.PromptType, compiled interface{}, config *promptDomain.ModelConfig) (*promptDomain.LLMResponse, error) {
	if s.config.AnthropicAPIKey == "" {
		return nil, errors.NewValidationError("ANTHROPIC_API_KEY not configured", "")
	}

	baseURL := s.config.AnthropicBaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	maxTokens := 4096
	if config.MaxTokens != nil {
		maxTokens = *config.MaxTokens
	}

	req := anthropicRequest{
		Model:       config.Model,
		MaxTokens:   maxTokens,
		Temperature: config.Temperature,
		TopP:        config.TopP,
		StopSeq:     config.Stop,
	}

	switch promptType {
	case promptDomain.PromptTypeChat:
		messages, ok := compiled.([]promptDomain.ChatMessage)
		if !ok {
			return nil, errors.NewValidationError("invalid compiled chat messages", "")
		}

		// Anthropic uses separate system field instead of system role message
		var anthropicMsgs []anthropicMessage
		for _, msg := range messages {
			if msg.Role == "system" {
				req.System = msg.Content
				continue
			}
			anthropicMsgs = append(anthropicMsgs, anthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
		req.Messages = anthropicMsgs

	case promptDomain.PromptTypeText:
		text, ok := compiled.(string)
		if !ok {
			return nil, errors.NewValidationError("invalid compiled text prompt", "")
		}
		req.Messages = []anthropicMessage{
			{Role: "user", Content: text},
		}

	default:
		return nil, errors.NewValidationError("unsupported prompt type: "+string(promptType), "")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.config.AnthropicAPIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("Anthropic API error: %s (%s)", anthropicResp.Error.Message, anthropicResp.Error.Type)
	}

	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	cost := s.calculateAnthropicCost(config.Model, anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)

	return &promptDomain.LLMResponse{
		Content: content,
		Model:   anthropicResp.Model,
		Usage: &promptDomain.LLMUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
		Cost: &cost,
	}, nil
}

// ----------------------------
// Cost Calculation
// ----------------------------

// calculateOpenAICost estimates the cost for OpenAI API calls.
// Prices are in USD per 1M tokens (simplified).
func (s *executionService) calculateOpenAICost(model string, promptTokens, completionTokens int) float64 {
	// Simplified pricing - in production, pull from analytics domain
	var inputPrice, outputPrice float64

	switch {
	case strings.HasPrefix(model, "gpt-4o-mini"):
		inputPrice, outputPrice = 0.15, 0.60
	case strings.HasPrefix(model, "gpt-4o"):
		inputPrice, outputPrice = 2.50, 10.00
	case strings.HasPrefix(model, "gpt-4-turbo"), strings.HasPrefix(model, "gpt-4-1106"):
		inputPrice, outputPrice = 10.00, 30.00
	case strings.HasPrefix(model, "gpt-4"):
		inputPrice, outputPrice = 30.00, 60.00
	case strings.HasPrefix(model, "gpt-3.5-turbo"):
		inputPrice, outputPrice = 0.50, 1.50
	case strings.HasPrefix(model, "o1-mini"):
		inputPrice, outputPrice = 3.00, 12.00
	case strings.HasPrefix(model, "o1"):
		inputPrice, outputPrice = 15.00, 60.00
	default:
		inputPrice, outputPrice = 1.00, 2.00 // Default fallback
	}

	return (float64(promptTokens)*inputPrice + float64(completionTokens)*outputPrice) / 1_000_000
}

// calculateAnthropicCost estimates the cost for Anthropic API calls.
// Prices are in USD per 1M tokens.
func (s *executionService) calculateAnthropicCost(model string, inputTokens, outputTokens int) float64 {
	var inputPrice, outputPrice float64

	switch {
	case strings.Contains(model, "claude-3-5-sonnet"), strings.Contains(model, "claude-sonnet-4"):
		inputPrice, outputPrice = 3.00, 15.00
	case strings.Contains(model, "claude-3-5-haiku"), strings.Contains(model, "claude-haiku-3-5"):
		inputPrice, outputPrice = 1.00, 5.00
	case strings.Contains(model, "claude-3-opus"), strings.Contains(model, "claude-opus"):
		inputPrice, outputPrice = 15.00, 75.00
	case strings.Contains(model, "claude-3-haiku"):
		inputPrice, outputPrice = 0.25, 1.25
	case strings.Contains(model, "claude-3"):
		inputPrice, outputPrice = 3.00, 15.00
	default:
		inputPrice, outputPrice = 3.00, 15.00 // Default fallback
	}

	return (float64(inputTokens)*inputPrice + float64(outputTokens)*outputPrice) / 1_000_000
}
