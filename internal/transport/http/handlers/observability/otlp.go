package observability

import (
	"encoding/hex"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

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

	// Read raw request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read OTLP request body")
		response.BadRequest(c, "invalid request", "Failed to read request body")
		return
	}

	// Detect content type and parse accordingly
	contentType := c.GetHeader("Content-Type")
	var otlpReq observability.OTLPRequest

	if strings.Contains(contentType, "application/x-protobuf") {
		// Parse Protobuf format (more efficient)
		h.logger.Debug("Parsing OTLP Protobuf request")

		var protoReq coltracepb.ExportTraceServiceRequest
		if err := proto.Unmarshal(body, &protoReq); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal OTLP protobuf")
			response.ValidationError(c, "invalid OTLP protobuf", err.Error())
			return
		}

		// Convert protobuf to internal format
		otlpReq, err = convertProtoToInternal(&protoReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to convert protobuf to internal format")
			response.InternalServerError(c, "Failed to process OTLP protobuf")
			return
		}
	} else {
		// Parse JSON format (default, for debugging)
		h.logger.Debug("Parsing OTLP JSON request")

		var protoReq coltracepb.ExportTraceServiceRequest
		if err := protojson.Unmarshal(body, &protoReq); err != nil {
			h.logger.WithError(err).Error("Failed to parse OTLP JSON")
			response.ValidationError(c, "invalid OTLP JSON", err.Error())
			return
		}

		// Convert protobuf to internal format (same as Protobuf path)
		otlpReq, err = convertProtoToInternal(&protoReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to convert JSON to internal format")
			response.InternalServerError(c, "Failed to process OTLP JSON")
			return
		}
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

// convertProtoToInternal converts official OTLP protobuf to internal format
func convertProtoToInternal(protoReq *coltracepb.ExportTraceServiceRequest) (observability.OTLPRequest, error) {
	var internalReq observability.OTLPRequest

	for _, protoRS := range protoReq.ResourceSpans {
		internalRS := observability.ResourceSpan{}

		// Convert Resource
		if protoRS.Resource != nil {
			internalResource := &observability.Resource{}
			for _, attr := range protoRS.Resource.Attributes {
				internalResource.Attributes = append(internalResource.Attributes, observability.KeyValue{
					Key:   attr.Key,
					Value: convertProtoAnyValue(attr.Value),
				})
			}
			internalRS.Resource = internalResource
		}

		// Convert ScopeSpans
		for _, protoSS := range protoRS.ScopeSpans {
			internalSS := observability.ScopeSpan{}

			// Convert Scope
			if protoSS.Scope != nil {
				internalScope := &observability.Scope{
					Name:    protoSS.Scope.Name,
					Version: protoSS.Scope.Version,
				}
				for _, attr := range protoSS.Scope.Attributes {
					internalScope.Attributes = append(internalScope.Attributes, observability.KeyValue{
						Key:   attr.Key,
						Value: convertProtoAnyValue(attr.Value),
					})
				}
				internalSS.Scope = internalScope
			}

			// Convert Spans
			for _, protoSpan := range protoSS.Spans {
				// Convert byte arrays to hex strings for internal format
				traceIDHex := hex.EncodeToString(protoSpan.TraceId)
				spanIDHex := hex.EncodeToString(protoSpan.SpanId)
				var parentSpanIDHex interface{}
				if len(protoSpan.ParentSpanId) > 0 {
					parentSpanIDHex = hex.EncodeToString(protoSpan.ParentSpanId)
				}

				internalSpan := observability.Span{
					TraceID:           traceIDHex,
					SpanID:            spanIDHex,
					ParentSpanID:      parentSpanIDHex,
					Name:              protoSpan.Name,
					Kind:              int(protoSpan.Kind),
					StartTimeUnixNano: int64(protoSpan.StartTimeUnixNano),
					EndTimeUnixNano:   int64(protoSpan.EndTimeUnixNano),
				}

				// Convert Attributes
				for _, attr := range protoSpan.Attributes {
					internalSpan.Attributes = append(internalSpan.Attributes, observability.KeyValue{
						Key:   attr.Key,
						Value: convertProtoAnyValue(attr.Value),
					})
				}

				// Convert Status
				if protoSpan.Status != nil {
					internalSpan.Status = &observability.Status{
						Code:    int(protoSpan.Status.Code),
						Message: protoSpan.Status.Message,
					}
				}

				// Convert Events
				for _, protoEvent := range protoSpan.Events {
					internalEvent := observability.Event{
						TimeUnixNano: int64(protoEvent.TimeUnixNano),
						Name:         protoEvent.Name,
					}
					for _, attr := range protoEvent.Attributes {
						internalEvent.Attributes = append(internalEvent.Attributes, observability.KeyValue{
							Key:   attr.Key,
							Value: convertProtoAnyValue(attr.Value),
						})
					}
					internalSpan.Events = append(internalSpan.Events, internalEvent)
				}

				internalSS.Spans = append(internalSS.Spans, internalSpan)
			}

			internalRS.ScopeSpans = append(internalRS.ScopeSpans, internalSS)
		}

		internalReq.ResourceSpans = append(internalReq.ResourceSpans, internalRS)
	}

	return internalReq, nil
}

// convertProtoAnyValue converts protobuf AnyValue to interface{}
func convertProtoAnyValue(value *commonpb.AnyValue) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue
	case *commonpb.AnyValue_IntValue:
		return v.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *commonpb.AnyValue_ArrayValue:
		if v.ArrayValue == nil {
			return nil
		}
		arr := make([]interface{}, len(v.ArrayValue.Values))
		for i, item := range v.ArrayValue.Values {
			arr[i] = convertProtoAnyValue(item)
		}
		return arr
	case *commonpb.AnyValue_KvlistValue:
		if v.KvlistValue == nil {
			return nil
		}
		m := make(map[string]interface{})
		for _, kv := range v.KvlistValue.Values {
			m[kv.Key] = convertProtoAnyValue(kv.Value)
		}
		return m
	case *commonpb.AnyValue_BytesValue:
		return v.BytesValue
	default:
		return nil
	}
}
