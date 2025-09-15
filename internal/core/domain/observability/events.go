package observability

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// EventType represents the type of observability event
type EventType string

const (
	// Trace events
	EventTypeTraceCreated   EventType = "trace.created"
	EventTypeTraceUpdated   EventType = "trace.updated"
	EventTypeTraceDeleted   EventType = "trace.deleted"
	EventTypeTraceCompleted EventType = "trace.completed"

	// Observation events
	EventTypeObservationCreated   EventType = "observation.created"
	EventTypeObservationUpdated   EventType = "observation.updated"
	EventTypeObservationCompleted EventType = "observation.completed"
	EventTypeObservationDeleted   EventType = "observation.deleted"

	// Quality events
	EventTypeQualityScoreAdded   EventType = "quality.score.added"
	EventTypeQualityScoreUpdated EventType = "quality.score.updated"
	EventTypeQualityScoreDeleted EventType = "quality.score.deleted"
	EventTypeEvaluationStarted   EventType = "quality.evaluation.started"
	EventTypeEvaluationCompleted EventType = "quality.evaluation.completed"
	EventTypeEvaluationFailed    EventType = "quality.evaluation.failed"

	// System events
	EventTypeBatchIngestionStarted   EventType = "system.batch.ingestion.started"
	EventTypeBatchIngestionCompleted EventType = "system.batch.ingestion.completed"
	EventTypeBatchIngestionFailed    EventType = "system.batch.ingestion.failed"
	EventTypeAlertTriggered          EventType = "system.alert.triggered"
	EventTypeThresholdExceeded       EventType = "system.threshold.exceeded"
)

// Event represents a domain event in the observability system
type Event struct {
	ID          ulid.ULID              `json:"id"`
	Type        EventType              `json:"type"`
	Source      string                 `json:"source"`
	Subject     string                 `json:"subject"`     // The resource being affected (trace_id, observation_id, etc.)
	ProjectID   ulid.ULID              `json:"project_id"`
	UserID      *ulid.ULID             `json:"user_id,omitempty"`
	Data        map[string]any         `json:"data"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Correlation *CorrelationInfo       `json:"correlation,omitempty"`
}

// CorrelationInfo represents correlation information for events
type CorrelationInfo struct {
	TraceID       *ulid.ULID `json:"trace_id,omitempty"`
	ObservationID *ulid.ULID `json:"observation_id,omitempty"`
	SessionID     *ulid.ULID `json:"session_id,omitempty"`
	RequestID     *string    `json:"request_id,omitempty"`
	UserID        *ulid.ULID `json:"user_id,omitempty"`
}

// EventHandler represents an event handler interface
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	EventTypes() []EventType
	Name() string
}

// EventPublisher represents an event publisher interface
type EventPublisher interface {
	Publish(ctx context.Context, event *Event) error
	PublishBatch(ctx context.Context, events []*Event) error
}

// EventSubscriber represents an event subscriber interface
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventTypes []EventType, handler EventHandler) error
	Unsubscribe(ctx context.Context, handlerName string) error
}

// EventStore represents an event store interface
type EventStore interface {
	Store(ctx context.Context, event *Event) error
	StoreBatch(ctx context.Context, events []*Event) error
	GetEvents(ctx context.Context, filter *EventFilter) ([]*Event, error)
	GetEventsBySubject(ctx context.Context, subject string) ([]*Event, error)
	GetEventsByType(ctx context.Context, eventType EventType, limit, offset int) ([]*Event, error)
}

// EventFilter represents filters for event queries
type EventFilter struct {
	EventTypes  []EventType `json:"event_types,omitempty"`
	ProjectID   *ulid.ULID  `json:"project_id,omitempty"`
	UserID      *ulid.ULID  `json:"user_id,omitempty"`
	Subject     *string     `json:"subject,omitempty"`
	Source      *string     `json:"source,omitempty"`
	StartTime   *time.Time  `json:"start_time,omitempty"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
}

// Event builder functions

// NewTraceCreatedEvent creates a new trace created event
func NewTraceCreatedEvent(trace *Trace, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeTraceCreated,
		Source:    "observability.trace_service",
		Subject:   trace.ID.String(),
		ProjectID: trace.ProjectID,
		UserID:    userID,
		Data: map[string]any{
			"trace_id":          trace.ID,
			"external_trace_id": trace.ExternalTraceID,
			"name":              trace.Name,
			"session_id":        trace.SessionID,
			"observation_count": len(trace.Observations),
		},
		Metadata: map[string]any{
			"tags":     trace.Tags,
			"metadata": trace.Metadata,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID:   &trace.ID,
			SessionID: trace.SessionID,
			UserID:    trace.UserID,
		},
	}
}

// NewTraceUpdatedEvent creates a new trace updated event
func NewTraceUpdatedEvent(trace *Trace, userID *ulid.ULID, changes map[string]any) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeTraceUpdated,
		Source:    "observability.trace_service",
		Subject:   trace.ID.String(),
		ProjectID: trace.ProjectID,
		UserID:    userID,
		Data: map[string]any{
			"trace_id":          trace.ID,
			"external_trace_id": trace.ExternalTraceID,
			"changes":           changes,
		},
		Metadata: map[string]any{
			"updated_at": trace.UpdatedAt,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID:   &trace.ID,
			SessionID: trace.SessionID,
			UserID:    trace.UserID,
		},
	}
}

// NewTraceDeletedEvent creates a new trace deleted event
func NewTraceDeletedEvent(traceID ulid.ULID, projectID ulid.ULID, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeTraceDeleted,
		Source:    "observability.trace_service",
		Subject:   traceID.String(),
		ProjectID: projectID,
		UserID:    userID,
		Data: map[string]any{
			"trace_id": traceID,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID: &traceID,
			UserID:  userID,
		},
	}
}

// NewObservationCreatedEvent creates a new observation created event
func NewObservationCreatedEvent(observation *Observation, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeObservationCreated,
		Source:    "observability.observation_service",
		Subject:   observation.ID.String(),
		ProjectID: getProjectIDFromContext(observation.TraceID), // This would need to be resolved
		UserID:    userID,
		Data: map[string]any{
			"observation_id":          observation.ID,
			"external_observation_id": observation.ExternalObservationID,
			"trace_id":                observation.TraceID,
			"type":                    observation.Type,
			"name":                    observation.Name,
			"provider":                observation.Provider,
			"model":                   observation.Model,
		},
		Metadata: map[string]any{
			"start_time": observation.StartTime,
			"level":      observation.Level,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID:       &observation.TraceID,
			ObservationID: &observation.ID,
			UserID:        userID,
		},
	}
}

// NewObservationCompletedEvent creates a new observation completed event
func NewObservationCompletedEvent(observation *Observation, userID *ulid.ULID) *Event {
	var latency *int
	if observation.EndTime != nil {
		latencyMs := int(observation.EndTime.Sub(observation.StartTime).Milliseconds())
		latency = &latencyMs
	}

	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeObservationCompleted,
		Source:    "observability.observation_service",
		Subject:   observation.ID.String(),
		ProjectID: getProjectIDFromContext(observation.TraceID),
		UserID:    userID,
		Data: map[string]any{
			"observation_id":    observation.ID,
			"trace_id":          observation.TraceID,
			"type":              observation.Type,
			"provider":          observation.Provider,
			"model":             observation.Model,
			"latency_ms":        latency,
			"total_tokens":      observation.TotalTokens,
			"total_cost":        observation.TotalCost,
			"quality_score":     observation.QualityScore,
		},
		Metadata: map[string]any{
			"start_time": observation.StartTime,
			"end_time":   observation.EndTime,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID:       &observation.TraceID,
			ObservationID: &observation.ID,
			UserID:        userID,
		},
	}
}

// NewQualityScoreAddedEvent creates a new quality score added event
func NewQualityScoreAddedEvent(score *QualityScore, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeQualityScoreAdded,
		Source:    "observability.quality_service",
		Subject:   score.ID.String(),
		ProjectID: getProjectIDFromContext(score.TraceID),
		UserID:    userID,
		Data: map[string]any{
			"score_id":        score.ID,
			"trace_id":        score.TraceID,
			"observation_id":  score.ObservationID,
			"score_name":      score.ScoreName,
			"score_value":     score.ScoreValue,
			"string_value":    score.StringValue,
			"data_type":       score.DataType,
			"source":          score.Source,
			"evaluator_name":  score.EvaluatorName,
		},
		Metadata: map[string]any{
			"evaluator_version": score.EvaluatorVersion,
			"comment":           score.Comment,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			TraceID:       &score.TraceID,
			ObservationID: score.ObservationID,
			UserID:        score.AuthorUserID,
		},
	}
}

// NewBatchIngestionStartedEvent creates a new batch ingestion started event
func NewBatchIngestionStartedEvent(projectID ulid.ULID, batchSize int, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeBatchIngestionStarted,
		Source:    "observability.ingestion_service",
		Subject:   "batch_ingestion",
		ProjectID: projectID,
		UserID:    userID,
		Data: map[string]any{
			"batch_size": batchSize,
			"started_at": time.Now(),
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			UserID: userID,
		},
	}
}

// NewBatchIngestionCompletedEvent creates a new batch ingestion completed event
func NewBatchIngestionCompletedEvent(projectID ulid.ULID, result *BatchIngestResult, userID *ulid.ULID) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeBatchIngestionCompleted,
		Source:    "observability.ingestion_service",
		Subject:   "batch_ingestion",
		ProjectID: projectID,
		UserID:    userID,
		Data: map[string]any{
			"processed_count": result.ProcessedCount,
			"failed_count":    result.FailedCount,
			"duration":        result.Duration,
			"errors_count":    len(result.Errors),
		},
		Metadata: map[string]any{
			"job_id": result.JobID,
		},
		Timestamp: time.Now(),
		Version:   "1.0",
		Correlation: &CorrelationInfo{
			UserID: userID,
		},
	}
}

// NewAlertTriggeredEvent creates a new alert triggered event
func NewAlertTriggeredEvent(projectID ulid.ULID, alertType, alertName, message string, severity string, metadata map[string]any) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeAlertTriggered,
		Source:    "observability.alert_service",
		Subject:   alertName,
		ProjectID: projectID,
		Data: map[string]any{
			"alert_type": alertType,
			"alert_name": alertName,
			"message":    message,
			"severity":   severity,
		},
		Metadata:  metadata,
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// NewThresholdExceededEvent creates a new threshold exceeded event
func NewThresholdExceededEvent(projectID ulid.ULID, metric, threshold string, currentValue, thresholdValue float64, metadata map[string]any) *Event {
	return &Event{
		ID:        ulid.New(),
		Type:      EventTypeThresholdExceeded,
		Source:    "observability.monitoring_service",
		Subject:   metric,
		ProjectID: projectID,
		Data: map[string]any{
			"metric":          metric,
			"threshold":       threshold,
			"current_value":   currentValue,
			"threshold_value": thresholdValue,
			"exceeded_by":     currentValue - thresholdValue,
			"exceeded_by_pct": ((currentValue - thresholdValue) / thresholdValue) * 100,
		},
		Metadata:  metadata,
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// Utility functions

// getProjectIDFromContext is a placeholder function that would resolve project ID from trace ID
// In a real implementation, this would query the database or use a cache
func getProjectIDFromContext(traceID ulid.ULID) ulid.ULID {
	// This would be implemented to resolve the project ID from the trace ID
	// For now, return a zero ULID as a placeholder
	return ulid.ULID{}
}

// EventMetrics represents metrics for events
type EventMetrics struct {
	TotalEvents       int64              `json:"total_events"`
	EventsByType      map[EventType]int64 `json:"events_by_type"`
	EventsBySource    map[string]int64   `json:"events_by_source"`
	EventsPerMinute   float64            `json:"events_per_minute"`
	EventsPerHour     float64            `json:"events_per_hour"`
	LastEventTime     time.Time          `json:"last_event_time"`
}

// EventSubscription represents an event subscription
type EventSubscription struct {
	ID           ulid.ULID   `json:"id"`
	HandlerName  string      `json:"handler_name"`
	EventTypes   []EventType `json:"event_types"`
	ProjectID    *ulid.ULID  `json:"project_id,omitempty"`
	IsActive     bool        `json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	LastTriggered *time.Time `json:"last_triggered,omitempty"`
}

// EventProcessingStats represents statistics for event processing
type EventProcessingStats struct {
	Processed       int64         `json:"processed"`
	Failed          int64         `json:"failed"`
	AverageLatency  time.Duration `json:"average_latency"`
	BacklogSize     int64         `json:"backlog_size"`
	ProcessingRate  float64       `json:"processing_rate"`
	LastProcessedAt time.Time     `json:"last_processed_at"`
}