package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// TelemetryBatchRepository implements the observability.TelemetryBatchRepository interface
type TelemetryBatchRepository struct {
	db *gorm.DB
}

// NewTelemetryBatchRepository creates a new telemetry batch repository instance
func NewTelemetryBatchRepository(db *gorm.DB) *TelemetryBatchRepository {
	return &TelemetryBatchRepository{
		db: db,
	}
}

// Create creates a new telemetry batch in the database
func (r *TelemetryBatchRepository) Create(ctx context.Context, batch *observability.TelemetryBatch) error {
	if batch.ID.IsZero() {
		batch.ID = ulid.New()
	}

	// Convert metadata to JSON for storage
	metadataJSON, err := json.Marshal(batch.BatchMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal batch metadata: %w", err)
	}

	// Prepare SQL statement
	query := `
		INSERT INTO telemetry_batches (
			id, project_id, batch_metadata, total_events, processed_events,
			failed_events, status, processing_time_ms, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	batch.CreatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		batch.ID,
		batch.ProjectID,
		string(metadataJSON),
		batch.TotalEvents,
		batch.ProcessedEvents,
		batch.FailedEvents,
		batch.Status,
		batch.ProcessingTimeMs,
		batch.CreatedAt,
		batch.CompletedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("telemetry batch with ID %s already exists: %w", batch.ID.String(), err)
		}
		return fmt.Errorf("failed to create telemetry batch: %w", err)
	}

	return nil
}

// GetByID retrieves a telemetry batch by its ID
func (r *TelemetryBatchRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	var batch observability.TelemetryBatch
	var metadataJSON string

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&batch.ID,
		&batch.ProjectID,
		&metadataJSON,
		&batch.TotalEvents,
		&batch.ProcessedEvents,
		&batch.FailedEvents,
		&batch.Status,
		&batch.ProcessingTimeMs,
		&batch.CreatedAt,
		&batch.CompletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("telemetry batch with ID %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get telemetry batch by ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(metadataJSON), &batch.BatchMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch metadata: %w", err)
	}

	return &batch, nil
}

// Update updates an existing telemetry batch
func (r *TelemetryBatchRepository) Update(ctx context.Context, batch *observability.TelemetryBatch) error {
	// Convert metadata to JSON for storage
	metadataJSON, err := json.Marshal(batch.BatchMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal batch metadata: %w", err)
	}

	query := `
		UPDATE telemetry_batches
		SET batch_metadata = $2, total_events = $3, processed_events = $4,
			failed_events = $5, status = $6, processing_time_ms = $7, completed_at = $8
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query,
		batch.ID,
		string(metadataJSON),
		batch.TotalEvents,
		batch.ProcessedEvents,
		batch.FailedEvents,
		batch.Status,
		batch.ProcessingTimeMs,
		batch.CompletedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update telemetry batch: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry batch with ID %s not found", batch.ID.String())
	}

	return nil
}

// Delete deletes a telemetry batch by its ID
func (r *TelemetryBatchRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM telemetry_batches WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete telemetry batch: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry batch with ID %s not found", id.String())
	}

	return nil
}

// GetByProjectID retrieves telemetry batches by project ID with pagination
func (r *TelemetryBatchRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry batches by project ID: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// GetActiveByProjectID retrieves active telemetry batches by project ID
func (r *TelemetryBatchRepository) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*observability.TelemetryBatch, error) {
	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE project_id = $1 AND status = 'processing'
		ORDER BY created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query active telemetry batches: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// GetByStatus retrieves telemetry batches by status with pagination
func (r *TelemetryBatchRepository) GetByStatus(ctx context.Context, status observability.BatchStatus, limit, offset int) ([]*observability.TelemetryBatch, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, status, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry batches by status: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// GetProcessingBatches retrieves all batches currently being processed
func (r *TelemetryBatchRepository) GetProcessingBatches(ctx context.Context) ([]*observability.TelemetryBatch, error) {
	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE status = 'processing'
		ORDER BY created_at ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query processing telemetry batches: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// GetCompletedBatches retrieves completed batches by project ID with pagination
func (r *TelemetryBatchRepository) GetCompletedBatches(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryBatch, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE project_id = $1 AND status IN ('completed', 'failed', 'partial')
		ORDER BY completed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query completed telemetry batches: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// SearchBatches searches telemetry batches with filters and returns results with total count
func (r *TelemetryBatchRepository) SearchBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) ([]*observability.TelemetryBatch, int, error) {
	// Build WHERE clause and arguments
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM telemetry_batches" + whereClause
	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count telemetry batches: %w", err)
	}

	// Get batches
	orderClause := r.buildOrderClause(filter)
	limitClause := r.buildLimitClause(filter)

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches` + whereClause + orderClause + limitClause

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search telemetry batches: %w", err)
	}
	defer rows.Close()

	batches, err := r.scanTelemetryBatches(rows)
	if err != nil {
		return nil, 0, err
	}

	return batches, int(totalCount), nil
}

// GetBatchWithEvents retrieves a telemetry batch with all its events
func (r *TelemetryBatchRepository) GetBatchWithEvents(ctx context.Context, id ulid.ULID) (*observability.TelemetryBatch, error) {
	// Get the batch first
	batch, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get events for the batch
	eventQuery := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE batch_id = $1
		ORDER BY created_at ASC
	`

	eventRows, err := r.db.WithContext(ctx).Raw(eventQuery, id).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry events: %w", err)
	}
	defer eventRows.Close()

	var events []observability.TelemetryEvent
	for eventRows.Next() {
		var event observability.TelemetryEvent
		var eventPayloadJSON string

		err := eventRows.Scan(
			&event.ID,
			&event.BatchID,
			&event.EventType,
			&eventPayloadJSON,
			&event.ProcessedAt,
			&event.ErrorMessage,
			&event.RetryCount,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry event: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(eventPayloadJSON), &event.EventPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
		}

		events = append(events, event)
	}

	batch.Events = events
	return batch, nil
}

// GetBatchStats retrieves aggregated statistics for a telemetry batch
func (r *TelemetryBatchRepository) GetBatchStats(ctx context.Context, id ulid.ULID) (*observability.BatchStats, error) {
	// First get basic batch info
	batch, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Calculate success rate
	successRate := batch.CalculateSuccessRate()

	// Calculate throughput
	var throughputPerSec float64
	if batch.ProcessingTimeMs != nil && *batch.ProcessingTimeMs > 0 {
		throughputPerSec = float64(batch.ProcessedEvents) / (float64(*batch.ProcessingTimeMs) / 1000.0)
	}

	// Get event type distribution
	eventTypeQuery := `
		SELECT event_type, COUNT(*) as count
		FROM telemetry_events
		WHERE batch_id = $1
		GROUP BY event_type
	`

	eventTypeRows, err := r.db.WithContext(ctx).Raw(eventTypeQuery, id).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query event type distribution: %w", err)
	}
	defer eventTypeRows.Close()

	eventTypeDistribution := make(map[observability.TelemetryEventType]int)
	for eventTypeRows.Next() {
		var eventType observability.TelemetryEventType
		var count int

		err := eventTypeRows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event type distribution: %w", err)
		}

		eventTypeDistribution[eventType] = count
	}

	// Get retry distribution
	retryQuery := `
		SELECT retry_count, COUNT(*) as count
		FROM telemetry_events
		WHERE batch_id = $1
		GROUP BY retry_count
	`

	retryRows, err := r.db.WithContext(ctx).Raw(retryQuery, id).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query retry distribution: %w", err)
	}
	defer retryRows.Close()

	retryDistribution := make(map[int]int)
	for retryRows.Next() {
		var retryCount int
		var count int

		err := retryRows.Scan(&retryCount, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retry distribution: %w", err)
		}

		retryDistribution[retryCount] = count
	}

	stats := &observability.BatchStats{
		BatchID:               id,
		TotalEvents:           batch.TotalEvents,
		ProcessedEvents:       batch.ProcessedEvents,
		FailedEvents:          batch.FailedEvents,
		SuccessRate:           successRate,
		ProcessingTimeMs:      batch.ProcessingTimeMs,
		ThroughputPerSec:      throughputPerSec,
		EventTypeDistribution: eventTypeDistribution,
		RetryDistribution:     retryDistribution,
	}

	return stats, nil
}

// CreateBatch creates multiple telemetry batches in a single transaction
func (r *TelemetryBatchRepository) CreateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	if len(batches) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, batch := range batches {
			if err := r.createWithTx(ctx, tx, batch); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple telemetry batches in a single transaction
func (r *TelemetryBatchRepository) UpdateBatch(ctx context.Context, batches []*observability.TelemetryBatch) error {
	if len(batches) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, batch := range batches {
			if err := r.updateWithTx(ctx, tx, batch); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatchStatus updates the status and processing time of a telemetry batch
func (r *TelemetryBatchRepository) UpdateBatchStatus(ctx context.Context, id ulid.ULID, status observability.BatchStatus, processingTimeMs *int) error {
	var completedAt *time.Time
	if status == observability.BatchStatusCompleted || status == observability.BatchStatusFailed || status == observability.BatchStatusPartial {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE telemetry_batches
		SET status = $2, processing_time_ms = $3, completed_at = $4
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, id, status, processingTimeMs, completedAt)

	if result.Error != nil {
		return fmt.Errorf("failed to update telemetry batch status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry batch with ID %s not found", id.String())
	}

	return nil
}

// GetBatchesByTimeRange retrieves telemetry batches within a time range
func (r *TelemetryBatchRepository) GetBatchesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.TelemetryBatch, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, startTime, endTime, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry batches by time range: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// CountBatches counts telemetry batches matching the filter criteria
func (r *TelemetryBatchRepository) CountBatches(ctx context.Context, filter *observability.TelemetryBatchFilter) (int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := "SELECT COUNT(*) FROM telemetry_batches" + whereClause
	var count int64

	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count telemetry batches: %w", err)
	}

	return count, nil
}

// GetRecentBatches retrieves the most recent telemetry batches for a project
func (r *TelemetryBatchRepository) GetRecentBatches(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.TelemetryBatch, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, project_id, batch_metadata, total_events, processed_events,
			   failed_events, status, processing_time_ms, created_at, completed_at
		FROM telemetry_batches
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query recent telemetry batches: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryBatches(rows)
}

// GetBatchThroughputStats retrieves throughput statistics for telemetry batches
func (r *TelemetryBatchRepository) GetBatchThroughputStats(ctx context.Context, projectID ulid.ULID, timeWindow time.Duration) (*observability.BatchThroughputStats, error) {
	startTime := time.Now().Add(-timeWindow)

	query := `
		SELECT
			COUNT(*) as batch_count,
			COALESCE(SUM(total_events), 0) as total_events,
			COALESCE(AVG(total_events), 0) as avg_events_per_batch,
			COALESCE(MAX(total_events), 0) as max_events_per_batch
		FROM telemetry_batches
		WHERE project_id = $1 AND created_at >= $2
	`

	var batchCount, totalEvents, maxEventsPerBatch int64
	var avgEventsPerBatch float64

	row := r.db.WithContext(ctx).Raw(query, projectID, startTime).Row()
	err := row.Scan(&batchCount, &totalEvents, &avgEventsPerBatch, &maxEventsPerBatch)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch throughput stats: %w", err)
	}

	// Calculate per-minute rates
	minutes := timeWindow.Minutes()
	batchesPerMinute := float64(batchCount) / minutes
	eventsPerMinute := float64(totalEvents) / minutes

	stats := &observability.BatchThroughputStats{
		BatchesPerMinute:      batchesPerMinute,
		EventsPerMinute:       eventsPerMinute,
		AverageEventsPerBatch: avgEventsPerBatch,
		PeakThroughput:        float64(maxEventsPerBatch),
		ThroughputTrend:       "stable", // TODO: Calculate trend based on historical data
		TimeWindow:            timeWindow,
		LastCalculated:        time.Now(),
	}

	return stats, nil
}

// GetBatchProcessingMetrics retrieves processing performance metrics for telemetry batches
func (r *TelemetryBatchRepository) GetBatchProcessingMetrics(ctx context.Context, filter *observability.TelemetryBatchFilter) (*observability.BatchProcessingMetrics, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := `
		SELECT
			COUNT(*) as total_batches,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_batches,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_batches,
			COUNT(CASE WHEN status = 'partial' THEN 1 END) as partial_batches,
			COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_batches,
			COALESCE(AVG(processing_time_ms), 0) as avg_processing_time,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY processing_time_ms), 0) as median_processing_time,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY processing_time_ms), 0) as p95_processing_time,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY processing_time_ms), 0) as p99_processing_time,
			COALESCE(AVG(total_events), 0) as avg_events_per_batch,
			COALESCE(SUM(total_events), 0) as total_events,
			COALESCE(SUM(processed_events), 0) as processed_events,
			COALESCE(SUM(failed_events), 0) as failed_events
		FROM telemetry_batches` + whereClause

	row := r.db.WithContext(ctx).Raw(query, args...).Row()

	var metrics observability.BatchProcessingMetrics
	err = row.Scan(
		&metrics.TotalBatches,
		&metrics.CompletedBatches,
		&metrics.FailedBatches,
		&metrics.PartialBatches,
		&metrics.ProcessingBatches,
		&metrics.AverageProcessingTime,
		&metrics.MedianProcessingTime,
		&metrics.P95ProcessingTime,
		&metrics.P99ProcessingTime,
		&metrics.AverageEventsPerBatch,
		&metrics.TotalEvents,
		&metrics.ProcessedEvents,
		&metrics.FailedEvents,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get batch processing metrics: %w", err)
	}

	// Calculate success rates
	if metrics.TotalBatches > 0 {
		metrics.SuccessRate = float64(metrics.CompletedBatches) / float64(metrics.TotalBatches) * 100.0
	}

	if metrics.TotalEvents > 0 {
		metrics.EventSuccessRate = float64(metrics.ProcessedEvents) / float64(metrics.TotalEvents) * 100.0
	}

	// Calculate deduplication rate (placeholder - would need actual dedup data)
	metrics.DeduplicationRate = 0.0

	return &metrics, nil
}

// Helper methods

// createWithTx creates a telemetry batch within a transaction
func (r *TelemetryBatchRepository) createWithTx(ctx context.Context, tx *gorm.DB, batch *observability.TelemetryBatch) error {
	if batch.ID.IsZero() {
		batch.ID = ulid.New()
	}

	metadataJSON, err := json.Marshal(batch.BatchMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal batch metadata: %w", err)
	}

	query := `
		INSERT INTO telemetry_batches (
			id, project_id, batch_metadata, total_events, processed_events,
			failed_events, status, processing_time_ms, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	batch.CreatedAt = now

	return tx.WithContext(ctx).Exec(query,
		batch.ID,
		batch.ProjectID,
		string(metadataJSON),
		batch.TotalEvents,
		batch.ProcessedEvents,
		batch.FailedEvents,
		batch.Status,
		batch.ProcessingTimeMs,
		batch.CreatedAt,
		batch.CompletedAt,
	).Error
}

// updateWithTx updates a telemetry batch within a transaction
func (r *TelemetryBatchRepository) updateWithTx(ctx context.Context, tx *gorm.DB, batch *observability.TelemetryBatch) error {
	metadataJSON, err := json.Marshal(batch.BatchMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal batch metadata: %w", err)
	}

	query := `
		UPDATE telemetry_batches
		SET batch_metadata = $2, total_events = $3, processed_events = $4,
			failed_events = $5, status = $6, processing_time_ms = $7, completed_at = $8
		WHERE id = $1
	`

	result := tx.WithContext(ctx).Exec(query,
		batch.ID,
		string(metadataJSON),
		batch.TotalEvents,
		batch.ProcessedEvents,
		batch.FailedEvents,
		batch.Status,
		batch.ProcessingTimeMs,
		batch.CompletedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update telemetry batch: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry batch with ID %s not found", batch.ID.String())
	}

	return nil
}

// scanTelemetryBatches scans multiple telemetry batches from SQL rows
func (r *TelemetryBatchRepository) scanTelemetryBatches(rows *sql.Rows) ([]*observability.TelemetryBatch, error) {
	var batches []*observability.TelemetryBatch

	for rows.Next() {
		var batch observability.TelemetryBatch
		var metadataJSON string

		err := rows.Scan(
			&batch.ID,
			&batch.ProjectID,
			&metadataJSON,
			&batch.TotalEvents,
			&batch.ProcessedEvents,
			&batch.FailedEvents,
			&batch.Status,
			&batch.ProcessingTimeMs,
			&batch.CreatedAt,
			&batch.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry batch: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(metadataJSON), &batch.BatchMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal batch metadata: %w", err)
		}

		batches = append(batches, &batch)
	}

	return batches, nil
}

// buildWhereClause builds a WHERE clause based on filter criteria
func (r *TelemetryBatchRepository) buildWhereClause(filter *observability.TelemetryBatchFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *filter.ProjectID)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if len(filter.Statuses) > 0 {
		statusPlaceholders := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	if filter.MinTotalEvents != nil {
		conditions = append(conditions, fmt.Sprintf("total_events >= $%d", argIndex))
		args = append(args, *filter.MinTotalEvents)
		argIndex++
	}

	if filter.MaxTotalEvents != nil {
		conditions = append(conditions, fmt.Sprintf("total_events <= $%d", argIndex))
		args = append(args, *filter.MaxTotalEvents)
		argIndex++
	}

	if filter.MinProcessedEvents != nil {
		conditions = append(conditions, fmt.Sprintf("processed_events >= $%d", argIndex))
		args = append(args, *filter.MinProcessedEvents)
		argIndex++
	}

	if filter.MinFailedEvents != nil {
		conditions = append(conditions, fmt.Sprintf("failed_events >= $%d", argIndex))
		args = append(args, *filter.MinFailedEvents)
		argIndex++
	}

	if filter.MinProcessingTime != nil {
		conditions = append(conditions, fmt.Sprintf("processing_time_ms >= $%d", argIndex))
		args = append(args, *filter.MinProcessingTime)
		argIndex++
	}

	if filter.MaxProcessingTime != nil {
		conditions = append(conditions, fmt.Sprintf("processing_time_ms <= $%d", argIndex))
		args = append(args, *filter.MaxProcessingTime)
		argIndex++
	}

	if len(filter.HasMetadata) > 0 {
		for key, value := range filter.HasMetadata {
			conditions = append(conditions, fmt.Sprintf("batch_metadata->>$%d = $%d", argIndex, argIndex+1))
			args = append(args, key, fmt.Sprintf("%v", value))
			argIndex += 2
		}
	}

	if len(conditions) == 0 {
		return "", args, nil
	}

	return " WHERE " + strings.Join(conditions, " AND "), args, nil
}

// buildOrderClause builds an ORDER BY clause based on filter criteria
func (r *TelemetryBatchRepository) buildOrderClause(filter *observability.TelemetryBatchFilter) string {
	if filter == nil || filter.SortBy == "" {
		return " ORDER BY created_at DESC"
	}

	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}

	switch filter.SortBy {
	case "created_at", "completed_at", "total_events", "processing_time_ms":
		return fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	default:
		return " ORDER BY created_at DESC"
	}
}

// buildLimitClause builds a LIMIT clause based on filter criteria
func (r *TelemetryBatchRepository) buildLimitClause(filter *observability.TelemetryBatchFilter) string {
	if filter == nil {
		return " LIMIT 50"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}

