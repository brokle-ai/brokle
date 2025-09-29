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

// TelemetryEventRepository implements the observability.TelemetryEventRepository interface
type TelemetryEventRepository struct {
	db *gorm.DB
}

// NewTelemetryEventRepository creates a new telemetry event repository instance
func NewTelemetryEventRepository(db *gorm.DB) *TelemetryEventRepository {
	return &TelemetryEventRepository{
		db: db,
	}
}

// Create creates a new telemetry event in the database
func (r *TelemetryEventRepository) Create(ctx context.Context, event *observability.TelemetryEvent) error {
	if event.ID.IsZero() {
		event.ID = ulid.New()
	}

	// Convert event payload to JSON for storage
	payloadJSON, err := json.Marshal(event.EventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Prepare SQL statement
	query := `
		INSERT INTO telemetry_events (
			id, batch_id, event_type, event_payload, processed_at,
			error_message, retry_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	event.CreatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		event.ID,
		event.BatchID,
		event.EventType,
		string(payloadJSON),
		event.ProcessedAt,
		event.ErrorMessage,
		event.RetryCount,
		event.CreatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("telemetry event with ID %s already exists: %w", event.ID.String(), err)
		}
		return fmt.Errorf("failed to create telemetry event: %w", err)
	}

	return nil
}

// GetByID retrieves a telemetry event by its ID
func (r *TelemetryEventRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.TelemetryEvent, error) {
	var event observability.TelemetryEvent
	var payloadJSON string

	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&event.ID,
		&event.BatchID,
		&event.EventType,
		&payloadJSON,
		&event.ProcessedAt,
		&event.ErrorMessage,
		&event.RetryCount,
		&event.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("telemetry event with ID %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get telemetry event by ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(payloadJSON), &event.EventPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	return &event, nil
}

// Update updates an existing telemetry event
func (r *TelemetryEventRepository) Update(ctx context.Context, event *observability.TelemetryEvent) error {
	// Convert event payload to JSON for storage
	payloadJSON, err := json.Marshal(event.EventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := `
		UPDATE telemetry_events
		SET event_type = $2, event_payload = $3, processed_at = $4,
			error_message = $5, retry_count = $6
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query,
		event.ID,
		event.EventType,
		string(payloadJSON),
		event.ProcessedAt,
		event.ErrorMessage,
		event.RetryCount,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update telemetry event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", event.ID.String())
	}

	return nil
}

// Delete deletes a telemetry event by its ID
func (r *TelemetryEventRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM telemetry_events WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete telemetry event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", id.String())
	}

	return nil
}

// GetByBatchID retrieves all telemetry events for a specific batch
func (r *TelemetryEventRepository) GetByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE batch_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, batchID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry events by batch ID: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// GetUnprocessedByBatchID retrieves unprocessed telemetry events for a specific batch
func (r *TelemetryEventRepository) GetUnprocessedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE batch_id = $1 AND processed_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, batchID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed telemetry events: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// GetFailedByBatchID retrieves failed telemetry events for a specific batch
func (r *TelemetryEventRepository) GetFailedByBatchID(ctx context.Context, batchID ulid.ULID) ([]*observability.TelemetryEvent, error) {
	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE batch_id = $1 AND error_message IS NOT NULL AND error_message != ''
		ORDER BY created_at ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, batchID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query failed telemetry events: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// GetByEventType retrieves telemetry events by event type with pagination
func (r *TelemetryEventRepository) GetByEventType(ctx context.Context, eventType observability.TelemetryEventType, limit, offset int) ([]*observability.TelemetryEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, eventType, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry events by type: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// GetUnprocessedByType retrieves unprocessed telemetry events by type
func (r *TelemetryEventRepository) GetUnprocessedByType(ctx context.Context, eventType observability.TelemetryEventType, limit int) ([]*observability.TelemetryEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE event_type = $1 AND processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, eventType, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed telemetry events by type: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// MarkAsProcessed marks a telemetry event as processed
func (r *TelemetryEventRepository) MarkAsProcessed(ctx context.Context, id ulid.ULID, processedAt time.Time) error {
	query := `
		UPDATE telemetry_events
		SET processed_at = $2, error_message = NULL
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, id, processedAt)

	if result.Error != nil {
		return fmt.Errorf("failed to mark telemetry event as processed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", id.String())
	}

	return nil
}

// MarkAsFailed marks a telemetry event as failed with an error message
func (r *TelemetryEventRepository) MarkAsFailed(ctx context.Context, id ulid.ULID, errorMessage string) error {
	query := `
		UPDATE telemetry_events
		SET error_message = $2
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, id, errorMessage)

	if result.Error != nil {
		return fmt.Errorf("failed to mark telemetry event as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", id.String())
	}

	return nil
}

// IncrementRetryCount increments the retry count for a telemetry event
func (r *TelemetryEventRepository) IncrementRetryCount(ctx context.Context, id ulid.ULID) error {
	query := `
		UPDATE telemetry_events
		SET retry_count = retry_count + 1
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, id)

	if result.Error != nil {
		return fmt.Errorf("failed to increment retry count: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", id.String())
	}

	return nil
}

// CreateBatch creates multiple telemetry events in a single transaction
func (r *TelemetryEventRepository) CreateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := r.createWithTx(ctx, tx, event); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple telemetry events in a single transaction
func (r *TelemetryEventRepository) UpdateBatch(ctx context.Context, events []*observability.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := r.updateWithTx(ctx, tx, event); err != nil {
				return err
			}
		}
		return nil
	})
}

// ProcessBatch processes all events in a batch using the provided processor function
func (r *TelemetryEventRepository) ProcessBatch(ctx context.Context, batchID ulid.ULID, processor func([]*observability.TelemetryEvent) error) error {
	// Get all unprocessed events for the batch
	events, err := r.GetUnprocessedByBatchID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to process
	}

	// Process events
	if err := processor(events); err != nil {
		return fmt.Errorf("event processing failed: %w", err)
	}

	return nil
}

// GetEventsForRetry retrieves events that should be retried
func (r *TelemetryEventRepository) GetEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*observability.TelemetryEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, batch_id, event_type, event_payload, processed_at,
			   error_message, retry_count, created_at
		FROM telemetry_events
		WHERE retry_count < $1
		  AND error_message IS NOT NULL
		  AND error_message != ''
		  AND processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, maxRetries, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query events for retry: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// GetFailedEvents retrieves failed events with optional batch filtering
func (r *TelemetryEventRepository) GetFailedEvents(ctx context.Context, batchID *ulid.ULID, limit, offset int) ([]*observability.TelemetryEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	var query string
	var args []interface{}

	if batchID != nil {
		query = `
			SELECT id, batch_id, event_type, event_payload, processed_at,
				   error_message, retry_count, created_at
			FROM telemetry_events
			WHERE batch_id = $1 AND error_message IS NOT NULL AND error_message != ''
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*batchID, limit, offset}
	} else {
		query = `
			SELECT id, batch_id, event_type, event_payload, processed_at,
				   error_message, retry_count, created_at
			FROM telemetry_events
			WHERE error_message IS NOT NULL AND error_message != ''
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query failed events: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEvents(rows)
}

// DeleteFailedEvents deletes failed events older than the specified time
func (r *TelemetryEventRepository) DeleteFailedEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM telemetry_events
		WHERE error_message IS NOT NULL
		  AND error_message != ''
		  AND processed_at IS NULL
		  AND created_at < $1
	`

	result := r.db.WithContext(ctx).Exec(query, olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete failed events: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetEventStats retrieves statistics for telemetry events
func (r *TelemetryEventRepository) GetEventStats(ctx context.Context, filter *observability.TelemetryEventFilter) (*observability.TelemetryEventStats, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := `
		SELECT
			COUNT(*) as total_events,
			COUNT(CASE WHEN processed_at IS NOT NULL THEN 1 END) as processed_events,
			COUNT(CASE WHEN error_message IS NOT NULL AND error_message != '' THEN 1 END) as failed_events,
			COUNT(CASE WHEN processed_at IS NULL AND (error_message IS NULL OR error_message = '') THEN 1 END) as pending_events,
			COALESCE(AVG(retry_count), 0) as avg_retry_count
		FROM telemetry_events` + whereClause

	row := r.db.WithContext(ctx).Raw(query, args...).Row()

	var stats observability.TelemetryEventStats
	err = row.Scan(
		&stats.TotalEvents,
		&stats.ProcessedEvents,
		&stats.FailedEvents,
		&stats.PendingEvents,
		&stats.AverageRetryCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	// Calculate success rate
	if stats.TotalEvents > 0 {
		stats.SuccessRate = float64(stats.ProcessedEvents) / float64(stats.TotalEvents) * 100.0
	}

	// Get event type distribution
	eventTypeDistribution, err := r.getEventTypeDistribution(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type distribution: %w", err)
	}
	stats.EventTypeDistribution = eventTypeDistribution

	// Get error distribution
	errorDistribution, err := r.getErrorDistribution(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get error distribution: %w", err)
	}
	stats.ErrorDistribution = errorDistribution

	// Get retry distribution
	retryDistribution, err := r.getRetryDistribution(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get retry distribution: %w", err)
	}
	stats.RetryDistribution = retryDistribution

	return &stats, nil
}

// GetEventTypeDistribution retrieves event type distribution for a batch or all events
func (r *TelemetryEventRepository) GetEventTypeDistribution(ctx context.Context, batchID *ulid.ULID) (map[observability.TelemetryEventType]int, error) {
	var query string
	var args []interface{}

	if batchID != nil {
		query = `
			SELECT event_type, COUNT(*) as count
			FROM telemetry_events
			WHERE batch_id = $1
			GROUP BY event_type
		`
		args = []interface{}{*batchID}
	} else {
		query = `
			SELECT event_type, COUNT(*) as count
			FROM telemetry_events
			GROUP BY event_type
		`
	}

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query event type distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[observability.TelemetryEventType]int)
	for rows.Next() {
		var eventType observability.TelemetryEventType
		var count int

		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event type distribution: %w", err)
		}

		distribution[eventType] = count
	}

	return distribution, nil
}

// CountEvents counts telemetry events matching the filter criteria
func (r *TelemetryEventRepository) CountEvents(ctx context.Context, filter *observability.TelemetryEventFilter) (int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := "SELECT COUNT(*) FROM telemetry_events" + whereClause
	var count int64

	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count telemetry events: %w", err)
	}

	return count, nil
}

// Helper methods

// createWithTx creates a telemetry event within a transaction
func (r *TelemetryEventRepository) createWithTx(ctx context.Context, tx *gorm.DB, event *observability.TelemetryEvent) error {
	if event.ID.IsZero() {
		event.ID = ulid.New()
	}

	payloadJSON, err := json.Marshal(event.EventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := `
		INSERT INTO telemetry_events (
			id, batch_id, event_type, event_payload, processed_at,
			error_message, retry_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	event.CreatedAt = now

	return tx.WithContext(ctx).Exec(query,
		event.ID,
		event.BatchID,
		event.EventType,
		string(payloadJSON),
		event.ProcessedAt,
		event.ErrorMessage,
		event.RetryCount,
		event.CreatedAt,
	).Error
}

// updateWithTx updates a telemetry event within a transaction
func (r *TelemetryEventRepository) updateWithTx(ctx context.Context, tx *gorm.DB, event *observability.TelemetryEvent) error {
	payloadJSON, err := json.Marshal(event.EventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := `
		UPDATE telemetry_events
		SET event_type = $2, event_payload = $3, processed_at = $4,
			error_message = $5, retry_count = $6
		WHERE id = $1
	`

	result := tx.WithContext(ctx).Exec(query,
		event.ID,
		event.EventType,
		string(payloadJSON),
		event.ProcessedAt,
		event.ErrorMessage,
		event.RetryCount,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update telemetry event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event with ID %s not found", event.ID.String())
	}

	return nil
}

// scanTelemetryEvents scans multiple telemetry events from SQL rows
func (r *TelemetryEventRepository) scanTelemetryEvents(rows *sql.Rows) ([]*observability.TelemetryEvent, error) {
	var events []*observability.TelemetryEvent

	for rows.Next() {
		var event observability.TelemetryEvent
		var payloadJSON string

		err := rows.Scan(
			&event.ID,
			&event.BatchID,
			&event.EventType,
			&payloadJSON,
			&event.ProcessedAt,
			&event.ErrorMessage,
			&event.RetryCount,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry event: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(payloadJSON), &event.EventPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
		}

		events = append(events, &event)
	}

	return events, nil
}

// buildWhereClause builds a WHERE clause based on filter criteria
func (r *TelemetryEventRepository) buildWhereClause(filter *observability.TelemetryEventFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.BatchID != nil {
		conditions = append(conditions, fmt.Sprintf("batch_id = $%d", argIndex))
		args = append(args, *filter.BatchID)
		argIndex++
	}

	if len(filter.BatchIDs) > 0 {
		batchPlaceholders := make([]string, len(filter.BatchIDs))
		for i, batchID := range filter.BatchIDs {
			batchPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, batchID)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("batch_id IN (%s)", strings.Join(batchPlaceholders, ",")))
	}

	if filter.EventType != nil {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIndex))
		args = append(args, *filter.EventType)
		argIndex++
	}

	if len(filter.EventTypes) > 0 {
		typePlaceholders := make([]string, len(filter.EventTypes))
		for i, eventType := range filter.EventTypes {
			typePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, eventType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(typePlaceholders, ",")))
	}

	if filter.IsProcessed != nil {
		if *filter.IsProcessed {
			conditions = append(conditions, "processed_at IS NOT NULL")
		} else {
			conditions = append(conditions, "processed_at IS NULL")
		}
	}

	if filter.HasError != nil {
		if *filter.HasError {
			conditions = append(conditions, "error_message IS NOT NULL AND error_message != ''")
		} else {
			conditions = append(conditions, "(error_message IS NULL OR error_message = '')")
		}
	}

	if filter.MinRetryCount != nil {
		conditions = append(conditions, fmt.Sprintf("retry_count >= $%d", argIndex))
		args = append(args, *filter.MinRetryCount)
		argIndex++
	}

	if filter.MaxRetryCount != nil {
		conditions = append(conditions, fmt.Sprintf("retry_count <= $%d", argIndex))
		args = append(args, *filter.MaxRetryCount)
		argIndex++
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

	if len(conditions) == 0 {
		return "", args, nil
	}

	return " WHERE " + strings.Join(conditions, " AND "), args, nil
}

// getEventTypeDistribution gets event type distribution for stats
func (r *TelemetryEventRepository) getEventTypeDistribution(ctx context.Context, filter *observability.TelemetryEventFilter) (map[observability.TelemetryEventType]int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT event_type, COUNT(*) as count
		FROM telemetry_events` + whereClause + `
		GROUP BY event_type
	`

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query event type distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[observability.TelemetryEventType]int64)
	for rows.Next() {
		var eventType observability.TelemetryEventType
		var count int64

		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event type distribution: %w", err)
		}

		distribution[eventType] = count
	}

	return distribution, nil
}

// getErrorDistribution gets error distribution for stats
func (r *TelemetryEventRepository) getErrorDistribution(ctx context.Context, filter *observability.TelemetryEventFilter) (map[string]int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, err
	}

	// Add error condition to where clause
	if whereClause == "" {
		whereClause = " WHERE error_message IS NOT NULL AND error_message != ''"
	} else {
		whereClause += " AND error_message IS NOT NULL AND error_message != ''"
	}

	query := `
		SELECT error_message, COUNT(*) as count
		FROM telemetry_events` + whereClause + `
		GROUP BY error_message
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query error distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var errorMessage string
		var count int64

		err := rows.Scan(&errorMessage, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error distribution: %w", err)
		}

		distribution[errorMessage] = count
	}

	return distribution, nil
}

// getRetryDistribution gets retry distribution for stats
func (r *TelemetryEventRepository) getRetryDistribution(ctx context.Context, filter *observability.TelemetryEventFilter) (map[int]int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT retry_count, COUNT(*) as count
		FROM telemetry_events` + whereClause + `
		GROUP BY retry_count
		ORDER BY retry_count ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query retry distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[int]int64)
	for rows.Next() {
		var retryCount int
		var count int64

		err := rows.Scan(&retryCount, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retry distribution: %w", err)
		}

		distribution[retryCount] = count
	}

	return distribution, nil
}