package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/gateway"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler handles AI Gateway endpoints (OpenAI-compatible)
type Handler struct {
	config         *config.Config
	logger         *logrus.Logger
	gatewayService gateway.GatewayService
	routingService gateway.RoutingService
	costService    gateway.CostService
}

// NewHandler creates a new AI handler
func NewHandler(
	config *config.Config,
	logger *logrus.Logger,
	gatewayService gateway.GatewayService,
	routingService gateway.RoutingService,
	costService gateway.CostService,
) *Handler {
	return &Handler{
		config:         config,
		logger:         logger,
		gatewayService: gatewayService,
		routingService: routingService,
		costService:    costService,
	}
}

// ChatCompletionRequest represents the chat completion request
// @Description OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model            string                 `json:"model" example:"gpt-3.5-turbo" description:"Model to use for completion"`
	Messages         []ChatMessage          `json:"messages" description:"Array of messages in the conversation"`
	MaxTokens        *int                   `json:"max_tokens,omitempty" example:"150" description:"Maximum number of tokens to generate"`
	Temperature      *float64               `json:"temperature,omitempty" example:"0.7" description:"Controls randomness (0.0 to 2.0)"`
	TopP             *float64               `json:"top_p,omitempty" example:"1.0" description:"Nucleus sampling parameter"`
	N                *int                   `json:"n,omitempty" example:"1" description:"Number of completions to generate"`
	Stream           *bool                  `json:"stream,omitempty" example:"false" description:"Whether to stream responses"`
	Stop             interface{}            `json:"stop,omitempty" description:"Up to 4 sequences where the API will stop generating"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty" example:"0.0" description:"Presence penalty (-2.0 to 2.0)"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty" example:"0.0" description:"Frequency penalty (-2.0 to 2.0)"`
	LogitBias        map[string]float64     `json:"logit_bias,omitempty" description:"Modify likelihood of specified tokens"`
	User             *string                `json:"user,omitempty" example:"user-123" description:"Unique identifier for the end-user"`
}

// ChatMessage represents a chat message
// @Description Single message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role" example:"user" description:"Message role (system, user, assistant)"`
	Content string `json:"content" example:"Hello, how are you?" description:"Message content"`
	Name    string `json:"name,omitempty" example:"John" description:"Optional name of the message author"`
}

// CompletionRequest represents the text completion request
// @Description OpenAI-compatible text completion request
type CompletionRequest struct {
	Model            string             `json:"model" example:"gpt-3.5-turbo-instruct" description:"Model to use for completion"`
	Prompt           interface{}        `json:"prompt" description:"Prompt text or array of prompts" swaggertype:"string" example:"Once upon a time"`
	MaxTokens        *int               `json:"max_tokens,omitempty" example:"150" description:"Maximum number of tokens to generate"`
	Temperature      *float64           `json:"temperature,omitempty" example:"0.7" description:"Controls randomness (0.0 to 2.0)"`
	TopP             *float64           `json:"top_p,omitempty" example:"1.0" description:"Nucleus sampling parameter"`
	N                *int               `json:"n,omitempty" example:"1" description:"Number of completions to generate"`
	Stream           *bool              `json:"stream,omitempty" example:"false" description:"Whether to stream responses"`
	Logprobs         *int               `json:"logprobs,omitempty" example:"0" description:"Include log probabilities on logprobs tokens"`
	Echo             *bool              `json:"echo,omitempty" example:"false" description:"Echo back the prompt in addition to completion"`
	Stop             interface{}        `json:"stop,omitempty" description:"Up to 4 sequences where the API will stop generating"`
	PresencePenalty  *float64           `json:"presence_penalty,omitempty" example:"0.0" description:"Presence penalty (-2.0 to 2.0)"`
	FrequencyPenalty *float64           `json:"frequency_penalty,omitempty" example:"0.0" description:"Frequency penalty (-2.0 to 2.0)"`
	BestOf           *int               `json:"best_of,omitempty" example:"1" description:"Generate best_of completions server-side"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty" description:"Modify likelihood of specified tokens"`
	User             *string            `json:"user,omitempty" example:"user-123" description:"Unique identifier for the end-user"`
}

// EmbeddingRequest represents the embedding request
// @Description OpenAI-compatible embedding request
type EmbeddingRequest struct {
	Model          string      `json:"model" example:"text-embedding-ada-002" description:"Model to use for embeddings"`
	Input          interface{} `json:"input" description:"Input text or array of texts" swaggertype:"string" example:"The food was delicious and the waiter..."`
	User           *string     `json:"user,omitempty" example:"user-123" description:"Unique identifier for the end-user"`
	EncodingFormat *string     `json:"encoding_format,omitempty" example:"float" description:"Format of returned embeddings (float or base64)"`
}

// Model represents an AI model
// @Description Available AI model information
type Model struct {
	ID      string `json:"id" example:"gpt-3.5-turbo" description:"Model identifier"`
	Object  string `json:"object" example:"model" description:"Object type (always 'model')"`
	Created int64  `json:"created" example:"1677610602" description:"Unix timestamp when model was created"`
	OwnedBy string `json:"owned_by" example:"openai" description:"Organization that owns the model"`
}

// RouteRequest represents AI routing request
// @Description Request data for AI routing decisions
type RouteRequest struct {
	Model       string                 `json:"model" example:"gpt-3.5-turbo" description:"Target AI model"`
	Provider    *string                `json:"provider,omitempty" example:"openai" description:"Preferred AI provider"`
	Messages    []ChatMessage          `json:"messages,omitempty" description:"Chat messages for context"`
	Prompt      *string                `json:"prompt,omitempty" example:"Hello world" description:"Text prompt for context"`
	MaxTokens   *int                   `json:"max_tokens,omitempty" example:"150" description:"Maximum tokens for estimation"`
	Temperature *float64               `json:"temperature,omitempty" example:"0.7" description:"Temperature parameter"`
	Strategy    *string                `json:"strategy,omitempty" example:"cost_optimized" description:"Routing strategy (cost_optimized, latency_optimized, quality_optimized)"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Additional metadata for routing decisions"`
}

// RouteResponse represents AI routing response
// @Description AI routing decision response
type RouteResponse struct {
	Provider         string                 `json:"provider" example:"openai" description:"Selected AI provider"`
	Model            string                 `json:"model" example:"gpt-3.5-turbo" description:"Selected model"`
	Endpoint         string                 `json:"endpoint" example:"https://api.openai.com/v1/chat/completions" description:"Provider endpoint URL"`
	Strategy         string                 `json:"strategy" example:"cost_optimized" description:"Applied routing strategy"`
	EstimatedCost    *float64               `json:"estimated_cost,omitempty" example:"0.0015" description:"Estimated cost in USD"`
	EstimatedLatency *int                   `json:"estimated_latency,omitempty" example:"250" description:"Estimated latency in milliseconds"`
	QualityScore     *float64               `json:"quality_score,omitempty" example:"0.95" description:"Expected quality score (0.0-1.0)"`
	CacheHit         bool                   `json:"cache_hit" example:"false" description:"Whether response can be served from cache"`
	ProviderHealth   *float64               `json:"provider_health,omitempty" example:"0.98" description:"Provider health score (0.0-1.0)"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" description:"Additional routing metadata"`
}

// CacheStatusResponse represents cache status information
// @Description Cache health and statistics response
type CacheStatusResponse struct {
	Status          string  `json:"status" example:"healthy" description:"Cache health status"`
	HitRate         float64 `json:"hit_rate" example:"0.85" description:"Cache hit rate (0.0-1.0)"`
	TotalEntries    int64   `json:"total_entries" example:"15420" description:"Total number of cached entries"`
	SizeBytes       int64   `json:"size_bytes" example:"1048576" description:"Total cache size in bytes"`
	MemoryUsage     float64 `json:"memory_usage" example:"0.45" description:"Memory usage percentage (0.0-1.0)"`
	EvictionCount   int64   `json:"eviction_count" example:"142" description:"Number of evicted entries"`
	LastEviction    *int64  `json:"last_eviction,omitempty" example:"1677610602" description:"Unix timestamp of last eviction"`
	ProviderBreakdown map[string]interface{} `json:"provider_breakdown,omitempty" description:"Cache statistics by provider"`
}

// InvalidateCacheRequest represents cache invalidation request
// @Description Request data for cache invalidation
type InvalidateCacheRequest struct {
	Provider    *string  `json:"provider,omitempty" example:"openai" description:"Target specific provider"`
	Model       *string  `json:"model,omitempty" example:"gpt-3.5-turbo" description:"Target specific model"`
	Keys        []string `json:"keys,omitempty" description:"Specific cache keys to invalidate"`
	ClearAll    *bool    `json:"clear_all,omitempty" example:"false" description:"Clear entire cache (use with caution)"`
	MaxAge      *int     `json:"max_age,omitempty" example:"3600" description:"Invalidate entries older than this (seconds)"`
	Pattern     *string  `json:"pattern,omitempty" example:"chat:*" description:"Pattern for key matching"`
}

// InvalidateCacheResponse represents cache invalidation response
// @Description Cache invalidation result
type InvalidateCacheResponse struct {
	Success         bool     `json:"success" example:"true" description:"Whether invalidation succeeded"`
	InvalidatedKeys []string `json:"invalidated_keys,omitempty" description:"List of invalidated cache keys"`
	Count           int      `json:"count" example:"25" description:"Number of entries invalidated"`
	Message         string   `json:"message" example:"Cache invalidated successfully" description:"Operation result message"`
	Error           *string  `json:"error,omitempty" description:"Error message if operation failed"`
}

// ChatCompletions handles OpenAI-compatible chat completions
// @Summary Create chat completion
// @Description Generate AI chat completions using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChatCompletionRequest true "Chat completion request"
// @Success 200 {object} response.SuccessResponse "Chat completion generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/chat/completions [post]
func (h *Handler) ChatCompletions(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/chat/completions",
		"method":     "POST",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing chat completion request")

	// Parse request body
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to parse chat completion request")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Invalid request payload: "+err.Error())
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
		"model":           req.Model,
	})

	// Validate request
	if err := h.validateChatCompletionRequest(&req); err != nil {
		logger.WithError(err).Error("Chat completion request validation failed")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Transform to gateway request
	gatewayReq := h.transformChatCompletionRequest(&req, orgID, environment)

	// Process through gateway
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	gatewayResp, err := h.gatewayService.CreateChatCompletion(ctx, orgID, string(environment), gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Gateway processing failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to OpenAI-compatible response
	resp := h.transformChatCompletionResponse(gatewayResp)

	logger.WithFields(logrus.Fields{
		"response_tokens": gatewayResp.Usage.TotalTokens,
		"choices_count":   len(gatewayResp.Choices),
	}).Info("Chat completion request completed successfully")

	// Return response
	if req.Stream != nil && *req.Stream {
		// Handle streaming response with SSE
		h.handleChatCompletionStream(c, orgID, string(environment), gatewayReq)
		return
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

// Completions handles OpenAI-compatible text completions
// @Summary Create text completion
// @Description Generate AI text completions using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CompletionRequest true "Text completion request"
// @Success 200 {object} response.SuccessResponse "Text completion generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/completions [post]
func (h *Handler) Completions(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/completions",
		"method":     "POST",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing completion request")

	// Parse request body
	var req CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to parse completion request")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Invalid request payload: "+err.Error())
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
		"model":           req.Model,
	})

	// Validate request
	if err := h.validateCompletionRequest(&req); err != nil {
		logger.WithError(err).Error("Completion request validation failed")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Transform to gateway request
	gatewayReq := h.transformCompletionRequest(&req, orgID, environment)

	// Process through gateway
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	gatewayResp, err := h.gatewayService.CreateCompletion(ctx, orgID, string(environment), gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Gateway processing failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to OpenAI-compatible response
	resp := h.transformCompletionResponse(gatewayResp)

	logger.WithFields(logrus.Fields{
		"response_tokens": gatewayResp.Usage.TotalTokens,
		"choices_count":   len(gatewayResp.Choices),
	}).Info("Completion request completed successfully")

	// Return response
	if req.Stream != nil && *req.Stream {
		// Handle streaming response with SSE
		h.handleCompletionStream(c, orgID, string(environment), gatewayReq)
		return
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

// Embeddings handles OpenAI-compatible embeddings
// @Summary Create embeddings
// @Description Generate text embeddings using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body EmbeddingRequest true "Embedding request"
// @Success 200 {object} response.SuccessResponse "Embeddings generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/embeddings [post]
func (h *Handler) Embeddings(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/embeddings",
		"method":     "POST",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing embeddings request")

	// Parse request body
	var req EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to parse embeddings request")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Invalid request payload: "+err.Error())
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
		"model":           req.Model,
	})

	// Validate request
	if err := h.validateEmbeddingsRequest(&req); err != nil {
		logger.WithError(err).Error("Embeddings request validation failed")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Transform to gateway request
	gatewayReq := h.transformEmbeddingsRequest(&req, orgID, environment)

	// Process through gateway
	ctx := c.Request.Context()
	gatewayResp, err := h.gatewayService.CreateEmbedding(ctx, orgID, string(environment), gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Gateway processing failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to OpenAI-compatible response
	resp := h.transformEmbeddingsResponse(gatewayResp)

	logger.WithFields(logrus.Fields{
		"response_tokens":     gatewayResp.Usage.TotalTokens,
		"embeddings_count":    len(gatewayResp.Data),
	}).Info("Embeddings request completed successfully")

	c.JSON(http.StatusOK, resp)
}

// ListModels handles listing available AI models
// @Summary List available models
// @Description Get list of available AI models
// @Tags AI Gateway
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} Model "List of available models"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/models [get]
func (h *Handler) ListModels(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/models",
		"method":     "GET",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing list models request")

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
	})

	// Create gateway request
	gatewayReq := &gateway.ListModelsRequest{
		OrganizationID: orgID,
		Environment:    environment,
	}

	// Process through gateway
	ctx := c.Request.Context()
	models, err := h.gatewayService.ListAvailableModels(ctx, orgID)
	if err != nil {
		logger.WithError(err).Error("Gateway processing failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to OpenAI-compatible response
	respData := make([]Model, len(models))
	for i, model := range models {
		respData[i] = Model{
			ID:      model.ID,
			Object:  "model",
			Created: model.Created,
			OwnedBy: model.OwnedBy,
		}
	}
	
	resp := ModelsResponse{
		Object: "list",
		Data:   respData,
	}

	logger.WithField("model_count", len(resp.Data)).Info("List models request completed successfully")

	c.JSON(http.StatusOK, resp)
}

// GetModel handles getting specific model information
// @Summary Get model information
// @Description Get detailed information about a specific AI model
// @Tags AI Gateway
// @Produce json
// @Security ApiKeyAuth
// @Param model path string true "Model ID" example("gpt-3.5-turbo")
// @Success 200 {object} Model "Model information"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 404 {object} response.ErrorResponse "Model not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/models/{model} [get]
func (h *Handler) GetModel(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/models/{model}",
		"method":     "GET",
		"request_id": c.GetString("request_id"),
	})

	modelID := c.Param("model")
	logger = logger.WithField("model_id", modelID)

	logger.Info("Processing get model request")

	if modelID == "" {
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Model ID is required")
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)

	// Get available models and find the requested one
	gatewayReq := &gateway.ListModelsRequest{
		OrganizationID: orgID,
		Environment:    environment,
	}

	ctx := c.Request.Context()
	models, err := h.gatewayService.ListAvailableModels(ctx, orgID)
	if err != nil {
		logger.WithError(err).Error("Gateway processing failed")
		h.handleGatewayError(c, err)
		return
	}

	// Find the specific model
	var foundModel *gateway.ModelInfo
	for _, model := range models {
		if model.ID == modelID {
			foundModel = model
			break
		}
	}

	if foundModel == nil {
		response.ErrorWithCode(c, http.StatusNotFound, "model_not_found", "The model does not exist")
		return
	}

	// Transform to OpenAI-compatible response
	resp := Model{
		ID:      foundModel.ID,
		Object:  "model",
		Created: foundModel.Created,
		OwnedBy: foundModel.OwnedBy,
	}

	logger.Info("Get model request completed successfully")

	c.JSON(http.StatusOK, resp)
}

// RouteRequest handles AI routing decisions
// @Summary Make AI routing decision
// @Description Determine optimal AI provider and model for a request
// @Tags SDK - Routing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body RouteRequest true "Routing request data"
// @Success 200 {object} response.SuccessResponse{data=RouteResponse} "Routing decision returned"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid or missing API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/route [post]
func (h *Handler) RouteRequest(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/route",
		"method":     "POST",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing route request")

	// Parse request body
	var req RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to parse route request")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Invalid request payload: "+err.Error())
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
		"model":           req.Model,
	})

	// Validate request
	if err := h.validateRouteRequest(&req); err != nil {
		logger.WithError(err).Error("Route request validation failed")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Transform to gateway routing request
	gatewayReq := h.transformRouteRequest(&req, orgID, environment)

	// Process through routing service
	ctx := c.Request.Context()
	routingResp, err := h.routingService.SelectProvider(ctx, gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Routing service failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to API response
	resp := h.transformRouteResponse(routingResp)

	logger.WithFields(logrus.Fields{
		"selected_provider": resp.Provider,
		"selected_model":    resp.Model,
		"strategy":          resp.Strategy,
		"estimated_cost":    resp.EstimatedCost,
	}).Info("Route request completed successfully")

	response.Success(c, resp)
}

// CacheStatus handles cache health checks
// @Summary Get cache status
// @Description Get current cache health and statistics
// @Tags SDK - Cache
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.SuccessResponse{data=CacheStatusResponse} "Cache status returned"
// @Failure 401 {object} response.ErrorResponse "Invalid or missing API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/cache/status [get]
func (h *Handler) CacheStatus(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/cache/status",
		"method":     "GET",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing cache status request")

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
	})

	// Create cache status request
	gatewayReq := &gateway.CacheStatusRequest{
		OrganizationID: orgID,
		Environment:    environment,
	}

	// Get cache status from gateway service
	ctx := c.Request.Context()
	gatewayResp, err := h.gatewayService.GetCacheStatus(ctx, gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Gateway cache status failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to API response
	resp := h.transformCacheStatusResponse(gatewayResp)

	logger.WithFields(logrus.Fields{
		"status":          resp.Status,
		"hit_rate":        resp.HitRate,
		"total_entries":   resp.TotalEntries,
		"memory_usage":    resp.MemoryUsage,
	}).Info("Cache status request completed successfully")

	response.Success(c, resp)
}

// InvalidateCache handles cache invalidation
// @Summary Invalidate cache entries
// @Description Invalidate specific cache entries or clear cache
// @Tags SDK - Cache
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body InvalidateCacheRequest true "Cache invalidation data"
// @Success 200 {object} response.SuccessResponse{data=InvalidateCacheResponse} "Cache invalidated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid or missing API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/cache/invalidate [post]
func (h *Handler) InvalidateCache(c *gin.Context) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint":   "/v1/cache/invalidate",
		"method":     "POST",
		"request_id": c.GetString("request_id"),
	})

	logger.Info("Processing cache invalidation request")

	// Parse request body
	var req InvalidateCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to parse cache invalidation request")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", "Invalid request payload: "+err.Error())
		return
	}

	// Extract organization and environment from context
	orgID, err := h.extractOrganizationID(c)
	if err != nil {
		logger.WithError(err).Error("Failed to extract organization ID")
		response.ErrorWithCode(c, http.StatusUnauthorized, "unauthorized", "Invalid organization context")
		return
	}

	environment := h.extractEnvironment(c)
	logger = logger.WithFields(logrus.Fields{
		"organization_id": orgID,
		"environment":     environment,
		"clear_all":       req.ClearAll != nil && *req.ClearAll,
	})

	// Validate request
	if err := h.validateInvalidateCacheRequest(&req); err != nil {
		logger.WithError(err).Error("Cache invalidation request validation failed")
		response.ErrorWithCode(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Transform to gateway request
	gatewayReq := h.transformInvalidateCacheRequest(&req, orgID, environment)

	// Process cache invalidation through gateway service
	ctx := c.Request.Context()
	gatewayResp, err := h.gatewayService.InvalidateCache(ctx, gatewayReq)
	if err != nil {
		logger.WithError(err).Error("Gateway cache invalidation failed")
		h.handleGatewayError(c, err)
		return
	}

	// Transform to API response
	resp := h.transformInvalidateCacheResponse(gatewayResp)

	logger.WithFields(logrus.Fields{
		"success":           resp.Success,
		"invalidated_count": resp.Count,
	}).Info("Cache invalidation request completed successfully")

	response.Success(c, resp)
}

// OpenAI-compatible response structures

type ChatCompletionResponse struct {
	ID      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Model   string                     `json:"model"`
	Choices []ChatCompletionChoice     `json:"choices"`
	Usage   ChatCompletionUsage        `json:"usage"`
	
	// Brokle extensions
	Provider        *string                `json:"provider,omitempty"`
	Cost            *gateway.CostCalculation `json:"cost,omitempty"`
	RoutingDecision *string                `json:"routing_decision,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason *string      `json:"finish_reason"`
	LogProbs     interface{}  `json:"logprobs"`
}

type ChatCompletionUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

type CompletionResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []CompletionChoice  `json:"choices"`
	Usage   ChatCompletionUsage `json:"usage"`
	
	// Brokle extensions
	Provider        *string                `json:"provider,omitempty"`
	Cost            *gateway.CostCalculation `json:"cost,omitempty"`
	RoutingDecision *string                `json:"routing_decision,omitempty"`
}

type CompletionChoice struct {
	Index        int         `json:"index"`
	Text         string      `json:"text"`
	FinishReason *string     `json:"finish_reason"`
	LogProbs     interface{} `json:"logprobs"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
	
	// Brokle extensions
	Provider        *string                `json:"provider,omitempty"`
	Cost            *gateway.CostCalculation `json:"cost,omitempty"`
	RoutingDecision *string                `json:"routing_decision,omitempty"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type EmbeddingUsage struct {
	PromptTokens int32 `json:"prompt_tokens"`
	TotalTokens  int32 `json:"total_tokens"`
}

type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Helper functions for request/response transformation

func (h *Handler) extractOrganizationID(c *gin.Context) (ulid.ULID, error) {
	// Extract organization ID from JWT token or API key context
	// This would typically be set by authentication middleware
	orgIDStr := c.GetString("organization_id")
	if orgIDStr == "" {
		return ulid.ULID{}, fmt.Errorf("organization ID not found in context")
	}
	
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("invalid organization ID format: %w", err)
	}
	
	return orgID, nil
}

func (h *Handler) extractEnvironment(c *gin.Context) gateway.Environment {
	// Extract environment from header or default to production
	env := c.GetHeader("X-Brokle-Environment")
	if env == "" {
		env = "production"
	}
	
	switch strings.ToLower(env) {
	case "development", "dev":
		return gateway.EnvironmentDevelopment
	case "staging", "stage":
		return gateway.EnvironmentStaging
	case "production", "prod":
		return gateway.EnvironmentProduction
	default:
		return gateway.EnvironmentProduction
	}
}

// Validation functions

func (h *Handler) validateChatCompletionRequest(req *ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages array cannot be empty")
	}
	
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message[%d].role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message[%d].content is required", i)
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return fmt.Errorf("message[%d].role must be one of: system, user, assistant", i)
		}
	}
	
	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	return nil
}

func (h *Handler) validateCompletionRequest(req *CompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if req.Prompt == nil {
		return fmt.Errorf("prompt is required")
	}
	
	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	return nil
}

func (h *Handler) validateEmbeddingsRequest(req *EmbeddingRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if req.Input == nil {
		return fmt.Errorf("input is required")
	}
	
	return nil
}

func (h *Handler) validateRouteRequest(req *RouteRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	// Either messages or prompt should be provided for cost estimation
	if len(req.Messages) == 0 && (req.Prompt == nil || *req.Prompt == "") {
		return fmt.Errorf("either messages or prompt must be provided for routing decision")
	}
	
	// Validate strategy if provided
	if req.Strategy != nil {
		validStrategies := []string{"cost_optimized", "latency_optimized", "quality_optimized", "reliability_optimized"}
		found := false
		for _, valid := range validStrategies {
			if *req.Strategy == valid {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid strategy: must be one of cost_optimized, latency_optimized, quality_optimized, reliability_optimized")
		}
	}
	
	return nil
}

func (h *Handler) validateInvalidateCacheRequest(req *InvalidateCacheRequest) error {
	// At least one invalidation criteria must be specified
	hasTarget := (req.Provider != nil && *req.Provider != "") ||
		(req.Model != nil && *req.Model != "") ||
		(len(req.Keys) > 0) ||
		(req.ClearAll != nil && *req.ClearAll) ||
		(req.MaxAge != nil && *req.MaxAge > 0) ||
		(req.Pattern != nil && *req.Pattern != "")
	
	if !hasTarget {
		return fmt.Errorf("at least one invalidation criteria must be specified (provider, model, keys, clear_all, max_age, or pattern)")
	}
	
	// Validate MaxAge if provided
	if req.MaxAge != nil && *req.MaxAge < 0 {
		return fmt.Errorf("max_age must be non-negative")
	}
	
	// Warn about clear_all
	if req.ClearAll != nil && *req.ClearAll {
		h.logger.Warn("Cache clear_all requested - this will invalidate all cached entries")
	}
	
	return nil
}

// Request transformation functions

func (h *Handler) transformChatCompletionRequest(req *ChatCompletionRequest, orgID ulid.ULID, env gateway.Environment) *gateway.ChatCompletionRequest {
	messages := make([]gateway.ChatMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = gateway.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}
	
	gatewayReq := &gateway.ChatCompletionRequest{
		OrganizationID: orgID,
		Environment:    env,
		Model:          req.Model,
		Messages:       messages,
		Stream:         req.Stream != nil && *req.Stream,
	}
	
	if req.MaxTokens != nil {
		gatewayReq.MaxTokens = int32(*req.MaxTokens)
	}
	if req.Temperature != nil {
		gatewayReq.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		gatewayReq.TopP = *req.TopP
	}
	if req.N != nil {
		gatewayReq.N = int32(*req.N)
	}
	if req.PresencePenalty != nil {
		gatewayReq.PresencePenalty = *req.PresencePenalty
	}
	if req.FrequencyPenalty != nil {
		gatewayReq.FrequencyPenalty = *req.FrequencyPenalty
	}
	if req.User != nil {
		gatewayReq.User = *req.User
	}
	
	return gatewayReq
}

func (h *Handler) transformCompletionRequest(req *CompletionRequest, orgID ulid.ULID, env gateway.Environment) *gateway.CompletionRequest {
	gatewayReq := &gateway.CompletionRequest{
		OrganizationID: orgID,
		Environment:    env,
		Model:          req.Model,
		Prompt:         req.Prompt,
		Stream:         req.Stream != nil && *req.Stream,
	}
	
	if req.MaxTokens != nil {
		gatewayReq.MaxTokens = int32(*req.MaxTokens)
	}
	if req.Temperature != nil {
		gatewayReq.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		gatewayReq.TopP = *req.TopP
	}
	if req.N != nil {
		gatewayReq.N = int32(*req.N)
	}
	if req.Echo != nil {
		gatewayReq.Echo = *req.Echo
	}
	if req.PresencePenalty != nil {
		gatewayReq.PresencePenalty = *req.PresencePenalty
	}
	if req.FrequencyPenalty != nil {
		gatewayReq.FrequencyPenalty = *req.FrequencyPenalty
	}
	if req.User != nil {
		gatewayReq.User = *req.User
	}
	
	return gatewayReq
}

func (h *Handler) transformEmbeddingsRequest(req *EmbeddingRequest, orgID ulid.ULID, env gateway.Environment) *gateway.EmbeddingRequest {
	return &gateway.EmbeddingRequest{
		Model:           req.Model,
		Input:           req.Input,
		EncodingFormat:  req.EncodingFormat,
		User:            req.User,
	}
}

func (h *Handler) transformRouteRequest(req *RouteRequest, orgID ulid.ULID, env gateway.Environment) *gateway.RoutingRequest {
	gatewayReq := &gateway.RoutingRequest{
		OrganizationID: orgID,
		Environment:    env,
		Model:          req.Model,
		Metadata:       req.Metadata,
	}
	
	// Set preferred provider if specified
	if req.Provider != nil && *req.Provider != "" {
		gatewayReq.PreferredProvider = *req.Provider
	}
	
	// Set routing strategy
	if req.Strategy != nil {
		gatewayReq.Strategy = *req.Strategy
	} else {
		gatewayReq.Strategy = "cost_optimized" // Default strategy
	}
	
	// Transform messages for context estimation
	if len(req.Messages) > 0 {
		messages := make([]gateway.ChatMessage, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = gateway.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
				Name:    msg.Name,
			}
		}
		gatewayReq.Messages = messages
	}
	
	// Set prompt if provided
	if req.Prompt != nil {
		gatewayReq.Prompt = *req.Prompt
	}
	
	// Set parameters for cost estimation
	if req.MaxTokens != nil {
		gatewayReq.MaxTokens = int32(*req.MaxTokens)
	}
	if req.Temperature != nil {
		gatewayReq.Temperature = *req.Temperature
	}
	
	return gatewayReq
}

func (h *Handler) transformInvalidateCacheRequest(req *InvalidateCacheRequest, orgID ulid.ULID, env gateway.Environment) *gateway.InvalidateCacheRequest {
	gatewayReq := &gateway.InvalidateCacheRequest{
		OrganizationID: orgID,
		Environment:    env,
		Keys:          req.Keys,
	}
	
	if req.Provider != nil {
		gatewayReq.Provider = *req.Provider
	}
	if req.Model != nil {
		gatewayReq.Model = *req.Model
	}
	if req.ClearAll != nil {
		gatewayReq.ClearAll = *req.ClearAll
	}
	if req.MaxAge != nil {
		gatewayReq.MaxAge = int64(*req.MaxAge)
	}
	if req.Pattern != nil {
		gatewayReq.Pattern = *req.Pattern
	}
	
	return gatewayReq
}

// Response transformation functions

func (h *Handler) transformChatCompletionResponse(resp *gateway.ChatCompletionResponse) *ChatCompletionResponse {
	choices := make([]ChatCompletionChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = ChatCompletionChoice{
			Index: choice.Index,
			Message: ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
				Name:    choice.Message.Name,
			},
			FinishReason: &choice.FinishReason,
		}
	}
	
	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: ChatCompletionUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Provider:        &resp.Provider,
		Cost:            resp.Cost,
		RoutingDecision: &resp.RoutingDecision,
	}
}

func (h *Handler) transformCompletionResponse(resp *gateway.CompletionResponse) *CompletionResponse {
	choices := make([]CompletionChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = CompletionChoice{
			Index:        choice.Index,
			Text:         choice.Text,
			FinishReason: &choice.FinishReason,
		}
	}
	
	return &CompletionResponse{
		ID:      resp.ID,
		Object:  "text_completion",
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: ChatCompletionUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Provider:        &resp.Provider,
		Cost:            resp.Cost,
		RoutingDecision: &resp.RoutingDecision,
	}
}

func (h *Handler) transformEmbeddingsResponse(resp *gateway.EmbeddingsResponse) *EmbeddingResponse {
	data := make([]EmbeddingData, len(resp.Data))
	for i, embedding := range resp.Data {
		data[i] = EmbeddingData{
			Object:    "embedding",
			Index:     embedding.Index,
			Embedding: embedding.Embedding,
		}
	}
	
	return &EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  resp.Model,
		Usage: EmbeddingUsage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
		Provider:        &resp.Provider,
		Cost:            resp.Cost,
		RoutingDecision: &resp.RoutingDecision,
	}
}

func (h *Handler) transformModelsResponse(resp *gateway.ListModelsResponse) *ModelsResponse {
	models := make([]Model, len(resp.Data))
	for i, model := range resp.Data {
		models[i] = Model{
			ID:      model.ID,
			Object:  "model",
			Created: model.Created,
			OwnedBy: model.OwnedBy,
		}
	}
	
	return &ModelsResponse{
		Object: "list",
		Data:   models,
	}
}

func (h *Handler) transformRouteResponse(resp *gateway.RoutingResponse) *RouteResponse {
	return &RouteResponse{
		Provider:         resp.Provider,
		Model:            resp.Model,
		Endpoint:         resp.Endpoint,
		Strategy:         resp.Strategy,
		EstimatedCost:    resp.EstimatedCost,
		EstimatedLatency: resp.EstimatedLatency,
		QualityScore:     resp.QualityScore,
		CacheHit:         resp.CacheHit,
		ProviderHealth:   resp.ProviderHealth,
		Metadata:         resp.Metadata,
	}
}

func (h *Handler) transformCacheStatusResponse(resp *gateway.CacheStatusResponse) *CacheStatusResponse {
	return &CacheStatusResponse{
		Status:            resp.Status,
		HitRate:           resp.HitRate,
		TotalEntries:      resp.TotalEntries,
		SizeBytes:         resp.SizeBytes,
		MemoryUsage:       resp.MemoryUsage,
		EvictionCount:     resp.EvictionCount,
		LastEviction:      resp.LastEviction,
		ProviderBreakdown: resp.ProviderBreakdown,
	}
}

func (h *Handler) transformInvalidateCacheResponse(resp *gateway.InvalidateCacheResponse) *InvalidateCacheResponse {
	apiResp := &InvalidateCacheResponse{
		Success:         resp.Success,
		InvalidatedKeys: resp.InvalidatedKeys,
		Count:           resp.Count,
		Message:         resp.Message,
	}
	
	if resp.Error != "" {
		apiResp.Error = &resp.Error
	}
	
	return apiResp
}

// Streaming handlers for SSE support

// handleChatCompletionStream handles streaming chat completion responses using SSE
func (h *Handler) handleChatCompletionStream(c *gin.Context, orgID ulid.ULID, environment string, req *gateway.ChatCompletionRequest) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint": "/v1/chat/completions",
		"stream": true,
		"organization_id": orgID,
		"environment": environment,
	})

	logger.Info("Starting streaming chat completion")

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// Create context for streaming
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Use gin's response writer directly for streaming
	err := h.gatewayService.CreateChatCompletionStream(ctx, orgID, environment, req, c.Writer)
	if err != nil {
		logger.WithError(err).Error("Streaming chat completion failed")
		h.sendSSEError(c, err)
		return
	}

	logger.Info("Streaming chat completion finished successfully")
}

// handleCompletionStream handles streaming completion responses using SSE
func (h *Handler) handleCompletionStream(c *gin.Context, orgID ulid.ULID, environment string, req *gateway.CompletionRequest) {
	logger := h.logger.WithFields(logrus.Fields{
		"endpoint": "/v1/completions",
		"stream": true,
		"organization_id": orgID,
		"environment": environment,
	})

	logger.Info("Starting streaming completion")

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// Create context for streaming
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Use gin's response writer directly for streaming
	err := h.gatewayService.CreateCompletionStream(ctx, orgID, environment, req, c.Writer)
	if err != nil {
		logger.WithError(err).Error("Streaming completion failed")
		h.sendSSEError(c, err)
		return
	}

	logger.Info("Streaming completion finished successfully")
}

// SSE helper functions

func (h *Handler) sendSSEData(c *gin.Context, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal SSE data")
		return
	}

	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", jsonData))
	c.Writer.Flush()
}

func (h *Handler) sendSSEError(c *gin.Context, err error) {
	errorData := map[string]interface{}{
		"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "server_error",
		},
	}

	jsonData, jsonErr := json.Marshal(errorData)
	if jsonErr != nil {
		h.logger.WithError(jsonErr).Error("Failed to marshal SSE error")
		return
	}

	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", jsonData))
	c.Writer.Flush()
}

// Error handling

func (h *Handler) handleGatewayError(c *gin.Context, err error) {
	// Map gateway errors to appropriate HTTP responses
	switch err {
	case gateway.ErrModelNotFound:
		response.ErrorWithCode(c, http.StatusNotFound, "model_not_found", "The requested model was not found")
	case gateway.ErrProviderNotFound:
		response.ErrorWithCode(c, http.StatusNotFound, "provider_not_found", "No suitable provider was found")
	case gateway.ErrProviderDisabled:
		response.ErrorWithCode(c, http.StatusServiceUnavailable, "provider_disabled", "The provider is currently disabled")
	case gateway.ErrProviderConfigInactive:
		response.ErrorWithCode(c, http.StatusServiceUnavailable, "provider_config_inactive", "The provider configuration is inactive")
	case gateway.ErrUnsupportedOperation:
		response.ErrorWithCode(c, http.StatusBadRequest, "unsupported_operation", "The operation is not supported by the selected provider")
	case gateway.ErrNoProvidersAvailable:
		response.ErrorWithCode(c, http.StatusServiceUnavailable, "no_providers_available", "No providers are currently available")
	case gateway.ErrNoFallbackProviders:
		response.ErrorWithCode(c, http.StatusServiceUnavailable, "no_fallback_providers", "No fallback providers are available")
	default:
		h.logger.WithError(err).Error("Unhandled gateway error")
		response.ErrorWithCode(c, http.StatusInternalServerError, "internal_server_error", "An internal error occurred")
	}
}
