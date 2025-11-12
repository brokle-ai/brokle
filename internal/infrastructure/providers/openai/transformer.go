package openai

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"

	"brokle/internal/infrastructure/providers"
)

// Request transformation methods

func (p *OpenAIProvider) transformChatCompletionRequest(req *providers.ChatCompletionRequest) openai.ChatCompletionRequest {
	openaiReq := openai.ChatCompletionRequest{
		Model:            req.Model,
		MaxTokens:        p.derefInt(req.MaxTokens),
		Temperature:      p.derefFloat32(req.Temperature),
		TopP:             p.derefFloat32(req.TopP),
		N:                p.derefInt(req.N),
		Stream:           req.Stream,
		Stop:             req.Stop,
		PresencePenalty:  p.derefFloat32(req.PresencePenalty),
		FrequencyPenalty: p.derefFloat32(req.FrequencyPenalty),
		LogitBias:        p.transformLogitBias(req.LogitBias),
		User:             p.derefString(req.User),
		Seed:             req.Seed,
	}

	// Transform messages
	openaiReq.Messages = make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		openaiReq.Messages[i] = p.transformChatMessage(&msg)
	}

	// Transform functions (legacy)
	if len(req.Functions) > 0 {
		openaiReq.Functions = make([]openai.FunctionDefinition, len(req.Functions))
		for i, fn := range req.Functions {
			openaiReq.Functions[i] = openai.FunctionDefinition{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
			}
		}
		openaiReq.FunctionCall = req.FunctionCall
	}

	// Transform tools (new format)
	if len(req.Tools) > 0 {
		openaiReq.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			openaiReq.Tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		openaiReq.ToolChoice = req.ToolChoice
	}

	// Transform response format
	if req.ResponseFormat != nil {
		openaiReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatType(req.ResponseFormat.Type),
		}
	}

	return openaiReq
}

func (p *OpenAIProvider) transformChatMessage(msg *providers.ChatMessage) openai.ChatCompletionMessage {
	openaiMsg := openai.ChatCompletionMessage{
		Role:    msg.Role,
		Content: p.transformMessageContent(msg.Content),
		Name:    p.derefString(msg.Name),
	}

	// Handle function call (legacy)
	if msg.FunctionCall != nil {
		if fcMap, ok := msg.FunctionCall.(map[string]interface{}); ok {
			if name, exists := fcMap["name"].(string); exists {
				if args, exists := fcMap["arguments"].(string); exists {
					openaiMsg.FunctionCall = &openai.FunctionCall{
						Name:      name,
						Arguments: args,
					}
				}
			}
		}
	}

	// Handle tool calls (new format)
	if len(msg.ToolCalls) > 0 {
		openaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			openaiMsg.ToolCalls[i] = openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolType(tc.Type),
				Function: openai.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	// Handle tool call ID for tool response messages
	if msg.ToolCallID != nil {
		openaiMsg.ToolCallID = p.derefString(msg.ToolCallID)
	}

	return openaiMsg
}

func (p *OpenAIProvider) transformMessageContent(content interface{}) string {
	if content == nil {
		return ""
	}

	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// Handle multimodal content (text + images)
		// For now, extract text content; full multimodal support would require more complex handling
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if contentType, exists := itemMap["type"].(string); exists && contentType == "text" {
					if text, exists := itemMap["text"].(string); exists {
						return text
					}
				}
			}
		}
		return ""
	default:
		// Try to serialize as JSON as fallback
		if data, err := json.Marshal(content); err == nil {
			return string(data)
		}
		return ""
	}
}

func (p *OpenAIProvider) transformCompletionRequest(req *providers.CompletionRequest) openai.CompletionRequest {
	return openai.CompletionRequest{
		Model:            req.Model,
		Prompt:           p.transformPrompt(req.Prompt),
		MaxTokens:        p.derefInt(req.MaxTokens),
		Temperature:      p.derefFloat32(req.Temperature),
		TopP:             p.derefFloat32(req.TopP),
		N:                p.derefInt(req.N),
		Stream:           req.Stream,
		LogProbs:         p.derefInt(req.Logprobs), // Fixed to dereference pointer
		Echo:             req.Echo,
		Stop:             req.Stop,
		PresencePenalty:  p.derefFloat32(req.PresencePenalty),
		FrequencyPenalty: p.derefFloat32(req.FrequencyPenalty),
		BestOf:           p.derefInt(req.BestOf),
		LogitBias:        p.transformLogitBias(req.LogitBias),
		User:             p.derefString(req.User),
		Suffix:           p.derefString(req.Suffix),
	}
}

func (p *OpenAIProvider) transformPrompt(prompt interface{}) interface{} {
	// OpenAI supports string or array of strings for prompt
	return prompt
}

func (p *OpenAIProvider) transformEmbeddingRequest(req *providers.EmbeddingRequest) openai.EmbeddingRequest {
	openaiReq := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(req.Model),
		Input: req.Input,
		User:  p.derefString(req.User),
	}

	if req.EncodingFormat != nil {
		openaiReq.EncodingFormat = openai.EmbeddingEncodingFormat(*req.EncodingFormat)
	}

	if req.Dimensions != nil {
		openaiReq.Dimensions = *req.Dimensions
	}

	return openaiReq
}

// Response transformation methods

func (p *OpenAIProvider) transformChatCompletionResponse(resp *openai.ChatCompletionResponse) *providers.ChatCompletionResponse {
	result := &providers.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
	}

	// Only set SystemFingerprint if non-empty
	if resp.SystemFingerprint != "" {
		result.SystemFingerprint = &resp.SystemFingerprint
	}

	// Transform choices
	if len(resp.Choices) > 0 {
		result.Choices = make([]providers.ChatCompletionChoice, len(resp.Choices))
		for i, choice := range resp.Choices {
			result.Choices[i] = providers.ChatCompletionChoice{
				Index: choice.Index,
			}

			// Only set FinishReason if non-empty
			if choice.FinishReason != "" {
				finishReason := string(choice.FinishReason)
				result.Choices[i].FinishReason = &finishReason
			}

			if choice.Message.Role != "" {
				result.Choices[i].Message = &providers.ChatMessage{
					Role:    choice.Message.Role,
					Content: choice.Message.Content,
				}

				// Only set Name if non-empty
				if choice.Message.Name != "" {
					result.Choices[i].Message.Name = &choice.Message.Name
				}

				// Transform function call
				if choice.Message.FunctionCall != nil {
					result.Choices[i].Message.FunctionCall = map[string]interface{}{
						"name":      choice.Message.FunctionCall.Name,
						"arguments": choice.Message.FunctionCall.Arguments,
					}
				}

				// Transform tool calls
				if len(choice.Message.ToolCalls) > 0 {
					result.Choices[i].Message.ToolCalls = make([]providers.ToolCall, len(choice.Message.ToolCalls))
					for j, tc := range choice.Message.ToolCalls {
						result.Choices[i].Message.ToolCalls[j] = providers.ToolCall{
							ID:   tc.ID,
							Type: string(tc.Type),
							Function: providers.FunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						}
					}
				}
			}
		}
	}

	// Transform usage
	if resp.Usage.TotalTokens > 0 {
		result.Usage = &providers.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result
}

func (p *OpenAIProvider) transformChatCompletionStreamResponse(resp *openai.ChatCompletionStreamResponse) *providers.ChatCompletionResponse {
	result := &providers.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
	}

	// Only set SystemFingerprint if non-empty
	if resp.SystemFingerprint != "" {
		result.SystemFingerprint = &resp.SystemFingerprint
	}

	// Transform choices for streaming
	if len(resp.Choices) > 0 {
		result.Choices = make([]providers.ChatCompletionChoice, len(resp.Choices))
		for i, choice := range resp.Choices {
			result.Choices[i] = providers.ChatCompletionChoice{
				Index: choice.Index,
			}

			// Only set FinishReason if non-empty
			if choice.FinishReason != "" {
				finishReason := string(choice.FinishReason)
				result.Choices[i].FinishReason = &finishReason
			}

			// For streaming, we have delta instead of message
			if choice.Delta.Role != "" || choice.Delta.Content != "" {
				result.Choices[i].Delta = &providers.ChatMessage{
					Role:    choice.Delta.Role,
					Content: choice.Delta.Content,
				}

				// Transform function call delta
				if choice.Delta.FunctionCall != nil {
					result.Choices[i].Delta.FunctionCall = map[string]interface{}{
						"name":      choice.Delta.FunctionCall.Name,
						"arguments": choice.Delta.FunctionCall.Arguments,
					}
				}

				// Transform tool calls delta
				if len(choice.Delta.ToolCalls) > 0 {
					result.Choices[i].Delta.ToolCalls = make([]providers.ToolCall, len(choice.Delta.ToolCalls))
					for j, tc := range choice.Delta.ToolCalls {
						result.Choices[i].Delta.ToolCalls[j] = providers.ToolCall{
							ID:   tc.ID,
							Type: string(tc.Type),
						}

						if tc.Function.Name != "" || tc.Function.Arguments != "" {
							result.Choices[i].Delta.ToolCalls[j].Function = providers.FunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							}
						}
					}
				}
			}
		}
	}

	return result
}

func (p *OpenAIProvider) transformCompletionResponse(resp *openai.CompletionResponse) *providers.CompletionResponse {
	result := &providers.CompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
	}

	// Transform choices
	if len(resp.Choices) > 0 {
		result.Choices = make([]providers.CompletionChoice, len(resp.Choices))
		for i, choice := range resp.Choices {
			result.Choices[i] = providers.CompletionChoice{
				Text:     choice.Text,
				Index:    choice.Index,
				Logprobs: choice.LogProbs,
			}

			// Only set FinishReason if non-empty
			if choice.FinishReason != "" {
				finishReason := choice.FinishReason
				result.Choices[i].FinishReason = &finishReason
			}
		}
	}

	// Transform usage
	if resp.Usage.TotalTokens > 0 {
		result.Usage = &providers.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result
}

func (p *OpenAIProvider) transformCompletionStreamResponse(resp *openai.CompletionResponse) *providers.CompletionResponse {
	// Same transformation as regular completion for now
	return p.transformCompletionResponse(resp)
}

func (p *OpenAIProvider) transformEmbeddingResponse(resp *openai.EmbeddingResponse) *providers.EmbeddingResponse {
	result := &providers.EmbeddingResponse{
		Object: resp.Object,
		Model:  string(resp.Model),
	}

	// Transform embeddings
	if len(resp.Data) > 0 {
		result.Data = make([]providers.Embedding, len(resp.Data))
		for i, embedding := range resp.Data {
			// Convert []float32 to []float64
			embeddingFloat64 := make([]float64, len(embedding.Embedding))
			for j, val := range embedding.Embedding {
				embeddingFloat64[j] = float64(val)
			}
			result.Data[i] = providers.Embedding{
				Object:    embedding.Object,
				Index:     embedding.Index,
				Embedding: embeddingFloat64,
			}
		}
	}

	// Transform usage
	if resp.Usage.TotalTokens > 0 {
		result.Usage = &providers.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: 0, // Embeddings don't have completion tokens
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result
}

func (p *OpenAIProvider) transformModel(model *openai.Model) *providers.Model {
	result := &providers.Model{
		ID:     model.ID,
		Object: model.Object,
		// Created field doesn't exist in openai.Model, use 0 as default
		Created: 0,
		OwnedBy: model.OwnedBy,
		Root:    model.Root,
		Parent:  &model.Parent,
	}

	// Transform permissions if present
	if len(model.Permission) > 0 {
		result.Permission = make([]providers.ModelPermission, len(model.Permission))
		for i, perm := range model.Permission {
			// Convert Group interface{} to *string safely
			var groupStr *string
			if groupVal, ok := perm.Group.(string); ok && groupVal != "" {
				groupStr = &groupVal
			}
			result.Permission[i] = providers.ModelPermission{
				ID:     perm.ID,
				Object: perm.Object,
				// Created field doesn't exist in openai.Permission, use 0 as default
				Created:            0,
				AllowCreateEngine:  perm.AllowCreateEngine,
				AllowSampling:      perm.AllowSampling,
				AllowLogprobs:      perm.AllowLogprobs,
				AllowSearchIndices: perm.AllowSearchIndices,
				AllowView:          perm.AllowView,
				AllowFineTuning:    perm.AllowFineTuning,
				Organization:       perm.Organization,
				Group:              groupStr,
				IsBlocking:         perm.IsBlocking,
			}
		}
	}

	return result
}

// Helper method to marshal JSON for streaming
func (p *OpenAIProvider) marshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		p.logger.WithError(err).Error("Failed to marshal JSON for streaming")
		return "{}"
	}
	return string(data)
}

// Helper methods for dereferencing pointers safely

func (p *OpenAIProvider) derefInt(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func (p *OpenAIProvider) derefFloat32(ptr *float64) float32 {
	if ptr == nil {
		return 0
	}
	return float32(*ptr)
}

func (p *OpenAIProvider) derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (p *OpenAIProvider) transformLogitBias(logitBias map[string]interface{}) map[string]int {
	if logitBias == nil {
		return nil
	}
	result := make(map[string]int)
	for k, v := range logitBias {
		if intVal, ok := v.(int); ok {
			result[k] = intVal
		} else if floatVal, ok := v.(float64); ok {
			result[k] = int(floatVal)
		}
	}
	return result
}
