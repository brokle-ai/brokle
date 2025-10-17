package observability

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// TelemetryBatchRequest represents the high-throughput telemetry batch request
// @Description High-performance batch request for telemetry events with ULID-based deduplication
type TelemetryBatchRequest struct {
	Environment   *string                      `json:"environment,omitempty" example:"production" description:"Environment tag (optional)"`
	Metadata      map[string]interface{}       `json:"metadata,omitempty" description:"Batch-level metadata"`
	Events        []*TelemetryEventRequest     `json:"events" binding:"required,min=1,max=1000" description:"Array of telemetry events (max 1000)"`
	Async         bool                         `json:"async,omitempty" description:"Process batch asynchronously"`
	Deduplication *DeduplicationConfigRequest  `json:"deduplication,omitempty" description:"Deduplication configuration"`
}

// TelemetryEventRequest represents an individual telemetry event in a batch
// @Description Individual telemetry event using envelope pattern for high throughput
type TelemetryEventRequest struct {
	EventID   string                 `json:"event_id" binding:"required" example:"01ABCDEFGHIJKLMNOPQRSTUVWXYZ" description:"ULID event identifier"`
	EventType string                 `json:"event_type" binding:"required" example:"trace_create" description:"Type of telemetry event"`
	Payload   map[string]interface{} `json:"payload" binding:"required" description:"Event payload data"`
	Timestamp *int64                 `json:"timestamp,omitempty" example:"1677610602" description:"Unix timestamp (defaults to current time)"`
}

// DeduplicationConfigRequest represents deduplication configuration
// @Description Configuration for event deduplication behavior
type DeduplicationConfigRequest struct {
	Enabled          bool `json:"enabled" example:"true" description:"Enable deduplication"`
	TTL              int  `json:"ttl,omitempty" example:"3600" description:"Deduplication TTL in seconds"`
	UseRedisCache    bool `json:"use_redis_cache" example:"true" description:"Use Redis cache for deduplication"`
	FailOnDuplicate  bool `json:"fail_on_duplicate,omitempty" description:"Fail request on duplicate detection"`
}

// TelemetryBatchResponse represents the response for telemetry batch processing
// @Description Response for high-throughput telemetry batch processing
type TelemetryBatchResponse struct {
	BatchID           string     `json:"batch_id" example:"01ABCDEFGHIJKLMNOPQRSTUVWXYZ" description:"Generated batch identifier"`
	ProcessedEvents   int        `json:"processed_events" example:"95" description:"Number of successfully processed events"`
	DuplicateEvents   int        `json:"duplicate_events" example:"3" description:"Number of duplicate events skipped"`
	FailedEvents      int        `json:"failed_events" example:"2" description:"Number of failed events"`
	ProcessingTimeMs  int        `json:"processing_time_ms" example:"123" description:"Total processing time in milliseconds"`
	Errors            []TelemetryEventError `json:"errors,omitempty" description:"Processing errors if any"`
	DuplicateEventIDs []string   `json:"duplicate_event_ids,omitempty" description:"IDs of duplicate events"`
	JobID             *string    `json:"job_id,omitempty" example:"job_01ABC123" description:"Async job ID if async=true"`
}

// TelemetryEventError represents an error processing a telemetry event
// @Description Error details for failed telemetry event processing
type TelemetryEventError struct {
	EventID string `json:"event_id" example:"01ABCDEFGHIJKLMNOPQRSTUVWXYZ" description:"Event ID that failed"`
	Error   string `json:"error" example:"Invalid payload format" description:"Error message"`
	Details string `json:"details,omitempty" description:"Additional error details"`
}

// TelemetryHealthResponse represents telemetry service health status
// @Description Health status of telemetry processing system
type TelemetryHealthResponse struct {
	Healthy               bool                         `json:"healthy" example:"true" description:"Overall health status"`
	Database              *DatabaseHealthResponse      `json:"database,omitempty" description:"Database health status"`
	Redis                 *RedisHealthResponse         `json:"redis,omitempty" description:"Redis health status"`
	ProcessingQueue       *QueueHealthResponse         `json:"processing_queue,omitempty" description:"Processing queue health"`
	ActiveWorkers         int                          `json:"active_workers" example:"5" description:"Number of active workers"`
	AverageProcessingTime float64                      `json:"average_processing_time_ms" example:"45.7" description:"Average processing time in milliseconds"`
	ThroughputPerMinute   float64                      `json:"throughput_per_minute" example:"1200.5" description:"Events processed per minute"`
	ErrorRate             float64                      `json:"error_rate" example:"0.01" description:"Error rate (0.0-1.0)"`
}

// DatabaseHealthResponse represents database health
// @Description Database connectivity and performance status
type DatabaseHealthResponse struct {
	Connected         bool    `json:"connected" example:"true" description:"Database connection status"`
	LatencyMs         float64 `json:"latency_ms" example:"1.5" description:"Database latency in milliseconds"`
	ActiveConnections int     `json:"active_connections" example:"10" description:"Active database connections"`
	MaxConnections    int     `json:"max_connections" example:"100" description:"Maximum database connections"`
}

// RedisHealthResponse represents Redis health
// @Description Redis connectivity and performance status
type RedisHealthResponse struct {
	Available   bool    `json:"available" example:"true" description:"Redis availability status"`
	LatencyMs   float64 `json:"latency_ms" example:"0.5" description:"Redis latency in milliseconds"`
	Connections int     `json:"connections" example:"5" description:"Active Redis connections"`
	LastError   *string `json:"last_error,omitempty" description:"Last Redis error if any"`
	Uptime      string  `json:"uptime" example:"24h0m0s" description:"Redis uptime duration"`
}

// QueueHealthResponse represents processing queue health
// @Description Processing queue status and performance
type QueueHealthResponse struct {
	Size             int64   `json:"size" example:"150" description:"Current queue size"`
	ProcessingRate   float64 `json:"processing_rate" example:"100.5" description:"Processing rate per second"`
	AverageWaitTime  float64 `json:"average_wait_time_ms" example:"10.2" description:"Average wait time in milliseconds"`
	OldestMessageAge float64 `json:"oldest_message_age_ms" example:"500.0" description:"Age of oldest message in milliseconds"`
}

// TelemetryMetricsResponse represents comprehensive telemetry metrics
// @Description Comprehensive metrics for telemetry processing system
type TelemetryMetricsResponse struct {
	TotalBatches         int64   `json:"total_batches" example:"1250" description:"Total batches processed"`
	CompletedBatches     int64   `json:"completed_batches" example:"1230" description:"Successfully completed batches"`
	FailedBatches        int64   `json:"failed_batches" example:"15" description:"Failed batches"`
	ProcessingBatches    int64   `json:"processing_batches" example:"5" description:"Currently processing batches"`
	TotalEvents          int64   `json:"total_events" example:"125000" description:"Total events processed"`
	ProcessedEvents      int64   `json:"processed_events" example:"123000" description:"Successfully processed events"`
	FailedEvents         int64   `json:"failed_events" example:"1500" description:"Failed events"`
	DuplicateEvents      int64   `json:"duplicate_events" example:"500" description:"Duplicate events found"`
	AverageEventsPerBatch float64 `json:"average_events_per_batch" example:"100.0" description:"Average events per batch"`
	ThroughputPerSecond  float64 `json:"throughput_per_second" example:"85.5" description:"Events processed per second"`
	SuccessRate          float64 `json:"success_rate" example:"99.2" description:"Success rate percentage"`
	DeduplicationRate    float64 `json:"deduplication_rate" example:"0.4" description:"Deduplication rate percentage"`
}

// TelemetryPerformanceStatsResponse represents performance statistics
// @Description Performance statistics over a time window
type TelemetryPerformanceStatsResponse struct {
	TimeWindow           string  `json:"time_window" example:"1h" description:"Time window for statistics"`
	TotalRequests        int64   `json:"total_requests" example:"1200" description:"Total requests in time window"`
	SuccessfulRequests   int64   `json:"successful_requests" example:"1185" description:"Successful requests"`
	AverageLatencyMs     float64 `json:"average_latency_ms" example:"45.7" description:"Average latency in milliseconds"`
	P95LatencyMs         float64 `json:"p95_latency_ms" example:"89.2" description:"95th percentile latency"`
	P99LatencyMs         float64 `json:"p99_latency_ms" example:"156.8" description:"99th percentile latency"`
	ThroughputPerSecond  float64 `json:"throughput_per_second" example:"85.5" description:"Throughput per second"`
	PeakThroughput       float64 `json:"peak_throughput" example:"120.3" description:"Peak throughput observed"`
	CacheHitRate         float64 `json:"cache_hit_rate" example:"0.85" description:"Cache hit rate (0.0-1.0)"`
	DatabaseFallbackRate float64 `json:"database_fallback_rate" example:"0.15" description:"Database fallback rate"`
	ErrorRate            float64 `json:"error_rate" example:"0.01" description:"Error rate (0.0-1.0)"`
	RetryRate            float64 `json:"retry_rate" example:"0.02" description:"Retry rate (0.0-1.0)"`
}

// ProcessTelemetryBatch handles POST /v1/telemetry/batch
// @Summary Process high-throughput telemetry batch (async via Redis Streams)
// @Description Process a batch of telemetry events asynchronously with ULID-based deduplication and Redis Streams. Returns 202 Accepted immediately while events are processed in the background.
// @Tags SDK - Telemetry
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body TelemetryBatchRequest true "Telemetry batch data"
// @Success 202 {object} response.APIResponse{data=TelemetryBatchResponse} "Batch accepted for async processing"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request payload"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 422 {object} response.APIResponse{error=response.APIError} "Validation failed"
// @Failure 429 {object} response.APIResponse{error=response.APIError} "Rate limit exceeded"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/telemetry/batch [post]
func (h *Handler) ProcessTelemetryBatch(c *gin.Context) {
	var req TelemetryBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid telemetry batch request JSON")
		response.ValidationError(c, "Invalid request payload", err.Error())
		return
	}

	// Comprehensive request validation
	if validationErrors := ValidateTelemetryBatchRequest(&req); len(validationErrors) > 0 {
		h.logger.WithField("validation_errors", validationErrors).Error("Telemetry batch request validation failed")
		response.ValidationError(c, "Request validation failed", FormatValidationErrors(validationErrors))
		return
	}

	// Validate request size
	if err := ValidateRequestSize(&req); err != nil {
		h.logger.WithError(err).Error("Telemetry batch request size exceeded")
		response.ValidationError(c, "Request too large", err.Error())
		return
	}

	// Get project ID from SDK authentication context
	projectIDPtr, exists := c.Get("project_id")
	if !exists {
		h.logger.Error("Project ID not found in context")
		response.Unauthorized(c, "Authentication context missing")
		return
	}

	projectID, ok := projectIDPtr.(*ulid.ULID)
	if !ok || projectID == nil {
		h.logger.Error("Invalid project ID type in context")
		response.Unauthorized(c, "Invalid authentication context")
		return
	}

	// Get optional environment from header with validation
	environment := c.GetHeader("X-Environment")
	if environment == "" && req.Environment != nil {
		environment = *req.Environment
	}

	// Validate environment tag if provided
	if environment != "" {
		if err := ValidateEnvironmentTag(environment); err != nil {
			h.logger.WithError(err).WithField("environment", environment).Error("Invalid environment tag")
			response.ValidationError(c, "Invalid environment tag", err.Error())
			return
		}
	}

	// Convert request to domain request
	domainEvents := make([]*observability.TelemetryEventRequest, len(req.Events))
	for i, event := range req.Events {
		// Parse ULID
		eventID, err := ulid.Parse(event.EventID)
		if err != nil {
			h.logger.WithError(err).WithField("event_id", event.EventID).Error("Invalid event ID")
			response.ValidationError(c, "Invalid event ID", fmt.Sprintf("Event at index %d has invalid ULID format", i))
			return
		}

		// Convert event type
		eventType := observability.TelemetryEventType(event.EventType)

		// Set timestamp
		var timestamp *time.Time
		if event.Timestamp != nil {
			t := time.Unix(*event.Timestamp, 0)
			timestamp = &t
		}

		domainEvents[i] = &observability.TelemetryEventRequest{
			EventID:   eventID,
			EventType: eventType,
			Payload:   SanitizeMetadata(event.Payload), // Also sanitize event payloads
			Timestamp: timestamp,
		}
	}

	// Build domain request with sanitized metadata
	domainReq := &observability.TelemetryBatchRequest{
		ProjectID:   *projectID,
		Environment: func() *string {
			if environment != "" {
				return &environment
			}
			return nil
		}(),
		Metadata: SanitizeMetadata(req.Metadata),
		Events:   domainEvents,
		Async:    req.Async,
	}

	// Add deduplication config if provided
	if req.Deduplication != nil {
		domainReq.Deduplication = &observability.DeduplicationConfig{
			Enabled:         req.Deduplication.Enabled,
			TTL:            time.Duration(req.Deduplication.TTL) * time.Second,
			UseRedisCache:  req.Deduplication.UseRedisCache,
			FailOnDuplicate: req.Deduplication.FailOnDuplicate,
		}
	}

	// Process batch through telemetry service
	resp, err := h.services.GetTelemetryService().ProcessTelemetryBatch(c.Request.Context(), domainReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process telemetry batch")
		response.Error(c, err)
		return
	}

	// Convert domain response to API response
	apiResp := &TelemetryBatchResponse{
		BatchID:           resp.BatchID.String(),
		ProcessedEvents:   resp.ProcessedEvents,
		DuplicateEvents:   resp.DuplicateEvents,
		FailedEvents:      resp.FailedEvents,
		ProcessingTimeMs:  resp.ProcessingTimeMs,
		DuplicateEventIDs: func() []string {
			ids := make([]string, len(resp.DuplicateEventIDs))
			for i, id := range resp.DuplicateEventIDs {
				ids[i] = id.String()
			}
			return ids
		}(),
	}

	// Add errors if any
	if len(resp.Errors) > 0 {
		apiResp.Errors = make([]TelemetryEventError, len(resp.Errors))
		for i, err := range resp.Errors {
			apiResp.Errors[i] = TelemetryEventError{
				EventID: err.EventID.String(),
				Error:   err.ErrorMessage,
				Details: err.ErrorCode,
			}
		}
	}

	// Add job ID for async processing
	if resp.JobID != nil {
		apiResp.JobID = resp.JobID
	}

	h.logger.WithFields(logrus.Fields{
		"batch_id":         apiResp.BatchID,
		"processed_events": apiResp.ProcessedEvents,
		"duplicate_events": apiResp.DuplicateEvents,
		"failed_events":    apiResp.FailedEvents,
		"processing_time":  apiResp.ProcessingTimeMs,
	}).Info("Telemetry batch accepted for async processing")

	// Return 202 Accepted for async processing via Redis Streams
	response.Accepted(c, apiResp)
}

// GetTelemetryHealth handles GET /v1/telemetry/health
// @Summary Get telemetry service health status
// @Description Get comprehensive health status of telemetry processing system
// @Tags SDK - Telemetry
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.APIResponse{data=TelemetryHealthResponse} "Health status retrieved successfully"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/telemetry/health [get]
func (h *Handler) GetTelemetryHealth(c *gin.Context) {
	health, err := h.services.GetTelemetryService().GetHealth(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get telemetry health")
		response.Error(c, err)
		return
	}

	// Convert domain response to API response
	apiResp := &TelemetryHealthResponse{
		Healthy:               health.Healthy,
		ActiveWorkers:         health.ActiveWorkers,
		AverageProcessingTime: health.AverageProcessingTime,
		ThroughputPerMinute:   health.ThroughputPerMinute,
		ErrorRate:             health.ErrorRate,
	}

	// Add database health
	if health.Database != nil {
		apiResp.Database = &DatabaseHealthResponse{
			Connected:         health.Database.Connected,
			LatencyMs:         health.Database.LatencyMs,
			ActiveConnections: health.Database.ActiveConnections,
			MaxConnections:    health.Database.MaxConnections,
		}
	}

	// Add Redis health
	if health.Redis != nil {
		apiResp.Redis = &RedisHealthResponse{
			Available:   health.Redis.Available,
			LatencyMs:   health.Redis.LatencyMs,
			Connections: health.Redis.Connections,
			LastError:   health.Redis.LastError,
			Uptime:      health.Redis.Uptime.String(),
		}
	}

	// Add queue health
	if health.ProcessingQueue != nil {
		apiResp.ProcessingQueue = &QueueHealthResponse{
			Size:             health.ProcessingQueue.Size,
			ProcessingRate:   health.ProcessingQueue.ProcessingRate,
			AverageWaitTime:  health.ProcessingQueue.AverageWaitTime,
			OldestMessageAge: health.ProcessingQueue.OldestMessageAge,
		}
	}

	response.Success(c, apiResp)
}

// GetTelemetryMetrics handles GET /v1/telemetry/metrics
// @Summary Get telemetry service metrics
// @Description Get comprehensive metrics for telemetry processing system
// @Tags SDK - Telemetry
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.APIResponse{data=TelemetryMetricsResponse} "Metrics retrieved successfully"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/telemetry/metrics [get]
func (h *Handler) GetTelemetryMetrics(c *gin.Context) {
	metrics, err := h.services.GetTelemetryService().GetMetrics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get telemetry metrics")
		response.Error(c, err)
		return
	}

	// Convert domain response to API response
	apiResp := &TelemetryMetricsResponse{
		TotalBatches:         metrics.TotalBatches,
		CompletedBatches:     metrics.CompletedBatches,
		FailedBatches:        metrics.FailedBatches,
		ProcessingBatches:    metrics.ProcessingBatches,
		TotalEvents:          metrics.TotalEvents,
		ProcessedEvents:      metrics.ProcessedEvents,
		FailedEvents:         metrics.FailedEvents,
		DuplicateEvents:      metrics.DuplicateEvents,
		AverageEventsPerBatch: metrics.AverageEventsPerBatch,
		ThroughputPerSecond:  metrics.ThroughputPerSecond,
		SuccessRate:          metrics.SuccessRate,
		DeduplicationRate:    metrics.DeduplicationRate,
	}

	response.Success(c, apiResp)
}

// GetTelemetryPerformanceStats handles GET /v1/telemetry/performance
// @Summary Get telemetry performance statistics
// @Description Get performance statistics for telemetry processing over a time window
// @Tags SDK - Telemetry
// @Produce json
// @Security ApiKeyAuth
// @Param window query string false "Time window for statistics" Enums(1m,5m,15m,1h,6h,24h) default(1h)
// @Success 200 {object} response.APIResponse{data=TelemetryPerformanceStatsResponse} "Performance stats retrieved successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid time window parameter"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/telemetry/performance [get]
func (h *Handler) GetTelemetryPerformanceStats(c *gin.Context) {
	// Parse time window parameter
	windowStr := c.DefaultQuery("window", "1h")
	timeWindow, err := time.ParseDuration(windowStr)
	if err != nil {
		response.ValidationError(c, "Invalid time window", "Time window must be a valid duration (e.g., 1m, 5m, 1h, 24h)")
		return
	}

	// Validate time window (1 minute to 7 days)
	if timeWindow < time.Minute || timeWindow > 7*24*time.Hour {
		response.ValidationError(c, "Invalid time window", "Time window must be between 1m and 168h (7 days)")
		return
	}

	stats, err := h.services.GetTelemetryService().GetPerformanceStats(c.Request.Context(), timeWindow)
	if err != nil {
		h.logger.WithError(err).WithField("time_window", windowStr).Error("Failed to get telemetry performance stats")
		response.Error(c, err)
		return
	}

	// Convert domain response to API response
	apiResp := &TelemetryPerformanceStatsResponse{
		TimeWindow:           stats.TimeWindow.String(),
		TotalRequests:        stats.TotalRequests,
		SuccessfulRequests:   stats.SuccessfulRequests,
		AverageLatencyMs:     stats.AverageLatencyMs,
		P95LatencyMs:         stats.P95LatencyMs,
		P99LatencyMs:         stats.P99LatencyMs,
		ThroughputPerSecond:  stats.ThroughputPerSecond,
		PeakThroughput:       stats.PeakThroughput,
		CacheHitRate:         stats.CacheHitRate,
		DatabaseFallbackRate: stats.DatabaseFallbackRate,
		ErrorRate:            stats.ErrorRate,
		RetryRate:            stats.RetryRate,
	}

	response.Success(c, apiResp)
}

// ValidateEvent handles POST /v1/telemetry/validate
// @Summary Validate telemetry event structure
// @Description Validate a telemetry event without processing it
// @Tags SDK - Telemetry
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body TelemetryEventRequest true "Event to validate"
// @Success 200 {object} response.APIResponse{data=TelemetryValidationResponse} "Event validation result"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request payload"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 422 {object} response.APIResponse{error=response.APIError} "Validation failed"
// @Router /v1/telemetry/validate [post]
func (h *Handler) ValidateEvent(c *gin.Context) {
	var req TelemetryEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid telemetry event validation request")
		response.ValidationError(c, "Invalid request payload", err.Error())
		return
	}

	// Parse ULID
	eventID, err := ulid.Parse(req.EventID)
	if err != nil {
		response.ValidationError(c, "Invalid event ID", "Event ID must be a valid ULID")
		return
	}

	// Validate event type
	eventType := observability.TelemetryEventType(req.EventType)
	validTypes := []observability.TelemetryEventType{
		observability.TelemetryEventTypeTraceCreate,
		observability.TelemetryEventTypeTraceUpdate,
		observability.TelemetryEventTypeObservationCreate,
		observability.TelemetryEventTypeObservationUpdate,
		observability.TelemetryEventTypeObservationComplete,
		observability.TelemetryEventTypeQualityScoreCreate,
	}

	isValidType := false
	for _, validType := range validTypes {
		if eventType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.ValidationError(c, "Invalid event type", fmt.Sprintf("Event type must be one of: %v", validTypes))
		return
	}

	// Validate payload
	if len(req.Payload) == 0 {
		response.ValidationError(c, "Empty payload", "Event payload cannot be empty")
		return
	}

	// Build validation response
	apiResp := &TelemetryValidationResponse{
		Valid:     true,
		EventID:   eventID.String(),
		EventType: string(eventType),
		Message:   "Event structure is valid",
	}

	response.Success(c, apiResp)
}

// TelemetryValidationResponse represents event validation result
// @Description Result of telemetry event validation
type TelemetryValidationResponse struct {
	Valid     bool   `json:"valid" example:"true" description:"Whether the event is valid"`
	EventID   string `json:"event_id" example:"01ABCDEFGHIJKLMNOPQRSTUVWXYZ" description:"Validated event ID"`
	EventType string `json:"event_type" example:"trace_create" description:"Validated event type"`
	Message   string `json:"message" example:"Event structure is valid" description:"Validation result message"`
	Errors    []string `json:"errors,omitempty" description:"Validation errors if any"`
}