package openai

import (
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/infrastructure/providers"
)

// Test data
var (
	testMessage = providers.ChatMessage{
		Role:    "user",
		Content: "Hello, world!",
		Name:    stringPtr("John"),
	}

	testFunction = providers.Function{
		Name:        "get_weather",
		Description: "Get the current weather",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city and state",
				},
			},
			"required": []string{"location"},
		},
	}

	testTool = providers.Tool{
		Type:     "function",
		Function: testFunction,
	}
)

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestOpenAIProvider_transformChatCompletionRequest(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    *providers.ChatCompletionRequest
		expected func(*testing.T, openai.ChatCompletionRequest)
	}{
		{
			name: "Basic chat completion request",
			input: &providers.ChatCompletionRequest{
				Model:       "gpt-3.5-turbo",
				Messages:    []providers.ChatMessage{testMessage},
				MaxTokens:   intPtr(100),
				Temperature: float64Ptr(0.7),
				TopP:        float64Ptr(1.0),
				N:           intPtr(1),
				Stream:      false,
				Stop:        []string{"\\n"},
				User:        stringPtr("user-123"),
			},
			expected: func(t *testing.T, result openai.ChatCompletionRequest) {
				assert.Equal(t, "gpt-3.5-turbo", result.Model)
				assert.Len(t, result.Messages, 1)
				assert.Equal(t, "user", result.Messages[0].Role)
				assert.Equal(t, "Hello, world!", result.Messages[0].Content)
				assert.Equal(t, "John", result.Messages[0].Name)
				assert.Equal(t, intPtr(100), result.MaxTokens)
				assert.Equal(t, float32(0.7), result.Temperature)
				assert.Equal(t, float32(1.0), result.TopP)
				assert.Equal(t, 1, result.N)
				assert.Equal(t, false, result.Stream)
				assert.Equal(t, []string{"\\n"}, result.Stop)
				assert.Equal(t, "user-123", result.User)
			},
		},
		{
			name: "Chat completion with functions",
			input: &providers.ChatCompletionRequest{
				Model:        "gpt-3.5-turbo",
				Messages:     []providers.ChatMessage{testMessage},
				Functions:    []providers.Function{testFunction},
				FunctionCall: "auto",
			},
			expected: func(t *testing.T, result openai.ChatCompletionRequest) {
				assert.Equal(t, "gpt-3.5-turbo", result.Model)
				assert.Len(t, result.Functions, 1)
				assert.Equal(t, "get_weather", result.Functions[0].Name)
				assert.Equal(t, "Get the current weather", result.Functions[0].Description)
				assert.NotNil(t, result.Functions[0].Parameters)
				assert.Equal(t, "auto", result.FunctionCall)
			},
		},
		{
			name: "Chat completion with tools",
			input: &providers.ChatCompletionRequest{
				Model:      "gpt-4",
				Messages:   []providers.ChatMessage{testMessage},
				Tools:      []providers.Tool{testTool},
				ToolChoice: "auto",
			},
			expected: func(t *testing.T, result openai.ChatCompletionRequest) {
				assert.Equal(t, "gpt-4", result.Model)
				assert.Len(t, result.Tools, 1)
				assert.Equal(t, openai.ToolType("function"), result.Tools[0].Type)
				assert.Equal(t, "get_weather", result.Tools[0].Function.Name)
				assert.Equal(t, "auto", result.ToolChoice)
			},
		},
		{
			name: "Chat completion with response format",
			input: &providers.ChatCompletionRequest{
				Model:    "gpt-4",
				Messages: []providers.ChatMessage{testMessage},
				ResponseFormat: &providers.ResponseFormat{
					Type: "json_object",
				},
			},
			expected: func(t *testing.T, result openai.ChatCompletionRequest) {
				assert.NotNil(t, result.ResponseFormat)
				assert.Equal(t, openai.ChatCompletionResponseFormatType("json_object"), result.ResponseFormat.Type)
			},
		},
		{
			name: "Chat completion with all optional parameters",
			input: &providers.ChatCompletionRequest{
				Model:            "gpt-3.5-turbo",
				Messages:         []providers.ChatMessage{testMessage},
				MaxTokens:        intPtr(150),
				Temperature:      float64Ptr(0.8),
				TopP:             float64Ptr(0.9),
				N:                intPtr(2),
				Stream:           true,
				Stop:             []string{"END"},
				PresencePenalty:  float64Ptr(0.1),
				FrequencyPenalty: float64Ptr(0.2),
				LogitBias:        map[string]interface{}{"token": 0.5},
				User:             stringPtr("test-user"),
				Seed:             intPtr(12345),
			},
			expected: func(t *testing.T, result openai.ChatCompletionRequest) {
				assert.Equal(t, intPtr(150), result.MaxTokens)
				assert.Equal(t, float32(0.8), result.Temperature)
				assert.Equal(t, float32(0.9), result.TopP)
				assert.Equal(t, 2, result.N)
				assert.True(t, result.Stream)
				assert.Equal(t, []string{"END"}, result.Stop)
				assert.Equal(t, float32(0.1), result.PresencePenalty)
				assert.Equal(t, float32(0.2), result.FrequencyPenalty)
				assert.NotNil(t, result.LogitBias)
				assert.Equal(t, "test-user", result.User)
				assert.Equal(t, intPtr(12345), result.Seed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformChatCompletionRequest(tt.input)
			tt.expected(t, result)
		})
	}
}

func TestOpenAIProvider_transformChatMessage(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    *providers.ChatMessage
		expected openai.ChatCompletionMessage
	}{
		{
			name: "Basic message",
			input: &providers.ChatMessage{
				Role:    "user",
				Content: "Hello, world!",
				Name:    stringPtr("John"),
			},
			expected: openai.ChatCompletionMessage{
				Role:    "user",
				Content: "Hello, world!",
				Name:    "John",
			},
		},
		{
			name: "Message with function call",
			input: &providers.ChatMessage{
				Role:    "assistant",
				Content: "",
				FunctionCall: map[string]interface{}{
					"name":      "get_weather",
					"arguments": `{"location": "San Francisco"}`,
				},
			},
			expected: openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: "",
				FunctionCall: &openai.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location": "San Francisco"}`,
				},
			},
		},
		{
			name: "Message with tool calls",
			input: &providers.ChatMessage{
				Role:    "assistant",
				Content: "",
				ToolCalls: []providers.ToolCall{
					{
						ID:   "call_123",
						Type: "function",
						Function: providers.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location": "New York"}`,
						},
					},
				},
			},
			expected: openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: "",
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_123",
						Type: openai.ToolType("function"),
						Function: openai.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location": "New York"}`,
						},
					},
				},
			},
		},
		{
			name: "Tool response message",
			input: &providers.ChatMessage{
				Role:       "tool",
				Content:    "The weather is sunny",
				ToolCallID: stringPtr("call_123"),
			},
			expected: openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    "The weather is sunny",
				ToolCallID: stringPtr("call_123"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformChatMessage(tt.input)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Content, result.Content)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.ToolCallID, result.ToolCallID)
			
			// Compare function calls
			if tt.expected.FunctionCall != nil {
				require.NotNil(t, result.FunctionCall)
				assert.Equal(t, tt.expected.FunctionCall.Name, result.FunctionCall.Name)
				assert.Equal(t, tt.expected.FunctionCall.Arguments, result.FunctionCall.Arguments)
			} else {
				assert.Nil(t, result.FunctionCall)
			}
			
			// Compare tool calls
			assert.Len(t, result.ToolCalls, len(tt.expected.ToolCalls))
			for i, expectedToolCall := range tt.expected.ToolCalls {
				assert.Equal(t, expectedToolCall.ID, result.ToolCalls[i].ID)
				assert.Equal(t, expectedToolCall.Type, result.ToolCalls[i].Type)
				assert.Equal(t, expectedToolCall.Function.Name, result.ToolCalls[i].Function.Name)
				assert.Equal(t, expectedToolCall.Function.Arguments, result.ToolCalls[i].Function.Arguments)
			}
		})
	}
}

func TestOpenAIProvider_transformMessageContent(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "String content",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "Nil content",
			input:    nil,
			expected: "",
		},
		{
			name: "Multimodal content with text",
			input: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "What's in this image?",
				},
				map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": "https://example.com/image.jpg",
					},
				},
			},
			expected: "What's in this image?",
		},
		{
			name: "Multimodal content without text",
			input: []interface{}{
				map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": "https://example.com/image.jpg",
					},
				},
			},
			expected: "",
		},
		{
			name:     "Complex object as JSON",
			input:    map[string]interface{}{"key": "value", "number": 42},
			expected: `{"key":"value","number":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformMessageContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenAIProvider_transformCompletionRequest(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    *providers.CompletionRequest
		expected func(*testing.T, openai.CompletionRequest)
	}{
		{
			name: "Basic completion request",
			input: &providers.CompletionRequest{
				Model:       "text-davinci-003",
				Prompt:      "Once upon a time",
				MaxTokens:   intPtr(100),
				Temperature: float64Ptr(0.7),
				TopP:        float64Ptr(1.0),
				N:           intPtr(1),
				Stream:      false,
			},
			expected: func(t *testing.T, result openai.CompletionRequest) {
				assert.Equal(t, "text-davinci-003", result.Model)
				assert.Equal(t, "Once upon a time", result.Prompt)
				assert.Equal(t, intPtr(100), result.MaxTokens)
				assert.Equal(t, float32(0.7), result.Temperature)
				assert.Equal(t, float32(1.0), result.TopP)
				assert.Equal(t, 1, result.N)
				assert.False(t, result.Stream)
			},
		},
		{
			name: "Completion with all parameters",
			input: &providers.CompletionRequest{
				Model:            "text-davinci-003",
				Prompt:           "Hello",
				MaxTokens:        intPtr(50),
				Temperature:      float64Ptr(0.5),
				TopP:             float64Ptr(0.8),
				N:                intPtr(2),
				Stream:           true,
				Logprobs:         intPtr(5),
				Echo:             true,
				Stop:             []string{"\n"},
				PresencePenalty:  float64Ptr(0.1),
				FrequencyPenalty: float64Ptr(0.2),
				BestOf:           intPtr(3),
				LogitBias:        map[string]interface{}{"token": 1.0},
				User:             stringPtr("test-user"),
				Suffix:           stringPtr(" END"),
			},
			expected: func(t *testing.T, result openai.CompletionRequest) {
				assert.Equal(t, "text-davinci-003", result.Model)
				assert.Equal(t, "Hello", result.Prompt)
				assert.Equal(t, intPtr(50), result.MaxTokens)
				assert.Equal(t, float32(0.5), result.Temperature)
				assert.Equal(t, float32(0.8), result.TopP)
				assert.Equal(t, 2, result.N)
				assert.True(t, result.Stream)
				assert.Equal(t, intPtr(5), result.Logprobs)
				assert.True(t, result.Echo)
				assert.Equal(t, []string{"\n"}, result.Stop)
				assert.Equal(t, float32(0.1), result.PresencePenalty)
				assert.Equal(t, float32(0.2), result.FrequencyPenalty)
				assert.Equal(t, intPtr(3), result.BestOf)
				assert.NotNil(t, result.LogitBias)
				assert.Equal(t, "test-user", result.User)
				assert.Equal(t, stringPtr(" END"), result.Suffix)
			},
		},
		{
			name: "Completion with array prompt",
			input: &providers.CompletionRequest{
				Model:  "text-davinci-003",
				Prompt: []string{"Hello", "World"},
			},
			expected: func(t *testing.T, result openai.CompletionRequest) {
				assert.Equal(t, []string{"Hello", "World"}, result.Prompt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformCompletionRequest(tt.input)
			tt.expected(t, result)
		})
	}
}

func TestOpenAIProvider_transformEmbeddingRequest(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    *providers.EmbeddingRequest
		expected func(*testing.T, openai.EmbeddingRequest)
	}{
		{
			name: "Basic embedding request",
			input: &providers.EmbeddingRequest{
				Model: "text-embedding-ada-002",
				Input: "Hello, world!",
			},
			expected: func(t *testing.T, result openai.EmbeddingRequest) {
				assert.Equal(t, openai.EmbeddingModel("text-embedding-ada-002"), result.Model)
				assert.Equal(t, "Hello, world!", result.Input)
			},
		},
		{
			name: "Embedding with all parameters",
			input: &providers.EmbeddingRequest{
				Model:          "text-embedding-ada-002",
				Input:          []string{"Hello", "World"},
				EncodingFormat: stringPtr("float"),
				Dimensions:     intPtr(1536),
				User:           stringPtr("test-user"),
			},
			expected: func(t *testing.T, result openai.EmbeddingRequest) {
				assert.Equal(t, openai.EmbeddingModel("text-embedding-ada-002"), result.Model)
				assert.Equal(t, []string{"Hello", "World"}, result.Input)
				assert.Equal(t, openai.EmbeddingEncodingFormat("float"), result.EncodingFormat)
				assert.Equal(t, 1536, result.Dimensions)
				assert.Equal(t, "test-user", result.User)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformEmbeddingRequest(tt.input)
			tt.expected(t, result)
		})
	}
}

func TestOpenAIProvider_transformChatCompletionResponse(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    *openai.ChatCompletionResponse
		expected *providers.ChatCompletionResponse
	}{
		{
			name: "Basic chat completion response",
			input: &openai.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1677610602,
				Model:   "gpt-3.5-turbo-0301",
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "Hello! How can I help you?",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.Usage{
					PromptTokens:     13,
					CompletionTokens: 7,
					TotalTokens:      20,
				},
				SystemFingerprint: "fp_44709d6fcb",
			},
			expected: &providers.ChatCompletionResponse{
				ID:                "chatcmpl-123",
				Object:            "chat.completion",
				Created:           1677610602,
				Model:             "gpt-3.5-turbo-0301",
				SystemFingerprint: "fp_44709d6fcb",
				Choices: []providers.ChatCompletionChoice{
					{
						Index: 0,
						Message: &providers.ChatMessage{
							Role:    "assistant",
							Content: "Hello! How can I help you?",
						},
						FinishReason: "stop",
					},
				},
				Usage: &providers.TokenUsage{
					PromptTokens:     13,
					CompletionTokens: 7,
					TotalTokens:      20,
				},
			},
		},
		{
			name: "Response with function call",
			input: &openai.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1677610602,
				Model:   "gpt-3.5-turbo-0613",
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "",
							FunctionCall: &openai.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "Boston"}`,
							},
						},
						FinishReason: "function_call",
					},
				},
				Usage: openai.Usage{
					PromptTokens:     50,
					CompletionTokens: 20,
					TotalTokens:      70,
				},
			},
			expected: &providers.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1677610602,
				Model:   "gpt-3.5-turbo-0613",
				Choices: []providers.ChatCompletionChoice{
					{
						Index: 0,
						Message: &providers.ChatMessage{
							Role:    "assistant",
							Content: "",
							FunctionCall: map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location": "Boston"}`,
							},
						},
						FinishReason: "function_call",
					},
				},
				Usage: &providers.TokenUsage{
					PromptTokens:     50,
					CompletionTokens: 20,
					TotalTokens:      70,
				},
			},
		},
		{
			name: "Response with tool calls",
			input: &openai.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1677610602,
				Model:   "gpt-4-1106-preview",
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "",
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: openai.FunctionCall{
										Name:      "get_weather",
										Arguments: `{"location": "New York"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
				Usage: openai.Usage{
					PromptTokens:     30,
					CompletionTokens: 15,
					TotalTokens:      45,
				},
			},
			expected: &providers.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion",
				Created: 1677610602,
				Model:   "gpt-4-1106-preview",
				Choices: []providers.ChatCompletionChoice{
					{
						Index: 0,
						Message: &providers.ChatMessage{
							Role:    "assistant",
							Content: "",
							ToolCalls: []providers.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: providers.FunctionCall{
										Name:      "get_weather",
										Arguments: `{"location": "New York"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
				Usage: &providers.TokenUsage{
					PromptTokens:     30,
					CompletionTokens: 15,
					TotalTokens:      45,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.transformChatCompletionResponse(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Object, result.Object)
			assert.Equal(t, tt.expected.Created, result.Created)
			assert.Equal(t, tt.expected.Model, result.Model)
			assert.Equal(t, tt.expected.SystemFingerprint, result.SystemFingerprint)

			// Compare choices
			assert.Len(t, result.Choices, len(tt.expected.Choices))
			for i, expectedChoice := range tt.expected.Choices {
				assert.Equal(t, expectedChoice.Index, result.Choices[i].Index)
				assert.Equal(t, expectedChoice.FinishReason, result.Choices[i].FinishReason)

				if expectedChoice.Message != nil {
					require.NotNil(t, result.Choices[i].Message)
					assert.Equal(t, expectedChoice.Message.Role, result.Choices[i].Message.Role)
					assert.Equal(t, expectedChoice.Message.Content, result.Choices[i].Message.Content)
					assert.Equal(t, expectedChoice.Message.FunctionCall, result.Choices[i].Message.FunctionCall)
					assert.Equal(t, expectedChoice.Message.ToolCalls, result.Choices[i].Message.ToolCalls)
				}
			}

			// Compare usage
			if tt.expected.Usage != nil {
				require.NotNil(t, result.Usage)
				assert.Equal(t, tt.expected.Usage.PromptTokens, result.Usage.PromptTokens)
				assert.Equal(t, tt.expected.Usage.CompletionTokens, result.Usage.CompletionTokens)
				assert.Equal(t, tt.expected.Usage.TotalTokens, result.Usage.TotalTokens)
			}
		})
	}
}

func TestOpenAIProvider_transformChatCompletionStreamResponse(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	input := &openai.ChatCompletionStreamResponse{
		ID:      "chatcmpl-stream-123",
		Object:  "chat.completion.chunk",
		Created: 1677610602,
		Model:   "gpt-3.5-turbo-0301",
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    "assistant",
					Content: "Hello",
				},
				FinishReason: "stop",
			},
		},
		SystemFingerprint: "fp_44709d6fcb",
	}

	result := openaiProvider.transformChatCompletionStreamResponse(input)

	assert.Equal(t, "chatcmpl-stream-123", result.ID)
	assert.Equal(t, "chat.completion.chunk", result.Object)
	assert.Equal(t, int64(1677610602), result.Created)
	assert.Equal(t, "gpt-3.5-turbo-0301", result.Model)
	assert.Equal(t, "fp_44709d6fcb", result.SystemFingerprint)

	assert.Len(t, result.Choices, 1)
	assert.Equal(t, 0, result.Choices[0].Index)
	assert.Equal(t, "stop", result.Choices[0].FinishReason)
	
	require.NotNil(t, result.Choices[0].Delta)
	assert.Equal(t, "assistant", result.Choices[0].Delta.Role)
	assert.Equal(t, "Hello", result.Choices[0].Delta.Content)
}

func TestOpenAIProvider_transformCompletionResponse(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	input := &openai.CompletionResponse{
		ID:      "cmpl-123",
		Object:  "text_completion",
		Created: 1677610602,
		Model:   "text-davinci-003",
		Choices: []openai.CompletionChoice{
			{
				Text:         " World!",
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     1,
			CompletionTokens: 2,
			TotalTokens:      3,
		},
	}

	expected := &providers.CompletionResponse{
		ID:      "cmpl-123",
		Object:  "text_completion",
		Created: 1677610602,
		Model:   "text-davinci-003",
		Choices: []providers.CompletionChoice{
			{
				Text:         " World!",
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: &providers.TokenUsage{
			PromptTokens:     1,
			CompletionTokens: 2,
			TotalTokens:      3,
		},
	}

	result := openaiProvider.transformCompletionResponse(input)

	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Object, result.Object)
	assert.Equal(t, expected.Created, result.Created)
	assert.Equal(t, expected.Model, result.Model)
	
	assert.Len(t, result.Choices, 1)
	assert.Equal(t, expected.Choices[0].Text, result.Choices[0].Text)
	assert.Equal(t, expected.Choices[0].Index, result.Choices[0].Index)
	assert.Equal(t, expected.Choices[0].FinishReason, result.Choices[0].FinishReason)
	
	require.NotNil(t, result.Usage)
	assert.Equal(t, expected.Usage.PromptTokens, result.Usage.PromptTokens)
	assert.Equal(t, expected.Usage.CompletionTokens, result.Usage.CompletionTokens)
	assert.Equal(t, expected.Usage.TotalTokens, result.Usage.TotalTokens)
}

func TestOpenAIProvider_transformEmbeddingResponse(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	input := &openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
			},
			{
				Object:    "embedding",
				Index:     1,
				Embedding: []float32{0.6, 0.7, 0.8, 0.9, 1.0},
			},
		},
		Model: "text-embedding-ada-002",
		Usage: openai.Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	expected := &providers.EmbeddingResponse{
		Object: "list",
		Data: []providers.Embedding{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			},
			{
				Object:    "embedding",
				Index:     1,
				Embedding: []float64{0.6, 0.7, 0.8, 0.9, 1.0},
			},
		},
		Model: "text-embedding-ada-002",
		Usage: &providers.TokenUsage{
			PromptTokens:     5,
			CompletionTokens: 0,
			TotalTokens:      5,
		},
	}

	result := openaiProvider.transformEmbeddingResponse(input)

	assert.Equal(t, expected.Object, result.Object)
	assert.Equal(t, expected.Model, result.Model)
	
	assert.Len(t, result.Data, 2)
	for i, expectedData := range expected.Data {
		assert.Equal(t, expectedData.Object, result.Data[i].Object)
		assert.Equal(t, expectedData.Index, result.Data[i].Index)
		assert.Equal(t, expectedData.Embedding, result.Data[i].Embedding)
	}
	
	require.NotNil(t, result.Usage)
	assert.Equal(t, expected.Usage.PromptTokens, result.Usage.PromptTokens)
	assert.Equal(t, expected.Usage.CompletionTokens, result.Usage.CompletionTokens)
	assert.Equal(t, expected.Usage.TotalTokens, result.Usage.TotalTokens)
}

func TestOpenAIProvider_transformModel(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	input := &openai.Model{
		ID:      "gpt-3.5-turbo",
		Object:  "model",
		Created: 1677610602,
		OwnedBy: "openai",
		Root:    "gpt-3.5-turbo",
		Parent:  "",
		Permission: []openai.Permission{
			{
				ID:                 "modelperm-123",
				Object:             "model_permission",
				Created:            1677610602,
				AllowCreateEngine:  false,
				AllowSampling:      true,
				AllowLogprobs:      true,
				AllowSearchIndices: false,
				AllowView:          true,
				AllowFineTuning:    false,
				Organization:       "*",
				Group:              nil,
				IsBlocking:         false,
			},
		},
	}

	expected := &providers.Model{
		ID:      "gpt-3.5-turbo",
		Object:  "model",
		Created: 1677610602,
		OwnedBy: "openai",
		Root:    "gpt-3.5-turbo",
		Parent:  "",
		Permission: []providers.ModelPermission{
			{
				ID:                 "modelperm-123",
				Object:             "model_permission",
				Created:            1677610602,
				AllowCreateEngine:  false,
				AllowSampling:      true,
				AllowLogprobs:      true,
				AllowSearchIndices: false,
				AllowView:          true,
				AllowFineTuning:    false,
				Organization:       "*",
				Group:              nil,
				IsBlocking:         false,
			},
		},
	}

	result := openaiProvider.transformModel(input)

	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Object, result.Object)
	assert.Equal(t, expected.Created, result.Created)
	assert.Equal(t, expected.OwnedBy, result.OwnedBy)
	assert.Equal(t, expected.Root, result.Root)
	assert.Equal(t, expected.Parent, result.Parent)
	
	assert.Len(t, result.Permission, 1)
	assert.Equal(t, expected.Permission[0], result.Permission[0])
}

func TestOpenAIProvider_marshalJSON(t *testing.T) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(t, err)

	openaiProvider := provider.(*OpenAIProvider)

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Valid object",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "String",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "Number",
			input:    42,
			expected: `42`,
		},
		{
			name:     "Invalid object (channel)",
			input:    make(chan int),
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openaiProvider.marshalJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for transformations
func BenchmarkTransformChatCompletionRequest(b *testing.B) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(b, err)

	openaiProvider := provider.(*OpenAIProvider)

	req := &providers.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []providers.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   intPtr(100),
		Temperature: float64Ptr(0.7),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = openaiProvider.transformChatCompletionRequest(req)
	}
}

func BenchmarkTransformChatCompletionResponse(b *testing.B) {
	provider, err := NewOpenAIProvider(createValidConfig())
	require.NoError(b, err)

	openaiProvider := provider.(*OpenAIProvider)

	resp := &openai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1677610602,
		Model:   "gpt-3.5-turbo",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     13,
			CompletionTokens: 7,
			TotalTokens:      20,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = openaiProvider.transformChatCompletionResponse(resp)
	}
}