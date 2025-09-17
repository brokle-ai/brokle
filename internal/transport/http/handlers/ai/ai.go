package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"brokle/internal/config"
	"brokle/pkg/response"
)

// Handler handles AI Gateway endpoints (OpenAI-compatible)
type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

// NewHandler creates a new AI handler
func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{config: config, logger: logger}
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

// ChatCompletions handles OpenAI-compatible chat completions
// @Summary Create chat completion
// @Description Generate AI chat completions using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security KeyPairAuth
// @Param request body ChatCompletionRequest true "Chat completion request"
// @Success 200 {object} response.SuccessResponse "Chat completion generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/chat/completions [post]
func (h *Handler) ChatCompletions(c *gin.Context) { 
	response.Success(c, gin.H{"message": "Chat completions - TODO"}) 
}

// Completions handles OpenAI-compatible text completions
// @Summary Create text completion
// @Description Generate AI text completions using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security KeyPairAuth
// @Param request body CompletionRequest true "Text completion request"
// @Success 200 {object} response.SuccessResponse "Text completion generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/completions [post]
func (h *Handler) Completions(c *gin.Context) { 
	response.Success(c, gin.H{"message": "Completions - TODO"}) 
}

// Embeddings handles OpenAI-compatible embeddings
// @Summary Create embeddings
// @Description Generate text embeddings using OpenAI-compatible API
// @Tags AI Gateway
// @Accept json
// @Produce json
// @Security KeyPairAuth
// @Param request body EmbeddingRequest true "Embedding request"
// @Success 200 {object} response.SuccessResponse "Embeddings generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 429 {object} response.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/embeddings [post]
func (h *Handler) Embeddings(c *gin.Context) { 
	response.Success(c, gin.H{"message": "Embeddings - TODO"}) 
}

// ListModels handles listing available AI models
// @Summary List available models
// @Description Get list of available AI models
// @Tags AI Gateway
// @Produce json
// @Security KeyPairAuth
// @Success 200 {array} Model "List of available models"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/models [get]
func (h *Handler) ListModels(c *gin.Context) { 
	response.Success(c, gin.H{"message": "List models - TODO"}) 
}

// GetModel handles getting specific model information
// @Summary Get model information
// @Description Get detailed information about a specific AI model
// @Tags AI Gateway
// @Produce json
// @Security KeyPairAuth
// @Param model path string true "Model ID" example("gpt-3.5-turbo")
// @Success 200 {object} Model "Model information"
// @Failure 401 {object} response.ErrorResponse "Invalid API key"
// @Failure 404 {object} response.ErrorResponse "Model not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/models/{model} [get]
func (h *Handler) GetModel(c *gin.Context) { 
	response.Success(c, gin.H{"message": "Get model - TODO"}) 
}