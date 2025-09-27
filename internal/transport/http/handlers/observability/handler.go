package observability

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	obsServices "brokle/internal/services/observability"
)

// Handler contains all observability-related HTTP handlers
type Handler struct {
	config   *config.Config
	logger   *logrus.Logger
	services *obsServices.ServiceRegistry
}

// CreateEventRequest represents telemetry event creation request
// @Description Request data for creating telemetry events
type CreateEventRequest struct {
	EventType    string                 `json:"event_type" example:"ai.request.completed" description:"Type of telemetry event"`
	Timestamp    *int64                 `json:"timestamp,omitempty" example:"1677610602" description:"Unix timestamp (defaults to current time)"`
	SessionID    *string                `json:"session_id,omitempty" example:"sess_abc123" description:"Session identifier"`
	TraceID      *string                `json:"trace_id,omitempty" example:"trace_def456" description:"Distributed trace ID"`
	SpanID       *string                `json:"span_id,omitempty" example:"span_ghi789" description:"Span identifier"`
	UserID       *string                `json:"user_id,omitempty" example:"user_123" description:"User identifier"`
	Properties   map[string]interface{} `json:"properties,omitempty" description:"Event-specific properties"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" description:"Additional metadata"`
	Tags         []string               `json:"tags,omitempty" description:"Event tags for categorization"`
	Environment  *string                `json:"environment,omitempty" example:"production" description:"Environment name"`
	Version      *string                `json:"version,omitempty" example:"1.0.0" description:"Application version"`
	Source       *string                `json:"source,omitempty" example:"python-sdk" description:"Event source (sdk, api, etc.)"`
}

// EventResponse represents telemetry event creation response
// @Description Telemetry event creation result
type EventResponse struct {
	EventID     string `json:"event_id" example:"evt_abc123" description:"Created event identifier"`
	Status      string `json:"status" example:"created" description:"Event creation status"`
	ProcessedAt int64  `json:"processed_at" example:"1677610602" description:"Unix timestamp when event was processed"`
	Message     string `json:"message" example:"Event created successfully" description:"Operation result message"`
}

// NewHandler creates a new observability handler
func NewHandler(
	cfg *config.Config,
	logger *logrus.Logger,
	services *obsServices.ServiceRegistry,
) *Handler {
	return &Handler{
		config:   cfg,
		logger:   logger,
		services: services,
	}
}

// CreateEvent handles POST /v1/events
// @Summary Create a telemetry event
// @Description Create a new telemetry event for SDK observability
// @Tags SDK - Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateEventRequest true "Event creation data"
// @Success 201 {object} response.SuccessResponse{data=EventResponse} "Event created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid or missing API key"
// @Failure 422 {object} response.ErrorResponse "Validation failed"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/events [post]
func (h *Handler) CreateEvent(c *gin.Context) {
	h.logger.Info("CreateEvent handler called - placeholder implementation")

	// Placeholder response for now
	c.JSON(200, gin.H{
		"message": "Event endpoint placeholder - implementation pending",
		"path":    "/v1/events",
	})
}