package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
)

// AnalyticsRepository handles analytics data operations in ClickHouse
type AnalyticsRepository struct {
	db *database.ClickHouseDB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *database.ClickHouseDB) *AnalyticsRepository {
	return &AnalyticsRepository{
		db: db,
	}
}

// RequestLog represents an AI request log entry
type RequestLog struct {
	ID             string    `ch:"id"`
	UserID         string    `ch:"user_id"`
	OrganizationID string    `ch:"organization_id"`
	ProjectID      string    `ch:"project_id"`
	Environment    string    `ch:"environment"`
	Provider       string    `ch:"provider"`
	Model          string    `ch:"model"`
	Method         string    `ch:"method"`
	Endpoint       string    `ch:"endpoint"`
	StatusCode     int       `ch:"status_code"`
	InputTokens    int64     `ch:"input_tokens"`
	OutputTokens   int64     `ch:"output_tokens"`
	TotalTokens    int64     `ch:"total_tokens"`
	Cost           float64   `ch:"cost"`
	Latency        int64     `ch:"latency"`
	QualityScore   float64   `ch:"quality_score"`
	Cached         bool      `ch:"cached"`
	Error          string    `ch:"error"`
	Timestamp      time.Time `ch:"timestamp"`
}

// MetricPoint represents a time-series metric point
type MetricPoint struct {
	Timestamp      time.Time `ch:"timestamp"`
	OrganizationID string    `ch:"organization_id"`
	ProjectID      string    `ch:"project_id"`
	Environment    string    `ch:"environment"`
	MetricName     string    `ch:"metric_name"`
	MetricValue    float64   `ch:"metric_value"`
	Tags           string    `ch:"tags"`
}

// ProviderHealth represents provider health metrics
type ProviderHealth struct {
	Provider       string    `ch:"provider"`
	Model          string    `ch:"model"`
	SuccessRate    float64   `ch:"success_rate"`
	AverageLatency float64   `ch:"average_latency"`
	ErrorRate      float64   `ch:"error_rate"`
	Timestamp      time.Time `ch:"timestamp"`
}

// TelemetryEvent represents a telemetry event in ClickHouse
type TelemetryEvent struct {
	ID          string    `ch:"id"`
	BatchID     string    `ch:"batch_id"`
	ProjectID   string    `ch:"project_id"`
	Environment string    `ch:"environment"`
	EventType   string    `ch:"event_type"`
	EventData   string    `ch:"event_data"`
	Timestamp   time.Time `ch:"timestamp"`
	RetryCount  int       `ch:"retry_count"`
	ProcessedAt time.Time `ch:"processed_at"`
}

// TelemetryBatch represents a telemetry batch in ClickHouse
type TelemetryBatch struct {
	ID               string    `ch:"id"`
	ProjectID        string    `ch:"project_id"`
	Environment      string    `ch:"environment"`
	Status           string    `ch:"status"`
	TotalEvents      int       `ch:"total_events"`
	ProcessedEvents  int       `ch:"processed_events"`
	FailedEvents     int       `ch:"failed_events"`
	ProcessingTimeMs int       `ch:"processing_time_ms"`
	Metadata         string    `ch:"metadata"`
	Timestamp        time.Time `ch:"timestamp"`
	ProcessedAt      time.Time `ch:"processed_at"`
}

// TelemetryMetric represents a telemetry metric in ClickHouse
type TelemetryMetric struct {
	ProjectID    string    `ch:"project_id"`
	Environment  string    `ch:"environment"`
	MetricName   string    `ch:"metric_name"`
	MetricType   string    `ch:"metric_type"`
	MetricValue  float64   `ch:"metric_value"`
	Labels       string    `ch:"labels"`
	Metadata     string    `ch:"metadata"`
	Timestamp    time.Time `ch:"timestamp"`
	ProcessedAt  time.Time `ch:"processed_at"`
}

// InsertRequestLog inserts a new request log
func (r *AnalyticsRepository) InsertRequestLog(ctx context.Context, log *RequestLog) error {
	query := `
		INSERT INTO request_logs (
			id, user_id, organization_id, project_id, environment,
			provider, model, method, endpoint, status_code,
			input_tokens, output_tokens, total_tokens, cost, latency,
			quality_score, cached, error, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	return r.db.Execute(ctx, query,
		log.ID, log.UserID, log.OrganizationID, log.ProjectID, log.Environment,
		log.Provider, log.Model, log.Method, log.Endpoint, log.StatusCode,
		log.InputTokens, log.OutputTokens, log.TotalTokens, log.Cost, log.Latency,
		log.QualityScore, log.Cached, log.Error, log.Timestamp,
	)
}

// InsertBatchRequestLogs inserts multiple request logs efficiently
func (r *AnalyticsRepository) InsertBatchRequestLogs(ctx context.Context, logs []*RequestLog) error {
	if len(logs) == 0 {
		return nil
	}

	batch := make([][]interface{}, len(logs))
	for i, log := range logs {
		batch[i] = []interface{}{
			log.ID, log.UserID, log.OrganizationID, log.ProjectID, log.Environment,
			log.Provider, log.Model, log.Method, log.Endpoint, log.StatusCode,
			log.InputTokens, log.OutputTokens, log.TotalTokens, log.Cost, log.Latency,
			log.QualityScore, log.Cached, log.Error, log.Timestamp,
		}
	}

	query := `
		INSERT INTO request_logs (
			id, user_id, organization_id, project_id, environment,
			provider, model, method, endpoint, status_code,
			input_tokens, output_tokens, total_tokens, cost, latency,
			quality_score, cached, error, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range batch {
		if err := r.db.Execute(ctx, query, row...); err != nil {
			return fmt.Errorf("failed to insert request log batch: %w", err)
		}
	}

	return nil
}

// GetRequestStats retrieves request statistics for a time period
func (r *AnalyticsRepository) GetRequestStats(ctx context.Context, filter AnalyticsFilter) (*RequestStats, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			COUNT() as total_requests,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as successful_requests,
			AVG(latency) as avg_latency,
			SUM(cost) as total_cost,
			SUM(total_tokens) as total_tokens,
			AVG(quality_score) as avg_quality_score,
			SUM(CASE WHEN cached = true THEN 1 ELSE 0 END) as cache_hits
		FROM request_logs 
		%s`, whereClause)

	row := r.db.QueryRow(ctx, query, args...)
	
	var stats RequestStats
	if err := row.Scan(
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.AverageLatency,
		&stats.TotalCost,
		&stats.TotalTokens,
		&stats.AverageQualityScore,
		&stats.CacheHits,
	); err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}

	// Calculate derived metrics
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalRequests) * 100
	}

	return &stats, nil
}

// GetProviderStats retrieves provider-wise statistics
func (r *AnalyticsRepository) GetProviderStats(ctx context.Context, filter AnalyticsFilter) ([]*ProviderStats, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			provider,
			COUNT() as total_requests,
			AVG(latency) as avg_latency,
			SUM(cost) as total_cost,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as successful_requests,
			AVG(quality_score) as avg_quality_score
		FROM request_logs 
		%s
		GROUP BY provider
		ORDER BY total_requests DESC`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()

	var stats []*ProviderStats
	for rows.Next() {
		var stat ProviderStats
		if err := rows.Scan(
			&stat.Provider,
			&stat.TotalRequests,
			&stat.AverageLatency,
			&stat.TotalCost,
			&stat.SuccessfulRequests,
			&stat.AverageQualityScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan provider stats: %w", err)
		}

		// Calculate success rate
		if stat.TotalRequests > 0 {
			stat.SuccessRate = float64(stat.SuccessfulRequests) / float64(stat.TotalRequests) * 100
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetTimeSeriesData retrieves time-series data for charting
func (r *AnalyticsRepository) GetTimeSeriesData(ctx context.Context, filter AnalyticsFilter, interval string, metric string) ([]*TimeSeriesPoint, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	// Convert interval to ClickHouse interval function
	var intervalFunc string
	switch interval {
	case "hour":
		intervalFunc = "toStartOfHour(timestamp)"
	case "day":
		intervalFunc = "toDate(timestamp)"
	case "week":
		intervalFunc = "toMonday(timestamp)"
	case "month":
		intervalFunc = "toStartOfMonth(timestamp)"
	default:
		intervalFunc = "toStartOfHour(timestamp)"
	}

	var aggregateFunc string
	switch metric {
	case "requests":
		aggregateFunc = "COUNT()"
	case "cost":
		aggregateFunc = "SUM(cost)"
	case "latency":
		aggregateFunc = "AVG(latency)"
	case "tokens":
		aggregateFunc = "SUM(total_tokens)"
	case "quality_score":
		aggregateFunc = "AVG(quality_score)"
	default:
		aggregateFunc = "COUNT()"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as time_bucket,
			%s as value
		FROM request_logs 
		%s
		GROUP BY time_bucket
		ORDER BY time_bucket`, intervalFunc, aggregateFunc, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series data: %w", err)
	}
	defer rows.Close()

	var points []*TimeSeriesPoint
	for rows.Next() {
		var point TimeSeriesPoint
		if err := rows.Scan(&point.Timestamp, &point.Value); err != nil {
			return nil, fmt.Errorf("failed to scan time series point: %w", err)
		}
		points = append(points, &point)
	}

	return points, nil
}

// GetCostBreakdown retrieves cost breakdown by different dimensions
func (r *AnalyticsRepository) GetCostBreakdown(ctx context.Context, filter AnalyticsFilter, groupBy string) ([]*CostBreakdown, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			%s as dimension,
			SUM(cost) as total_cost,
			COUNT() as request_count,
			AVG(cost) as avg_cost_per_request
		FROM request_logs 
		%s
		GROUP BY %s
		ORDER BY total_cost DESC
		LIMIT 50`, groupBy, whereClause, groupBy)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []*CostBreakdown
	for rows.Next() {
		var item CostBreakdown
		if err := rows.Scan(
			&item.Dimension,
			&item.TotalCost,
			&item.RequestCount,
			&item.AverageCostPerRequest,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost breakdown: %w", err)
		}
		breakdown = append(breakdown, &item)
	}

	return breakdown, nil
}

// InsertMetric inserts a metric point
func (r *AnalyticsRepository) InsertMetric(ctx context.Context, metric *MetricPoint) error {
	query := `
		INSERT INTO metrics (
			timestamp, organization_id, project_id, environment,
			metric_name, metric_value, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.db.Execute(ctx, query,
		metric.Timestamp, metric.OrganizationID, metric.ProjectID, metric.Environment,
		metric.MetricName, metric.MetricValue, metric.Tags,
	)
}

// buildWhereClause builds WHERE clause from filter
func (r *AnalyticsRepository) buildWhereClause(filter AnalyticsFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "timestamp >= ? AND timestamp <= ?")
	args = append(args, filter.StartTime, filter.EndTime)

	if filter.OrganizationID != "" {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, filter.OrganizationID)
	}

	if filter.ProjectID != "" {
		conditions = append(conditions, "project_id = ?")
		args = append(args, filter.ProjectID)
	}

	if filter.Environment != "" {
		conditions = append(conditions, "environment = ?")
		args = append(args, filter.Environment)
	}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, filter.Provider)
	}

	if filter.Model != "" {
		conditions = append(conditions, "model = ?")
		args = append(args, filter.Model)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0])) + 
		func() string {
			if len(conditions) > 1 {
				return " AND " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[1:]))
			}
			return ""
		}(), args
}

// Data structures for analytics results

type AnalyticsFilter struct {
	StartTime      time.Time
	EndTime        time.Time
	OrganizationID string
	ProjectID      string
	Environment    string
	UserID         string
	Provider       string
	Model          string
}

type RequestStats struct {
	TotalRequests        int64   `json:"total_requests"`
	SuccessfulRequests   int64   `json:"successful_requests"`
	SuccessRate          float64 `json:"success_rate"`
	AverageLatency       float64 `json:"average_latency"`
	TotalCost            float64 `json:"total_cost"`
	TotalTokens          int64   `json:"total_tokens"`
	AverageQualityScore  float64 `json:"average_quality_score"`
	CacheHits            int64   `json:"cache_hits"`
	CacheHitRate         float64 `json:"cache_hit_rate"`
}

type ProviderStats struct {
	Provider             string  `json:"provider"`
	TotalRequests        int64   `json:"total_requests"`
	SuccessfulRequests   int64   `json:"successful_requests"`
	SuccessRate          float64 `json:"success_rate"`
	AverageLatency       float64 `json:"average_latency"`
	TotalCost            float64 `json:"total_cost"`
	AverageQualityScore  float64 `json:"average_quality_score"`
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type CostBreakdown struct {
	Dimension             string  `json:"dimension"`
	TotalCost             float64 `json:"total_cost"`
	RequestCount          int64   `json:"request_count"`
	AverageCostPerRequest float64 `json:"average_cost_per_request"`
}

// Telemetry-specific insert methods

// InsertTelemetryEvent inserts a single telemetry event (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryEvent(ctx context.Context, event *observability.TelemetryEvent) error {
	chEvent, err := domainEventToClickHouse(event)
	if err != nil {
		return fmt.Errorf("failed to convert domain event to ClickHouse: %w", err)
	}

	return r.insertClickHouseEvent(ctx, chEvent)
}

// insertClickHouseEvent inserts a ClickHouse telemetry event (internal method)
func (r *AnalyticsRepository) insertClickHouseEvent(ctx context.Context, event *TelemetryEvent) error {
	query := `
		INSERT INTO telemetry_events (
			id, batch_id, project_id, environment, event_type,
			event_data, timestamp, retry_count, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Execute(ctx, query,
		event.ID, event.BatchID, event.ProjectID, event.Environment, event.EventType,
		event.EventData, event.Timestamp, event.RetryCount, event.ProcessedAt,
	)
}

// InsertTelemetryEventsBatch inserts multiple telemetry events efficiently (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryEventsBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Convert domain events to ClickHouse events (context carried in domain structs)
	chEvents := make([]*TelemetryEvent, len(events))
	for i, event := range events {
		chEvent, err := domainEventToClickHouse(event)
		if err != nil {
			return fmt.Errorf("failed to convert domain event %d to ClickHouse: %w", i, err)
		}
		chEvents[i] = chEvent
	}

	return r.insertTelemetryEventsBatchClickHouse(ctx, chEvents)
}

// insertTelemetryEventsBatchClickHouse inserts multiple ClickHouse telemetry events (internal method)
func (r *AnalyticsRepository) insertTelemetryEventsBatchClickHouse(ctx context.Context, events []*TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	batchRows := make([][]interface{}, len(events))
	for i, event := range events {
		batchRows[i] = []interface{}{
			event.ID, event.BatchID, event.ProjectID, event.Environment, event.EventType,
			event.EventData, event.Timestamp, event.RetryCount, event.ProcessedAt,
		}
	}

	query := `
		INSERT INTO telemetry_events (
			id, batch_id, project_id, environment, event_type,
			event_data, timestamp, retry_count, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range batchRows {
		if err := r.db.Execute(ctx, query, row...); err != nil {
			return fmt.Errorf("failed to insert telemetry event batch: %w", err)
		}
	}

	return nil
}

// InsertTelemetryBatch inserts a single telemetry batch record (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryBatch(ctx context.Context, batch *observability.TelemetryBatch) error {
	chBatch, err := domainBatchToClickHouse(batch)
	if err != nil {
		return fmt.Errorf("failed to convert domain batch to ClickHouse: %w", err)
	}

	return r.insertTelemetryBatchClickHouse(ctx, chBatch)
}

// insertTelemetryBatchClickHouse inserts a ClickHouse telemetry batch record (internal method)
func (r *AnalyticsRepository) insertTelemetryBatchClickHouse(ctx context.Context, batch *TelemetryBatch) error {
	query := `
		INSERT INTO telemetry_batches (
			id, project_id, environment, status, total_events,
			processed_events, failed_events, processing_time_ms,
			metadata, timestamp, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Execute(ctx, query,
		batch.ID, batch.ProjectID, batch.Environment, batch.Status, batch.TotalEvents,
		batch.ProcessedEvents, batch.FailedEvents, batch.ProcessingTimeMs,
		batch.Metadata, batch.Timestamp, batch.ProcessedAt,
	)
}

// InsertTelemetryBatchesBatch inserts multiple telemetry batch records efficiently (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryBatchesBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	if len(batches) == 0 {
		return nil
	}

	// Convert domain batches to ClickHouse batches (context carried in domain structs)
	chBatches := make([]*TelemetryBatch, len(batches))
	for i, batch := range batches {
		chBatch, err := domainBatchToClickHouse(batch)
		if err != nil {
			return fmt.Errorf("failed to convert domain batch %d to ClickHouse: %w", i, err)
		}
		chBatches[i] = chBatch
	}

	return r.insertTelemetryBatchesBatchClickHouse(ctx, chBatches)
}

// insertTelemetryBatchesBatchClickHouse inserts multiple ClickHouse telemetry batch records (internal method)
func (r *AnalyticsRepository) insertTelemetryBatchesBatchClickHouse(ctx context.Context, batches []*TelemetryBatch) error {
	if len(batches) == 0 {
		return nil
	}

	batchRows := make([][]interface{}, len(batches))
	for i, batch := range batches {
		batchRows[i] = []interface{}{
			batch.ID, batch.ProjectID, batch.Environment, batch.Status, batch.TotalEvents,
			batch.ProcessedEvents, batch.FailedEvents, batch.ProcessingTimeMs,
			batch.Metadata, batch.Timestamp, batch.ProcessedAt,
		}
	}

	query := `
		INSERT INTO telemetry_batches (
			id, project_id, environment, status, total_events,
			processed_events, failed_events, processing_time_ms,
			metadata, timestamp, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range batchRows {
		if err := r.db.Execute(ctx, query, row...); err != nil {
			return fmt.Errorf("failed to insert telemetry batch record: %w", err)
		}
	}

	return nil
}

// InsertTelemetryMetric inserts a single telemetry metric (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryMetric(ctx context.Context, metric *observability.TelemetryMetric) error {
	chMetric, err := domainMetricToClickHouse(metric)
	if err != nil {
		return fmt.Errorf("failed to convert domain metric to ClickHouse: %w", err)
	}

	return r.insertTelemetryMetricClickHouse(ctx, chMetric)
}

// insertTelemetryMetricClickHouse inserts a ClickHouse telemetry metric (internal method)
func (r *AnalyticsRepository) insertTelemetryMetricClickHouse(ctx context.Context, metric *TelemetryMetric) error {
	query := `
		INSERT INTO telemetry_metrics (
			project_id, environment, metric_name, metric_type, metric_value,
			labels, metadata, timestamp, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Execute(ctx, query,
		metric.ProjectID, metric.Environment, metric.MetricName, metric.MetricType, metric.MetricValue,
		metric.Labels, metric.Metadata, metric.Timestamp, metric.ProcessedAt,
	)
}

// InsertTelemetryMetricsBatch inserts multiple telemetry metrics efficiently (domain interface implementation)
func (r *AnalyticsRepository) InsertTelemetryMetricsBatch(ctx context.Context, metrics []*observability.TelemetryMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Convert domain metrics to ClickHouse metrics
	chMetrics := make([]*TelemetryMetric, len(metrics))
	for i, metric := range metrics {
		chMetric, err := domainMetricToClickHouse(metric)
		if err != nil {
			return fmt.Errorf("failed to convert domain metric %d to ClickHouse: %w", i, err)
		}
		chMetrics[i] = chMetric
	}

	return r.insertTelemetryMetricsBatchClickHouse(ctx, chMetrics)
}

// insertTelemetryMetricsBatchClickHouse inserts multiple ClickHouse telemetry metrics (internal method)
func (r *AnalyticsRepository) insertTelemetryMetricsBatchClickHouse(ctx context.Context, metrics []*TelemetryMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	batchRows := make([][]interface{}, len(metrics))
	for i, metric := range metrics {
		batchRows[i] = []interface{}{
			metric.ProjectID, metric.Environment, metric.MetricName, metric.MetricType, metric.MetricValue,
			metric.Labels, metric.Metadata, metric.Timestamp, metric.ProcessedAt,
		}
	}

	query := `
		INSERT INTO telemetry_metrics (
			project_id, environment, metric_name, metric_type, metric_value,
			labels, metadata, timestamp, processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range batchRows {
		if err := r.db.Execute(ctx, query, row...); err != nil {
			return fmt.Errorf("failed to insert telemetry metric batch: %w", err)
		}
	}

	return nil
}

// Domain to ClickHouse conversion layer for clean architecture
// These methods convert rich domain types to ClickHouse-optimized DTOs

// domainEventToClickHouse converts a domain TelemetryEvent to ClickHouse TelemetryEvent
func domainEventToClickHouse(event *observability.TelemetryEvent) (*TelemetryEvent, error) {
	eventDataJSON, err := json.Marshal(event.EventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	processedAt := time.Time{}
	if event.ProcessedAt != nil {
		processedAt = *event.ProcessedAt
	}

	return &TelemetryEvent{
		ID:          event.ID.String(),
		BatchID:     event.BatchID.String(),
		ProjectID:   event.ProjectID.String(),   // Read from domain struct
		Environment: event.Environment,           // Read from domain struct
		EventType:   string(event.EventType),
		EventData:   string(eventDataJSON),
		Timestamp:   event.CreatedAt,
		RetryCount:  event.RetryCount,
		ProcessedAt: processedAt,
	}, nil
}

// domainBatchToClickHouse converts a domain TelemetryBatch to ClickHouse TelemetryBatch
func domainBatchToClickHouse(batch *observability.TelemetryBatch) (*TelemetryBatch, error) {
	metadataJSON, err := json.Marshal(batch.BatchMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch metadata: %w", err)
	}

	processingTimeMs := 0
	if batch.ProcessingTimeMs != nil {
		processingTimeMs = *batch.ProcessingTimeMs
	}

	processedAt := time.Time{}
	if batch.CompletedAt != nil {
		processedAt = *batch.CompletedAt
	}

	return &TelemetryBatch{
		ID:               batch.ID.String(),
		ProjectID:        batch.ProjectID.String(),
		Environment:      batch.Environment,  // Read from domain struct
		Status:           string(batch.Status),
		TotalEvents:      batch.TotalEvents,
		ProcessedEvents:  batch.ProcessedEvents,
		FailedEvents:     batch.FailedEvents,
		ProcessingTimeMs: processingTimeMs,
		Metadata:         string(metadataJSON),
		Timestamp:        batch.CreatedAt,
		ProcessedAt:      processedAt,
	}, nil
}

// domainMetricToClickHouse converts a domain TelemetryMetric to ClickHouse TelemetryMetric
func domainMetricToClickHouse(metric *observability.TelemetryMetric) (*TelemetryMetric, error) {
	labelsJSON, err := json.Marshal(metric.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric labels: %w", err)
	}

	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric metadata: %w", err)
	}

	processedAt := time.Time{}
	if metric.ProcessedAt != nil {
		processedAt = *metric.ProcessedAt
	}

	return &TelemetryMetric{
		ProjectID:   metric.ProjectID.String(),
		Environment: metric.Environment,
		MetricName:  metric.MetricName,
		MetricType:  metric.MetricType,
		MetricValue: metric.MetricValue,
		Labels:      string(labelsJSON),
		Metadata:    string(metadataJSON),
		Timestamp:   metric.Timestamp,
		ProcessedAt: processedAt,
	}, nil
}