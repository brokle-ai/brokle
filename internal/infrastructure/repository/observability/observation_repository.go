package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// ObservationRepository implements the observability.ObservationRepository interface
type ObservationRepository struct {
	db *gorm.DB
}

// NewObservationRepository creates a new observation repository instance
func NewObservationRepository(db *gorm.DB) *ObservationRepository {
	return &ObservationRepository{
		db: db,
	}
}

// Create creates a new observation in the database
func (r *ObservationRepository) Create(ctx context.Context, observation *observability.Observation) error {
	if observation.ID.IsZero() {
		observation.ID = ulid.New()
	}

	// Validate required fields
	if observation.TraceID.IsZero() {
		return observability.NewObservabilityError(
			observability.ErrCodeValidation,
			"trace_id is required",
		)
	}

	// Convert maps to JSON for storage
	inputJSON, err := json.Marshal(observation.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(observation.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	modelParamsJSON, err := json.Marshal(observation.ModelParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal model parameters: %w", err)
	}

	// Calculate latency if end_time is provided
	if observation.EndTime != nil && !observation.StartTime.IsZero() {
		latencyMs := int(observation.EndTime.Sub(observation.StartTime).Milliseconds())
		observation.LatencyMs = &latencyMs
	}

	// Prepare SQL statement
	query := `
		INSERT INTO llm_observations (
			id, trace_id, external_observation_id, parent_observation_id, type, name,
			start_time, end_time, level, status_message, version, model, provider,
			input, output, model_parameters, prompt_tokens, completion_tokens,
			total_tokens, input_cost, output_cost, total_cost, latency_ms,
			quality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	now := time.Now()
	observation.CreatedAt = now
	observation.UpdatedAt = now

	// Set start time if not provided
	if observation.StartTime.IsZero() {
		observation.StartTime = now
	}

	err = r.db.WithContext(ctx).Exec(query,
		observation.ID,
		observation.TraceID,
		observation.ExternalObservationID,
		observation.ParentObservationID,
		observation.Type,
		observation.Name,
		observation.StartTime,
		observation.EndTime,
		observation.Level,
		observation.StatusMessage,
		observation.Version,
		observation.Model,
		observation.Provider,
		string(inputJSON),
		string(outputJSON),
		string(modelParamsJSON),
		observation.PromptTokens,
		observation.CompletionTokens,
		observation.TotalTokens,
		observation.InputCost,
		observation.OutputCost,
		observation.TotalCost,
		observation.LatencyMs,
		observation.QualityScore,
		observation.CreatedAt,
		observation.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return observability.NewObservabilityError(
				observability.ErrCodeExternalObservationIDExists,
				"observation with external_observation_id already exists",
			).WithDetail("external_observation_id", observation.ExternalObservationID)
		}
		return fmt.Errorf("failed to create observation: %w", err)
	}

	return nil
}

// GetByID retrieves an observation by its internal ULID
func (r *ObservationRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Observation, error) {
	var observation observability.Observation
	var inputJSON, outputJSON, modelParamsJSON sql.NullString

	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&observation.ID,
		&observation.TraceID,
		&observation.ExternalObservationID,
		&observation.ParentObservationID,
		&observation.Type,
		&observation.Name,
		&observation.StartTime,
		&observation.EndTime,
		&observation.Level,
		&observation.StatusMessage,
		&observation.Version,
		&observation.Model,
		&observation.Provider,
		&inputJSON,
		&outputJSON,
		&modelParamsJSON,
		&observation.PromptTokens,
		&observation.CompletionTokens,
		&observation.TotalTokens,
		&observation.InputCost,
		&observation.OutputCost,
		&observation.TotalCost,
		&observation.LatencyMs,
		&observation.QualityScore,
		&observation.CreatedAt,
		&observation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, observability.NewObservationNotFoundError(id.String())
		}
		return nil, fmt.Errorf("failed to get observation by ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := r.unmarshalJSONFields(&observation, inputJSON, outputJSON, modelParamsJSON); err != nil {
		return nil, err
	}

	return &observation, nil
}

// GetByExternalObservationID retrieves an observation by its external ID
func (r *ObservationRepository) GetByExternalObservationID(ctx context.Context, externalObservationID string) (*observability.Observation, error) {
	var observation observability.Observation
	var inputJSON, outputJSON, modelParamsJSON sql.NullString

	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE external_observation_id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, externalObservationID).Row()

	err := row.Scan(
		&observation.ID,
		&observation.TraceID,
		&observation.ExternalObservationID,
		&observation.ParentObservationID,
		&observation.Type,
		&observation.Name,
		&observation.StartTime,
		&observation.EndTime,
		&observation.Level,
		&observation.StatusMessage,
		&observation.Version,
		&observation.Model,
		&observation.Provider,
		&inputJSON,
		&outputJSON,
		&modelParamsJSON,
		&observation.PromptTokens,
		&observation.CompletionTokens,
		&observation.TotalTokens,
		&observation.InputCost,
		&observation.OutputCost,
		&observation.TotalCost,
		&observation.LatencyMs,
		&observation.QualityScore,
		&observation.CreatedAt,
		&observation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, observability.NewObservationNotFoundError(externalObservationID)
		}
		return nil, fmt.Errorf("failed to get observation by external ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := r.unmarshalJSONFields(&observation, inputJSON, outputJSON, modelParamsJSON); err != nil {
		return nil, err
	}

	return &observation, nil
}

// Update updates an existing observation
func (r *ObservationRepository) Update(ctx context.Context, observation *observability.Observation) error {
	inputJSON, err := json.Marshal(observation.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(observation.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	modelParamsJSON, err := json.Marshal(observation.ModelParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal model parameters: %w", err)
	}

	// Recalculate latency if end_time is provided
	if observation.EndTime != nil && !observation.StartTime.IsZero() {
		latencyMs := int(observation.EndTime.Sub(observation.StartTime).Milliseconds())
		observation.LatencyMs = &latencyMs
	}

	query := `
		UPDATE llm_observations
		SET parent_observation_id = $2, type = $3, name = $4, start_time = $5,
			end_time = $6, level = $7, status_message = $8, version = $9,
			model = $10, provider = $11, input = $12, output = $13,
			model_parameters = $14, prompt_tokens = $15, completion_tokens = $16,
			total_tokens = $17, input_cost = $18, output_cost = $19,
			total_cost = $20, latency_ms = $21, quality_score = $22, updated_at = $23
		WHERE id = $1
	`

	observation.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		observation.ID,
		observation.ParentObservationID,
		observation.Type,
		observation.Name,
		observation.StartTime,
		observation.EndTime,
		observation.Level,
		observation.StatusMessage,
		observation.Version,
		observation.Model,
		observation.Provider,
		string(inputJSON),
		string(outputJSON),
		string(modelParamsJSON),
		observation.PromptTokens,
		observation.CompletionTokens,
		observation.TotalTokens,
		observation.InputCost,
		observation.OutputCost,
		observation.TotalCost,
		observation.LatencyMs,
		observation.QualityScore,
		observation.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update observation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewObservationNotFoundError(observation.ID.String())
	}

	return nil
}

// Delete deletes an observation by its ID
func (r *ObservationRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM llm_observations WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete observation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewObservationNotFoundError(id.String())
	}

	return nil
}

// GetByTraceID retrieves all observations for a trace
func (r *ObservationRepository) GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE trace_id = $1
		ORDER BY start_time ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, traceID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations by trace ID: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetByParentObservationID retrieves child observations for a specific parent
func (r *ObservationRepository) GetByParentObservationID(ctx context.Context, parentID ulid.ULID) ([]*observability.Observation, error) {
	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE parent_observation_id = $1
		ORDER BY start_time ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, parentID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations by parent ID: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetByType retrieves observations by type with pagination
func (r *ObservationRepository) GetByType(ctx context.Context, obsType observability.ObservationType, limit, offset int) ([]*observability.Observation, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE type = $1
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, obsType, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations by type: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetByProvider retrieves observations by provider with pagination
func (r *ObservationRepository) GetByProvider(ctx context.Context, provider string, limit, offset int) ([]*observability.Observation, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE provider = $1
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, provider, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations by provider: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetByModel retrieves observations by provider and model with pagination
func (r *ObservationRepository) GetByModel(ctx context.Context, provider, model string, limit, offset int) ([]*observability.Observation, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type, name,
			   start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens,
			   total_tokens, input_cost, output_cost, total_cost, latency_ms,
			   quality_score, created_at, updated_at
		FROM llm_observations
		WHERE provider = $1 AND model = $2
		ORDER BY start_time DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.WithContext(ctx).Raw(query, provider, model, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations by model: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// SearchObservations searches observations with filters
func (r *ObservationRepository) SearchObservations(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, int, error) {
	// Build WHERE clause and arguments
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM llm_observations o"
	if r.needsTraceJoin(filter) {
		countQuery += " INNER JOIN llm_traces t ON o.trace_id = t.id"
	}
	countQuery += whereClause

	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count observations: %w", err)
	}

	// Get observations
	orderClause := r.buildOrderClause(filter)
	limitClause := r.buildLimitClause(filter)

	query := `
		SELECT o.id, o.trace_id, o.external_observation_id, o.parent_observation_id, o.type, o.name,
			   o.start_time, o.end_time, o.level, o.status_message, o.version, o.model, o.provider,
			   o.input, o.output, o.model_parameters, o.prompt_tokens, o.completion_tokens,
			   o.total_tokens, o.input_cost, o.output_cost, o.total_cost, o.latency_ms,
			   o.quality_score, o.created_at, o.updated_at
		FROM llm_observations o`

	if r.needsTraceJoin(filter) {
		query += " INNER JOIN llm_traces t ON o.trace_id = t.id"
	}

	query += whereClause + orderClause + limitClause

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search observations: %w", err)
	}
	defer rows.Close()

	observations, err := r.scanObservations(rows)
	if err != nil {
		return nil, 0, err
	}

	return observations, int(totalCount), nil
}

// GetObservationStats retrieves aggregated statistics for observations
func (r *ObservationRepository) GetObservationStats(ctx context.Context, filter *observability.ObservationFilter) (*observability.ObservationStats, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := `
		SELECT
			COUNT(*) as total_count,
			COALESCE(AVG(latency_ms), 0) as avg_latency,
			COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms), 0) as p95_latency,
			COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY latency_ms), 0) as p99_latency,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(AVG(CASE WHEN total_tokens > 0 THEN total_cost / total_tokens ELSE 0 END), 0) as avg_cost_per_token,
			COALESCE(AVG(quality_score), 0) as avg_quality_score,
			COALESCE(COUNT(CASE WHEN status_message IS NOT NULL AND status_message != '' THEN 1 END)::FLOAT / COUNT(*), 0) as error_rate,
			COALESCE(COUNT(*) / EXTRACT(EPOCH FROM (MAX(start_time) - MIN(start_time))) * 60, 0) as throughput_per_minute
		FROM llm_observations o`

	if r.needsTraceJoin(filter) {
		query += " INNER JOIN llm_traces t ON o.trace_id = t.id"
	}

	query += whereClause

	var stats observability.ObservationStats

	row := r.db.WithContext(ctx).Raw(query, args...).Row()
	err = row.Scan(
		&stats.TotalCount,
		&stats.AverageLatencyMs,
		&stats.P95LatencyMs,
		&stats.P99LatencyMs,
		&stats.TotalCost,
		&stats.AverageCostPerToken,
		&stats.AverageQualityScore,
		&stats.ErrorRate,
		&stats.ThroughputPerMinute,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get observation stats: %w", err)
	}

	return &stats, nil
}

// CreateBatch creates multiple observations in a single transaction
func (r *ObservationRepository) CreateBatch(ctx context.Context, observations []*observability.Observation) error {
	if len(observations) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, obs := range observations {
			if err := r.createWithTx(ctx, tx, obs); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple observations in a single transaction
func (r *ObservationRepository) UpdateBatch(ctx context.Context, observations []*observability.Observation) error {
	if len(observations) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, obs := range observations {
			if err := r.updateWithTx(ctx, tx, obs); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple observations by their IDs
func (r *ObservationRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	// Convert ULIDs to strings for PostgreSQL array
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	query := `DELETE FROM llm_observations WHERE id = ANY($1)`
	result := r.db.WithContext(ctx).Exec(query, pq.Array(idStrings))

	if result.Error != nil {
		return fmt.Errorf("failed to delete observations batch: %w", result.Error)
	}

	return nil
}

// CompleteObservation marks an observation as completed with end time and additional data
func (r *ObservationRepository) CompleteObservation(ctx context.Context, id ulid.ULID, endTime time.Time, output interface{}, cost *float64) error {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	query := `
		UPDATE llm_observations
		SET end_time = $2, output = $3, total_cost = COALESCE($4, total_cost), updated_at = $5
		WHERE id = $1
	`

	result := r.db.WithContext(ctx).Exec(query, id, endTime, string(outputJSON), cost, time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to complete observation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewObservationNotFoundError(id.String())
	}

	return nil
}

// GetIncompleteObservations retrieves observations that haven't been completed yet
func (r *ObservationRepository) GetIncompleteObservations(ctx context.Context, projectID ulid.ULID) ([]*observability.Observation, error) {
	query := `
		SELECT o.id, o.trace_id, o.external_observation_id, o.parent_observation_id, o.type, o.name,
			   o.start_time, o.end_time, o.level, o.status_message, o.version, o.model, o.provider,
			   o.input, o.output, o.model_parameters, o.prompt_tokens, o.completion_tokens, o.total_tokens,
			   o.input_cost, o.output_cost, o.total_cost, o.latency_ms, o.quality_score, o.created_at, o.updated_at
		FROM llm_observations o
		JOIN llm_traces t ON o.trace_id = t.id
		WHERE t.project_id = $1 AND o.end_time IS NULL
		ORDER BY o.start_time ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query incomplete observations: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetObservationsByTimeRange retrieves observations within a time range
func (r *ObservationRepository) GetObservationsByTimeRange(ctx context.Context, filter *observability.ObservationFilter, startTime, endTime time.Time) ([]*observability.Observation, error) {
	// Add time range to filter
	if filter == nil {
		filter = &observability.ObservationFilter{}
	}
	filter.StartTime = &startTime
	filter.EndTime = &endTime

	observations, _, err := r.SearchObservations(ctx, filter)
	return observations, err
}

// CountObservations counts observations matching the filter criteria
func (r *ObservationRepository) CountObservations(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := "SELECT COUNT(*) FROM llm_observations o"
	if r.needsTraceJoin(filter) {
		query += " INNER JOIN llm_traces t ON o.trace_id = t.id"
	}
	query += whereClause

	var count int64
	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count observations: %w", err)
	}

	return count, nil
}

// Helper methods

// createWithTx creates an observation within a transaction
func (r *ObservationRepository) createWithTx(ctx context.Context, tx *gorm.DB, observation *observability.Observation) error {
	if observation.ID.IsZero() {
		observation.ID = ulid.New()
	}

	inputJSON, err := json.Marshal(observation.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(observation.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	modelParamsJSON, err := json.Marshal(observation.ModelParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal model parameters: %w", err)
	}

	if observation.EndTime != nil && !observation.StartTime.IsZero() {
		latencyMs := int(observation.EndTime.Sub(observation.StartTime).Milliseconds())
		observation.LatencyMs = &latencyMs
	}

	query := `
		INSERT INTO llm_observations (
			id, trace_id, external_observation_id, parent_observation_id, type, name,
			start_time, end_time, level, status_message, version, model, provider,
			input, output, model_parameters, prompt_tokens, completion_tokens,
			total_tokens, input_cost, output_cost, total_cost, latency_ms,
			quality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	now := time.Now()
	observation.CreatedAt = now
	observation.UpdatedAt = now

	if observation.StartTime.IsZero() {
		observation.StartTime = now
	}

	return tx.WithContext(ctx).Exec(query,
		observation.ID,
		observation.TraceID,
		observation.ExternalObservationID,
		observation.ParentObservationID,
		observation.Type,
		observation.Name,
		observation.StartTime,
		observation.EndTime,
		observation.Level,
		observation.StatusMessage,
		observation.Version,
		observation.Model,
		observation.Provider,
		string(inputJSON),
		string(outputJSON),
		string(modelParamsJSON),
		observation.PromptTokens,
		observation.CompletionTokens,
		observation.TotalTokens,
		observation.InputCost,
		observation.OutputCost,
		observation.TotalCost,
		observation.LatencyMs,
		observation.QualityScore,
		observation.CreatedAt,
		observation.UpdatedAt,
	).Error
}

// updateWithTx updates an observation within a transaction
func (r *ObservationRepository) updateWithTx(ctx context.Context, tx *gorm.DB, observation *observability.Observation) error {
	inputJSON, err := json.Marshal(observation.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	outputJSON, err := json.Marshal(observation.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	modelParamsJSON, err := json.Marshal(observation.ModelParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal model parameters: %w", err)
	}

	if observation.EndTime != nil && !observation.StartTime.IsZero() {
		latencyMs := int(observation.EndTime.Sub(observation.StartTime).Milliseconds())
		observation.LatencyMs = &latencyMs
	}

	query := `
		UPDATE llm_observations
		SET parent_observation_id = $2, type = $3, name = $4, start_time = $5,
			end_time = $6, level = $7, status_message = $8, version = $9,
			model = $10, provider = $11, input = $12, output = $13,
			model_parameters = $14, prompt_tokens = $15, completion_tokens = $16,
			total_tokens = $17, input_cost = $18, output_cost = $19,
			total_cost = $20, latency_ms = $21, quality_score = $22, updated_at = $23
		WHERE id = $1
	`

	observation.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		observation.ID,
		observation.ParentObservationID,
		observation.Type,
		observation.Name,
		observation.StartTime,
		observation.EndTime,
		observation.Level,
		observation.StatusMessage,
		observation.Version,
		observation.Model,
		observation.Provider,
		string(inputJSON),
		string(outputJSON),
		string(modelParamsJSON),
		observation.PromptTokens,
		observation.CompletionTokens,
		observation.TotalTokens,
		observation.InputCost,
		observation.OutputCost,
		observation.TotalCost,
		observation.LatencyMs,
		observation.QualityScore,
		observation.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update observation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewObservationNotFoundError(observation.ID.String())
	}

	return nil
}

// scanObservations scans multiple observations from SQL rows
func (r *ObservationRepository) scanObservations(rows *sql.Rows) ([]*observability.Observation, error) {
	var observations []*observability.Observation

	for rows.Next() {
		var observation observability.Observation
		var inputJSON, outputJSON, modelParamsJSON sql.NullString

		err := rows.Scan(
			&observation.ID,
			&observation.TraceID,
			&observation.ExternalObservationID,
			&observation.ParentObservationID,
			&observation.Type,
			&observation.Name,
			&observation.StartTime,
			&observation.EndTime,
			&observation.Level,
			&observation.StatusMessage,
			&observation.Version,
			&observation.Model,
			&observation.Provider,
			&inputJSON,
			&outputJSON,
			&modelParamsJSON,
			&observation.PromptTokens,
			&observation.CompletionTokens,
			&observation.TotalTokens,
			&observation.InputCost,
			&observation.OutputCost,
			&observation.TotalCost,
			&observation.LatencyMs,
			&observation.QualityScore,
			&observation.CreatedAt,
			&observation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan observation: %w", err)
		}

		// Unmarshal JSON fields
		if err := r.unmarshalJSONFields(&observation, inputJSON, outputJSON, modelParamsJSON); err != nil {
			return nil, err
		}

		observations = append(observations, &observation)
	}

	return observations, nil
}

// unmarshalJSONFields unmarshals JSON fields for an observation
func (r *ObservationRepository) unmarshalJSONFields(observation *observability.Observation, inputJSON, outputJSON, modelParamsJSON sql.NullString) error {
	if inputJSON.Valid && inputJSON.String != "null" {
		if err := json.Unmarshal([]byte(inputJSON.String), &observation.Input); err != nil {
			return fmt.Errorf("failed to unmarshal input: %w", err)
		}
	}

	if outputJSON.Valid && outputJSON.String != "null" {
		if err := json.Unmarshal([]byte(outputJSON.String), &observation.Output); err != nil {
			return fmt.Errorf("failed to unmarshal output: %w", err)
		}
	}

	if modelParamsJSON.Valid && modelParamsJSON.String != "null" {
		if err := json.Unmarshal([]byte(modelParamsJSON.String), &observation.ModelParameters); err != nil {
			return fmt.Errorf("failed to unmarshal model parameters: %w", err)
		}
	}

	return nil
}

// needsTraceJoin determines if the query needs a join with the traces table
func (r *ObservationRepository) needsTraceJoin(filter *observability.ObservationFilter) bool {
	if filter == nil {
		return false
	}
	return false // With current filter fields, we don't need joins since we removed project/user/session filters
}

// buildWhereClause builds a WHERE clause based on filter criteria
func (r *ObservationRepository) buildWhereClause(filter *observability.ObservationFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Observation-specific filters
	if filter.TraceID != nil {
		conditions = append(conditions, fmt.Sprintf("o.trace_id = $%d", argIndex))
		args = append(args, *filter.TraceID)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("o.type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Provider != nil {
		conditions = append(conditions, fmt.Sprintf("o.provider = $%d", argIndex))
		args = append(args, *filter.Provider)
		argIndex++
	}

	if filter.Model != nil {
		conditions = append(conditions, fmt.Sprintf("o.model = $%d", argIndex))
		args = append(args, *filter.Model)
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("o.start_time >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("o.start_time <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	if filter.Level != nil {
		conditions = append(conditions, fmt.Sprintf("o.level = $%d", argIndex))
		args = append(args, *filter.Level)
		argIndex++
	}

	if filter.MinLatency != nil {
		conditions = append(conditions, fmt.Sprintf("o.latency_ms >= $%d", argIndex))
		args = append(args, *filter.MinLatency)
		argIndex++
	}

	if filter.MaxLatency != nil {
		conditions = append(conditions, fmt.Sprintf("o.latency_ms <= $%d", argIndex))
		args = append(args, *filter.MaxLatency)
		argIndex++
	}

	if filter.MinCost != nil {
		conditions = append(conditions, fmt.Sprintf("o.total_cost >= $%d", argIndex))
		args = append(args, *filter.MinCost)
		argIndex++
	}

	if filter.MaxCost != nil {
		conditions = append(conditions, fmt.Sprintf("o.total_cost <= $%d", argIndex))
		args = append(args, *filter.MaxCost)
		argIndex++
	}

	if filter.IsCompleted != nil {
		if *filter.IsCompleted {
			conditions = append(conditions, "o.end_time IS NOT NULL")
		} else {
			conditions = append(conditions, "o.end_time IS NULL")
		}
	}

	if filter.HasError != nil {
		if *filter.HasError {
			conditions = append(conditions, "(o.status_message IS NOT NULL AND o.status_message != '')")
		} else {
			conditions = append(conditions, "(o.status_message IS NULL OR o.status_message = '')")
		}
	}

	if len(conditions) == 0 {
		return "", args, nil
	}

	return " WHERE " + strings.Join(conditions, " AND "), args, nil
}

// buildOrderClause builds an ORDER BY clause based on filter criteria
func (r *ObservationRepository) buildOrderClause(filter *observability.ObservationFilter) string {
	if filter == nil || filter.SortBy == "" {
		return " ORDER BY o.start_time DESC"
	}

	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}

	switch filter.SortBy {
	case "start_time", "end_time", "created_at", "updated_at", "name", "total_cost", "latency_ms", "quality_score":
		return fmt.Sprintf(" ORDER BY o.%s %s", filter.SortBy, order)
	default:
		return " ORDER BY o.start_time DESC"
	}
}

// buildLimitClause builds a LIMIT clause based on filter criteria
func (r *ObservationRepository) buildLimitClause(filter *observability.ObservationFilter) string {
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

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
func isDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}