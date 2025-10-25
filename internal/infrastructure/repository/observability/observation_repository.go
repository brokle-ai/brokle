package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type observationRepository struct {
	db clickhouse.Conn
}

// NewObservationRepository creates a new observation repository instance
func NewObservationRepository(db clickhouse.Conn) observability.ObservationRepository {
	return &observationRepository{db: db}
}

// Create inserts a new OTEL observation (span) into ClickHouse
func (r *observationRepository) Create(ctx context.Context, obs *observability.Observation) error {
	// Set version and event_ts for new observations
	if obs.Version == 0 {
		obs.Version = 1
	}
	obs.EventTs = time.Now()
	obs.UpdatedAt = time.Now()

	// Calculate duration if not set
	obs.CalculateDuration()

	query := `
		INSERT INTO observations (
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		obs.ID,
		obs.TraceID,
		obs.ParentObservationID,
		obs.ProjectID,
		obs.Name,
		obs.SpanKind,
		obs.Type,
		obs.StartTime,
		obs.EndTime,
		obs.DurationMs,
		obs.StatusCode,
		obs.StatusMessage,
		obs.Attributes,
		obs.Input,
		obs.Output,
		obs.Metadata,
		obs.Level,
		obs.ModelName,
		obs.Provider,
		obs.InternalModelID,
		obs.ModelParameters,
		obs.ProvidedUsageDetails,
		obs.UsageDetails,
		obs.ProvidedCostDetails,
		obs.CostDetails,
		obs.TotalCost,
		obs.PromptID,
		obs.PromptName,
		obs.PromptVersion,
		obs.CreatedAt,
		obs.UpdatedAt,
		obs.Version,
		obs.EventTs,
		boolToUint8(obs.IsDeleted),
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *observationRepository) Update(ctx context.Context, obs *observability.Observation) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	obs.Version++
	obs.EventTs = time.Now()
	obs.UpdatedAt = time.Now()

	// Calculate duration if not set
	obs.CalculateDuration()

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, obs)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *observationRepository) Delete(ctx context.Context, id string) error {
	query := `
		INSERT INTO observations
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM observations
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves an observation by its OTEL span_id (returns latest version)
func (r *observationRepository) GetByID(ctx context.Context, id string) (*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM observations
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanObservationRow(row)
}

// GetByTraceID retrieves all observations for a trace
func (r *observationRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM observations
		WHERE trace_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query observations by trace: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetRootSpan retrieves the root span for a trace (parent_observation_id IS NULL)
func (r *observationRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM observations
		WHERE trace_id = ? AND parent_observation_id IS NULL AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, traceID)
	return r.scanObservationRow(row)
}

// GetChildren retrieves child observations of a parent observation
func (r *observationRepository) GetChildren(ctx context.Context, parentObservationID string) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM observations
		WHERE parent_observation_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, parentObservationID)
	if err != nil {
		return nil, fmt.Errorf("query child observations: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetTreeByTraceID retrieves all observations for a trace (recursive tree)
func (r *observationRepository) GetTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Observation, error) {
	// Return all observations in start_time order (building tree is done in service layer)
	return r.GetByTraceID(ctx, traceID)
}

// GetByFilter retrieves observations by filter criteria
func (r *observationRepository) GetByFilter(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM observations
		WHERE is_deleted = 0
	`

	args := []interface{}{}

	// Apply filters
	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.ParentID != nil {
			query += " AND parent_observation_id = ?"
			args = append(args, *filter.ParentID)
		}
		if filter.Type != nil {
			query += " AND type = ?"
			args = append(args, *filter.Type)
		}
		if filter.SpanKind != nil {
			query += " AND span_kind = ?"
			args = append(args, *filter.SpanKind)
		}
		if filter.Model != nil {
			query += " AND model_name = ?"
			args = append(args, *filter.Model)
		}
		if filter.Level != nil {
			query += " AND level = ?"
			args = append(args, *filter.Level)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
		if filter.MinLatencyMs != nil {
			query += " AND duration_ms >= ?"
			args = append(args, *filter.MinLatencyMs)
		}
		if filter.MaxLatencyMs != nil {
			query += " AND duration_ms <= ?"
			args = append(args, *filter.MaxLatencyMs)
		}
		if filter.MinCost != nil {
			query += " AND total_cost >= ?"
			args = append(args, *filter.MinCost)
		}
		if filter.MaxCost != nil {
			query += " AND total_cost <= ?"
			args = append(args, *filter.MaxCost)
		}
		if filter.IsCompleted != nil {
			if *filter.IsCompleted {
				query += " AND end_time IS NOT NULL"
			} else {
				query += " AND end_time IS NULL"
			}
		}
	}

	// Order by start_time descending
	query += " ORDER BY start_time DESC"

	// Apply limit and offset
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query observations by filter: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// CreateBatch inserts multiple observations in a single batch
func (r *observationRepository) CreateBatch(ctx context.Context, observations []*observability.Observation) error {
	if len(observations) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO observations (
			id, trace_id, parent_observation_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, obs := range observations {
		// Set version and event_ts for new observations
		if obs.Version == 0 {
			obs.Version = 1
			obs.EventTs = time.Now()
		}
		if obs.UpdatedAt.IsZero() {
			obs.UpdatedAt = time.Now()
		}

		// Calculate duration if not set
		obs.CalculateDuration()

		err = batch.Append(
			obs.ID,
			obs.TraceID,
			obs.ParentObservationID,
			obs.ProjectID,
			obs.Name,
			obs.SpanKind,
			obs.Type,
			obs.StartTime,
			obs.EndTime,
			obs.DurationMs,
			obs.StatusCode,
			obs.StatusMessage,
			obs.Attributes,
			obs.Input,
			obs.Output,
			obs.Metadata,
			obs.Level,
			obs.ModelName,
			obs.Provider,
			obs.InternalModelID,
			obs.ModelParameters,
			obs.ProvidedUsageDetails,
			obs.UsageDetails,
			obs.ProvidedCostDetails,
			obs.CostDetails,
			obs.TotalCost,
			obs.PromptID,
			obs.PromptName,
			obs.PromptVersion,
			obs.CreatedAt,
			obs.UpdatedAt,
			obs.Version,
			obs.EventTs,
			boolToUint8(obs.IsDeleted),
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of observations matching the filter
func (r *observationRepository) Count(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	query := "SELECT count() FROM observations WHERE is_deleted = 0"
	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.Type != nil {
			query += " AND type = ?"
			args = append(args, *filter.Type)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a single observation from query row
func (r *observationRepository) scanObservationRow(row driver.Row) (*observability.Observation, error) {
	var obs observability.Observation
	var isDeleted uint8

	err := row.Scan(
		&obs.ID,
		&obs.TraceID,
		&obs.ParentObservationID,
		&obs.ProjectID,
		&obs.Name,
		&obs.SpanKind,
		&obs.Type,
		&obs.StartTime,
		&obs.EndTime,
		&obs.DurationMs,
		&obs.StatusCode,
		&obs.StatusMessage,
		&obs.Attributes,
		&obs.Input,
		&obs.Output,
		&obs.Metadata,
		&obs.Level,
		&obs.ModelName,
		&obs.Provider,
		&obs.InternalModelID,
		&obs.ModelParameters,
		&obs.ProvidedUsageDetails,
		&obs.UsageDetails,
		&obs.ProvidedCostDetails,
		&obs.CostDetails,
		&obs.TotalCost,
		&obs.PromptID,
		&obs.PromptName,
		&obs.PromptVersion,
		&obs.CreatedAt,
		&obs.UpdatedAt,
		&obs.Version,
		&obs.EventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("scan observation: %w", err)
	}

	obs.IsDeleted = isDeleted != 0

	return &obs, nil
}

// Helper function to scan observations from query rows
func (r *observationRepository) scanObservations(rows driver.Rows) ([]*observability.Observation, error) {
	var observations []*observability.Observation

	for rows.Next() {
		var obs observability.Observation
		var isDeleted uint8

		err := rows.Scan(
			&obs.ID,
			&obs.TraceID,
			&obs.ParentObservationID,
			&obs.ProjectID,
			&obs.Name,
			&obs.SpanKind,
			&obs.Type,
			&obs.StartTime,
			&obs.EndTime,
			&obs.DurationMs,
			&obs.StatusCode,
			&obs.StatusMessage,
			&obs.Attributes,
			&obs.Input,
			&obs.Output,
			&obs.Metadata,
			&obs.Level,
			&obs.ModelName,
			&obs.Provider,
			&obs.InternalModelID,
			&obs.ModelParameters,
			&obs.ProvidedUsageDetails,
			&obs.UsageDetails,
			&obs.ProvidedCostDetails,
			&obs.CostDetails,
			&obs.TotalCost,
			&obs.PromptID,
			&obs.PromptName,
			&obs.PromptVersion,
			&obs.CreatedAt,
			&obs.UpdatedAt,
			&obs.Version,
			&obs.EventTs,
			&isDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}

		obs.IsDeleted = isDeleted != 0

		observations = append(observations, &obs)
	}

	return observations, rows.Err()
}
