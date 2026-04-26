// OTLP HTTP ingestion (traces, logs, metrics).
//
// The OTLP HTTP endpoints accept raw protobuf (and protobuf-JSON) bodies
// per the OpenTelemetry spec. They are mounted in internal/server/routes.go
// alongside the rest of the SDK plane and MUST be wired under the chi Group
// that attaches middleware.RequireSDKAuth + middleware.LimitByAPIKey.
//
// Endpoints:
//
//	POST /v1/traces   — OTLP trace export (protobuf or application/json)
//	POST /v1/logs     — OTLP logs export
//	POST /v1/metrics  — OTLP metrics export
//
// Response shape: Brokle's Stripe/OpenAI-style success envelope (raw
// resource on 200, {"error":{...}} on 4xx/5xx) rather than strict OTLP
// ExportXxxServiceResponse protobuf. Brokle SDKs consume the standard
// envelope; third-party OTLP senders accept a 200 as success per the
// spec's "at-least-once" semantics.
package observability

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/streams"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/uid"
)

const otlpMaxRequestSize = 10 * 1024 * 1024 // 10MB, matches OTEL Collector default

// readOTLPBody validates Content-Type, enforces the 10MB cap, and
// transparently decompresses gzip. Returns the decoded content bytes and
// the matched content type ("application/x-protobuf" or
// "application/json").
func readOTLPBody(w http.ResponseWriter, r *http.Request, logger *slog.Logger, endpoint string) ([]byte, string, bool) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/x-protobuf") && !strings.Contains(contentType, "application/json") {
		logger.Warn("Unsupported Content-Type for OTLP endpoint", "endpoint", endpoint, "content_type", contentType)
		writeOTLPError(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Content-Type must be 'application/x-protobuf' or 'application/json'", "")
		return nil, "", false
	}

	r.Body = http.MaxBytesReader(w, r.Body, otlpMaxRequestSize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			logger.Warn("OTLP request exceeds maximum size limit", "endpoint", endpoint, "max_size", otlpMaxRequestSize)
			writeOTLPError(w, http.StatusRequestEntityTooLarge, "payload_too_large",
				fmt.Sprintf("Request body exceeds maximum size of %d bytes", otlpMaxRequestSize), "")
			return nil, "", false
		}
		logger.Error("Failed to read OTLP request body", "endpoint", endpoint, "error", err)
		writeOTLPError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body", "")
		return nil, "", false
	}

	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		originalSize := len(body)
		gzipReader, gerr := gzip.NewReader(bytes.NewReader(body))
		if gerr != nil {
			logger.Error("Failed to create gzip reader", "endpoint", endpoint, "error", gerr)
			writeOTLPError(w, http.StatusBadRequest, "invalid_encoding", "Failed to decompress gzip data", "")
			return nil, "", false
		}
		defer func() {
			if cerr := gzipReader.Close(); cerr != nil {
				logger.Warn("Failed to close gzip reader", "endpoint", endpoint, "error", cerr)
			}
		}()
		body, err = io.ReadAll(gzipReader)
		if err != nil {
			logger.Error("Failed to decompress gzip data", "endpoint", endpoint, "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_encoding", "Failed to read decompressed data", "")
			return nil, "", false
		}
		logger.Info("Gzip decompression successful", "endpoint", endpoint, "original_size", originalSize, "decompressed_size", len(body))
	}

	return body, contentType, true
}

// writeOTLPError writes an APIResponse error envelope directly to the
// stdlib ResponseWriter (no gin.Context in scope).
func writeOTLPError(w http.ResponseWriter, status int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]any{
			"type":    string(appErrors.FromHTTPStatus(status, message).Type),
			"code":    code,
			"message": message,
			"details": details,
		},
	})
}

// writeOTLPSuccess writes a 200 APIResponse envelope with the given data
// payload.
func writeOTLPSuccess(w http.ResponseWriter, data map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}

// writeOTLPServiceError maps an AppError / generic error to an HTTP status
// and APIResponse envelope. Used after the protobuf is parsed but a
// downstream service fails.
func writeOTLPServiceError(w http.ResponseWriter, err error) {
	response.WriteError(w, err)
}

// ==================================================================
// TRACES
// ==================================================================

func (h *OTLPHandler) HandleTraces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.deps.Logger

	projectUUID := httpctx.MustGetProjectID(ctx)
	projectID := projectUUID.String()
	organizationUUID := httpctx.MustGetOrganizationID(ctx)

	body, contentType, ok := readOTLPBody(w, r, logger, "traces")
	if !ok {
		return
	}

	var protoReq coltracepb.ExportTraceServiceRequest
	if strings.Contains(contentType, "application/x-protobuf") {
		if err := proto.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to unmarshal OTLP traces protobuf", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_protobuf", "Invalid OTLP protobuf", err.Error())
			return
		}
	} else {
		if err := protojson.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to parse OTLP traces JSON", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_json", "Invalid OTLP JSON", err.Error())
			return
		}
	}

	otlpReq, err := convertProtoToInternal(&protoReq)
	if err != nil {
		logger.Error("Failed to convert protobuf to internal format", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process OTLP traces", "")
		return
	}
	if len(otlpReq.ResourceSpans) == 0 {
		writeOTLPError(w, http.StatusBadRequest, "empty_request", "OTLP request must contain at least one resource span", "")
		return
	}

	brokleEvents, err := h.deps.OTLPConverter.ConvertOTLPToBrokleEvents(ctx, &otlpReq, projectID)
	if err != nil {
		logger.Error("Failed to convert OTLP to Brokle events", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process OTLP traces", "")
		return
	}

	// Deduplication across spans
	dedupIDs := make([]string, 0, len(brokleEvents))
	dedupIDToFirstIndex := make(map[string]int)
	for i, event := range brokleEvents {
		if event.EventType == observability.TelemetryEventTypeSpan {
			if event.SpanID == "" {
				logger.Error("Span missing span_id, skipping deduplication", "event_id", event.EventID.String(), "trace_id", event.TraceID)
				continue
			}
			dedupID := fmt.Sprintf("%s:%s", event.TraceID, event.SpanID)
			dedupIDs = append(dedupIDs, dedupID)
			if _, exists := dedupIDToFirstIndex[dedupID]; !exists {
				dedupIDToFirstIndex[dedupID] = i
			}
		}
	}

	batchID := uid.New()
	var claimedIDs, duplicateIDs []string
	if len(dedupIDs) > 0 {
		claimedIDs, duplicateIDs, err = h.deps.DeduplicationService.ClaimEvents(ctx, projectUUID, batchID, dedupIDs, 24*time.Hour)
		if err != nil {
			logger.Error("Failed to claim OTLP spans for deduplication", "error", err)
			writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to claim events for deduplication", "")
			return
		}
	}

	hasTraces := false
	for _, event := range brokleEvents {
		if event.EventType == observability.TelemetryEventTypeTrace {
			hasTraces = true
			break
		}
	}

	if len(claimedIDs) == 0 && !hasTraces {
		logger.Info("All OTLP spans were duplicates, skipping", "project_id", projectID, "duplicates", len(duplicateIDs))
		writeOTLPSuccess(w, map[string]any{"status": "all_duplicates", "duplicate_spans": len(duplicateIDs)})
		return
	}

	claimedSet := make(map[string]bool, len(claimedIDs))
	for _, id := range claimedIDs {
		claimedSet[id] = true
	}

	claimedEventData := make([]streams.TelemetryEventData, 0, len(brokleEvents))
	for i, event := range brokleEvents {
		if event.EventType == observability.TelemetryEventTypeTrace {
			claimedEventData = append(claimedEventData, streams.TelemetryEventData{
				EventID:      event.EventID,
				SpanID:       event.SpanID,
				TraceID:      event.TraceID,
				EventType:    string(event.EventType),
				EventPayload: event.Payload,
			})
			continue
		}
		if event.EventType == observability.TelemetryEventTypeSpan {
			dedupID := fmt.Sprintf("%s:%s", event.TraceID, event.SpanID)
			firstIndex := dedupIDToFirstIndex[dedupID]
			isFirstOccurrence := (i == firstIndex)
			if isFirstOccurrence && claimedSet[dedupID] {
				claimedEventData = append(claimedEventData, streams.TelemetryEventData{
					EventID:      event.EventID,
					SpanID:       event.SpanID,
					TraceID:      event.TraceID,
					EventType:    string(event.EventType),
					EventPayload: event.Payload,
				})
			}
		}
	}

	streamMsg := &streams.TelemetryStreamMessage{
		BatchID:          batchID,
		ProjectID:        projectUUID,
		OrganizationID:   organizationUUID,
		Events:           claimedEventData,
		ClaimedSpanIDs:   claimedIDs,
		DuplicateSpanIDs: duplicateIDs,
		Metadata: map[string]any{
			"source":         "otlp",
			"content_type":   contentType,
			"resource_spans": len(otlpReq.ResourceSpans),
			"total_spans":    countSpans(&otlpReq),
		},
		Timestamp: time.Now(),
	}

	streamID, err := h.deps.StreamProducer.PublishBatch(ctx, streamMsg)
	if err != nil {
		if rollbackErr := h.deps.DeduplicationService.ReleaseEvents(ctx, claimedIDs); rollbackErr != nil {
			logger.Error("CRITICAL: Failed to rollback OTLP deduplication claims after publish failure",
				"rollback_error", rollbackErr.Error(), "original_error", err.Error(), "batch_id", batchID.String())
		}
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to publish events to stream", "")
		return
	}

	logger.Info("OTLP traces published to stream successfully",
		"batch_id", batchID.String(), "stream_id", streamID,
		"claimed_events", len(claimedIDs), "duplicates", len(duplicateIDs), "project_id", projectID)

	writeOTLPSuccess(w, map[string]any{
		"status":          "accepted",
		"batch_id":        batchID.String(),
		"stream_id":       streamID,
		"processed_spans": len(claimedIDs),
		"duplicate_spans": len(duplicateIDs),
	})
}

// ==================================================================
// LOGS
// ==================================================================

func (h *OTLPHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.deps.Logger

	projectID := httpctx.MustGetProjectID(ctx)

	body, contentType, ok := readOTLPBody(w, r, logger, "logs")
	if !ok {
		return
	}

	var protoReq collogspb.ExportLogsServiceRequest
	if strings.Contains(contentType, "application/x-protobuf") {
		if err := proto.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to unmarshal OTLP logs protobuf", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_protobuf", "Invalid OTLP protobuf", err.Error())
			return
		}
	} else {
		if err := protojson.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to parse OTLP logs JSON", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_json", "Invalid OTLP JSON", err.Error())
			return
		}
	}

	if len(protoReq.GetResourceLogs()) == 0 {
		writeOTLPError(w, http.StatusBadRequest, "empty_request", "OTLP request must contain at least one resource logs", "")
		return
	}

	logsData := &logspb.LogsData{ResourceLogs: protoReq.GetResourceLogs()}

	logEvents, err := h.deps.LogsConverter.ConvertLogsRequest(ctx, logsData, projectID)
	if err != nil {
		logger.Error("Failed to convert OTLP logs to Brokle events", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process OTLP logs", "")
		return
	}
	genaiEvents, err := h.deps.EventsConverter.ConvertGenAIEventsRequest(ctx, logsData, projectID)
	if err != nil {
		logger.Error("Failed to convert GenAI events to Brokle events", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process GenAI events", "")
		return
	}

	brokleEvents := append(logEvents, genaiEvents...)

	eventData := make([]streams.TelemetryEventData, 0, len(brokleEvents))
	for _, event := range brokleEvents {
		eventData = append(eventData, streams.TelemetryEventData{
			EventID:      event.EventID,
			SpanID:       event.SpanID,
			TraceID:      event.TraceID,
			EventType:    string(event.EventType),
			EventPayload: event.Payload,
		})
	}

	batchID := uid.New()
	streamMessage := &streams.TelemetryStreamMessage{
		BatchID:   batchID,
		ProjectID: projectID,
		Events:    eventData,
		Timestamp: uid.TimeFromID(batchID),
	}

	streamID, err := h.deps.StreamProducer.PublishBatch(ctx, streamMessage)
	if err != nil {
		logger.Error("Failed to publish logs batch to Redis Streams", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process logs batch", "")
		return
	}

	logger.Info("Successfully published OTLP logs batch to Redis Streams",
		"project_id", projectID.String(), "batch_id", batchID.String(), "stream_id", streamID, "event_count", len(brokleEvents))

	writeOTLPSuccess(w, map[string]any{
		"batch_id":    batchID.String(),
		"event_count": len(brokleEvents),
		"status":      "accepted",
	})
}

// ==================================================================
// METRICS
// ==================================================================

func (h *OTLPHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.deps.Logger

	projectID := httpctx.MustGetProjectID(ctx)

	body, contentType, ok := readOTLPBody(w, r, logger, "metrics")
	if !ok {
		return
	}

	var protoReq colmetricspb.ExportMetricsServiceRequest
	if strings.Contains(contentType, "application/x-protobuf") {
		if err := proto.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to unmarshal OTLP metrics protobuf", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_protobuf", "Invalid OTLP protobuf", err.Error())
			return
		}
	} else {
		if err := protojson.Unmarshal(body, &protoReq); err != nil {
			logger.Error("Failed to parse OTLP metrics JSON", "error", err)
			writeOTLPError(w, http.StatusBadRequest, "invalid_otlp_json", "Invalid OTLP JSON", err.Error())
			return
		}
	}

	if len(protoReq.GetResourceMetrics()) == 0 {
		writeOTLPError(w, http.StatusBadRequest, "empty_request", "OTLP request must contain at least one resource metrics", "")
		return
	}

	metricsData := &metricspb.MetricsData{ResourceMetrics: protoReq.GetResourceMetrics()}
	brokleEvents, err := h.deps.MetricsConverter.ConvertMetricsRequest(ctx, metricsData, projectID)
	if err != nil {
		logger.Error("Failed to convert OTLP metrics to Brokle events", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process OTLP metrics", "")
		return
	}

	eventData := make([]streams.TelemetryEventData, 0, len(brokleEvents))
	for _, event := range brokleEvents {
		eventData = append(eventData, streams.TelemetryEventData{
			EventID:      event.EventID,
			SpanID:       event.SpanID,
			TraceID:      event.TraceID,
			EventType:    string(event.EventType),
			EventPayload: event.Payload,
		})
	}

	batchID := uid.New()
	streamMessage := &streams.TelemetryStreamMessage{
		BatchID:   batchID,
		ProjectID: projectID,
		Events:    eventData,
		Timestamp: uid.TimeFromID(batchID),
	}

	streamID, err := h.deps.StreamProducer.PublishBatch(ctx, streamMessage)
	if err != nil {
		logger.Error("Failed to publish metrics batch to Redis Streams", "error", err)
		writeOTLPError(w, http.StatusInternalServerError, "internal_error", "Failed to process metrics batch", "")
		return
	}

	logger.Info("Successfully published OTLP metrics batch to Redis Streams",
		"project_id", projectID.String(), "batch_id", batchID.String(), "stream_id", streamID, "event_count", len(brokleEvents))

	writeOTLPSuccess(w, map[string]any{
		"batch_id":    batchID.String(),
		"event_count": len(brokleEvents),
		"status":      "accepted",
	})
}

// ==================================================================
// Protobuf → internal conversion (trace path)
// ==================================================================

// countSpans counts total spans in OTLP request.
func countSpans(req *observability.OTLPRequest) int {
	count := 0
	for _, rs := range req.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			count += len(ss.Spans)
		}
	}
	return count
}

// convertProtoToInternal converts official OTLP protobuf to the internal
// Brokle OTLPRequest shape.
func convertProtoToInternal(protoReq *coltracepb.ExportTraceServiceRequest) (observability.OTLPRequest, error) {
	var internalReq observability.OTLPRequest

	for _, protoRS := range protoReq.ResourceSpans {
		internalRS := observability.ResourceSpan{}

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

		for _, protoSS := range protoRS.ScopeSpans {
			internalSS := observability.ScopeSpan{}

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

			for _, protoSpan := range protoSS.Spans {
				traceIDHex := hex.EncodeToString(protoSpan.TraceId)
				spanIDHex := hex.EncodeToString(protoSpan.SpanId)
				var parentSpanIDHex any
				if len(protoSpan.ParentSpanId) > 0 {
					parentSpanIDHex = hex.EncodeToString(protoSpan.ParentSpanId)
				}

				internalSpan := observability.OTLPSpan{
					TraceID:           traceIDHex,
					SpanID:            spanIDHex,
					ParentSpanID:      parentSpanIDHex,
					Name:              protoSpan.Name,
					Kind:              int(protoSpan.Kind),
					StartTimeUnixNano: int64(protoSpan.StartTimeUnixNano),
					EndTimeUnixNano:   int64(protoSpan.EndTimeUnixNano),
				}

				for _, attr := range protoSpan.Attributes {
					internalSpan.Attributes = append(internalSpan.Attributes, observability.KeyValue{
						Key:   attr.Key,
						Value: convertProtoAnyValue(attr.Value),
					})
				}

				if protoSpan.Status != nil {
					internalSpan.Status = &observability.Status{
						Code:    int(protoSpan.Status.Code),
						Message: protoSpan.Status.Message,
					}
				}

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

// convertProtoAnyValue converts a protobuf AnyValue to a Go `any`.
func convertProtoAnyValue(value *commonpb.AnyValue) any {
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
		arr := make([]any, len(v.ArrayValue.Values))
		for i, item := range v.ArrayValue.Values {
			arr[i] = convertProtoAnyValue(item)
		}
		return arr
	case *commonpb.AnyValue_KvlistValue:
		if v.KvlistValue == nil {
			return nil
		}
		m := make(map[string]any)
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

// writeOTLPServiceError is kept for future use if callers ever need to
// propagate an AppError directly from within an OTLP handler — currently
// unused because we emit specific error codes at each failure point. The
// function is intentionally retained so new failure paths have a drop-in
// error-renderer that matches the rest of the stack.
var _ = writeOTLPServiceError
