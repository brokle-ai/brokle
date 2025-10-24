package observability

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	obsServices "brokle/internal/services/observability"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
)

// OTLPHandler handles OTLP HTTP requests
type OTLPHandler struct {
	telemetryService observability.TelemetryService
	otlpConverter    *obsServices.OTLPConverterService
	logger           *logrus.Logger
}

// NewOTLPHandler creates a new OTLP handler
func NewOTLPHandler(
	telemetryService observability.TelemetryService,
	otlpConverter *obsServices.OTLPConverterService,
	logger *logrus.Logger,
) *OTLPHandler {
	return &OTLPHandler{
		telemetryService: telemetryService,
		otlpConverter:    otlpConverter,
		logger:           logger,
	}
}

// HandleTraces handles POST /v1/traces
// @Summary OTLP trace ingestion endpoint (OpenTelemetry spec compliant)
// @Description Accepts OpenTelemetry Protocol (OTLP) traces in JSON or Protobuf format
// @Tags SDK - OTLP
// @Accept json
// @Accept application/x-protobuf
// @Produce json
// @Security ApiKeyAuth
// @Param request body observability.OTLPRequest true "OTLP trace export request"
// @Success 200 {object} response.APIResponse{data=map[string]interface{}} "Traces accepted"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid OTLP request"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Invalid or missing API key"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /v1/traces [post]
func (h *OTLPHandler) HandleTraces(c *gin.Context) {
	ctx := c.Request.Context()

	// Get project ID from SDK auth middleware (already authenticated)
	projectIDPtr, exists := middleware.GetProjectID(c)
	if !exists || projectIDPtr == nil {
		h.logger.Error("Project ID not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}
	projectID := projectIDPtr.String()

	// Get environment from context (optional)
	environment, _ := middleware.GetEnvironment(c)

	// Parse OTLP request (JSON format)
	// TODO: Add protobuf support later
	var otlpReq observability.OTLPRequest
	if err := c.ShouldBindJSON(&otlpReq); err != nil {
		h.logger.WithError(err).Error("Failed to parse OTLP request")
		response.ValidationError(c, "invalid OTLP request", err.Error())
		return
	}

	// Validate request has resource spans
	if len(otlpReq.ResourceSpans) == 0 {
		response.ValidationError(c, "empty request", "OTLP request must contain at least one resource span")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":     projectID,
		"resource_spans": len(otlpReq.ResourceSpans),
	}).Debug("Received OTLP trace request")

	// Convert OTLP spans to Brokle telemetry events using converter service
	brokleEvents, err := h.otlpConverter.ConvertOTLPToBrokleEvents(&otlpReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert OTLP to Brokle events")
		response.InternalServerError(c, "Failed to process OTLP traces")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":    projectID,
		"otlp_spans":    countSpans(&otlpReq),
		"brokle_events": len(brokleEvents),
	}).Debug("Converted OTLP spans to Brokle events")

	// Construct Brokle telemetry batch request
	batchReq := &observability.TelemetryBatchRequest{
		ProjectID:   *projectIDPtr,
		Environment: &environment,
		Events:      brokleEvents,
		Metadata: map[string]interface{}{
			"source":         "otlp",
			"otlp_version":   "1.0",
			"resource_spans": len(otlpReq.ResourceSpans),
		},
		Async: false, // Process synchronously for OTLP (low latency)
	}

	// Process batch using existing telemetry service infrastructure
	batchResp, err := h.telemetryService.ProcessTelemetryBatch(ctx, batchReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process telemetry batch")
		response.Error(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"batch_id":         batchResp.BatchID,
		"processed_events": batchResp.ProcessedEvents,
		"duplicate_events": batchResp.DuplicateEvents,
	}).Info("OTLP traces processed successfully")

	// Return OTLP-compatible success response
	response.Success(c, map[string]interface{}{
		"status":          "success",
		"batch_id":        batchResp.BatchID,
		"processed_spans": batchResp.ProcessedEvents,
	})
}

// countSpans counts total spans in OTLP request
func countSpans(req *observability.OTLPRequest) int {
	count := 0
	for _, rs := range req.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			count += len(ss.Spans)
		}
	}
	return count
}
