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

// TraceRepository implements the observability.TraceRepository interface
type TraceRepository struct {
	db *gorm.DB
}

// NewTraceRepository creates a new trace repository instance
func NewTraceRepository(db *gorm.DB) *TraceRepository {
	return &TraceRepository{
		db: db,
	}
}

// Create creates a new trace in the database
func (r *TraceRepository) Create(ctx context.Context, trace *observability.Trace) error {
	if trace.ID.IsZero() {
		trace.ID = ulid.New()
	}

	// Convert maps to JSON for storage
	tagsJSON, err := json.Marshal(trace.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(trace.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Prepare SQL statement
	query := `
		INSERT INTO llm_traces (
			id, project_id, session_id, external_trace_id, parent_trace_id,
			name, user_id, tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	trace.CreatedAt = now
	trace.UpdatedAt = now

	err = r.db.WithContext(ctx).Exec(query,
		trace.ID,
		trace.ProjectID,
		trace.SessionID,
		trace.ExternalTraceID,
		trace.ParentTraceID,
		trace.Name,
		trace.UserID,
		string(tagsJSON),
		string(metadataJSON),
		trace.CreatedAt,
		trace.UpdatedAt,
	).Error

	if err != nil {
		if isDuplicateKeyError(err) {
			return observability.NewObservabilityError(
				observability.ErrCodeExternalTraceIDExists,
				"trace with external_trace_id already exists",
			).WithDetail("external_trace_id", trace.ExternalTraceID)
		}
		return fmt.Errorf("failed to create trace: %w", err)
	}

	return nil
}

// GetByID retrieves a trace by its internal ULID
func (r *TraceRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	var trace observability.Trace
	var tagsJSON, metadataJSON string

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, id).Row()

	err := row.Scan(
		&trace.ID,
		&trace.ProjectID,
		&trace.SessionID,
		&trace.ExternalTraceID,
		&trace.ParentTraceID,
		&trace.Name,
		&trace.UserID,
		&tagsJSON,
		&metadataJSON,
		&trace.CreatedAt,
		&trace.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, observability.NewTraceNotFoundError(id.String())
		}
		return nil, fmt.Errorf("failed to get trace by ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(tagsJSON), &trace.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &trace.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &trace, nil
}

// GetByExternalTraceID retrieves a trace by its external trace ID (from SDK)
func (r *TraceRepository) GetByExternalTraceID(ctx context.Context, externalTraceID string) (*observability.Trace, error) {
	var trace observability.Trace
	var tagsJSON, metadataJSON string

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE external_trace_id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, externalTraceID).Row()

	err := row.Scan(
		&trace.ID,
		&trace.ProjectID,
		&trace.SessionID,
		&trace.ExternalTraceID,
		&trace.ParentTraceID,
		&trace.Name,
		&trace.UserID,
		&tagsJSON,
		&metadataJSON,
		&trace.CreatedAt,
		&trace.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, observability.NewTraceNotFoundError(externalTraceID)
		}
		return nil, fmt.Errorf("failed to get trace by external ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(tagsJSON), &trace.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &trace.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &trace, nil
}

// Update updates an existing trace
func (r *TraceRepository) Update(ctx context.Context, trace *observability.Trace) error {
	// Convert maps to JSON for storage
	tagsJSON, err := json.Marshal(trace.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(trace.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE llm_traces
		SET session_id = $2, parent_trace_id = $3, name = $4, user_id = $5,
			tags = $6, metadata = $7, updated_at = $8
		WHERE id = $1
	`

	trace.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Exec(query,
		trace.ID,
		trace.SessionID,
		trace.ParentTraceID,
		trace.Name,
		trace.UserID,
		string(tagsJSON),
		string(metadataJSON),
		trace.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update trace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewTraceNotFoundError(trace.ID.String())
	}

	return nil
}

// Delete deletes a trace by its ID
func (r *TraceRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM llm_traces WHERE id = $1`

	result := r.db.WithContext(ctx).Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete trace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewTraceNotFoundError(id.String())
	}

	return nil
}

// GetByProjectID retrieves traces by project ID with pagination
func (r *TraceRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.Trace, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query traces by project ID: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetByUserID retrieves traces by user ID with pagination
func (r *TraceRepository) GetByUserID(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*observability.Trace, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, userID, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query traces by user ID: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetBySessionID retrieves traces by session ID
func (r *TraceRepository) GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Trace, error) {
	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE session_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, sessionID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query traces by session ID: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// SearchTraces searches traces with filters and returns results with total count
func (r *TraceRepository) SearchTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.Trace, int, error) {
	// Build WHERE clause and arguments
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM llm_traces" + whereClause
	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count traces: %w", err)
	}

	// Get traces
	orderClause := r.buildOrderClause(filter)
	limitClause := r.buildLimitClause(filter)

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces` + whereClause + orderClause + limitClause

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search traces: %w", err)
	}
	defer rows.Close()

	traces, err := r.scanTraces(rows)
	if err != nil {
		return nil, 0, err
	}

	return traces, int(totalCount), nil
}

// GetTraceWithObservations retrieves a trace with all its observations
func (r *TraceRepository) GetTraceWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	// Get the trace first
	trace, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get observations for the trace
	observationQuery := `
		SELECT id, trace_id, external_observation_id, parent_observation_id, type,
			   name, start_time, end_time, level, status_message, version, model, provider,
			   input, output, model_parameters, prompt_tokens, completion_tokens, total_tokens,
			   input_cost, output_cost, total_cost, latency_ms, quality_score, created_at, updated_at
		FROM llm_observations
		WHERE trace_id = $1
		ORDER BY start_time ASC
	`

	obsRows, err := r.db.WithContext(ctx).Raw(observationQuery, id).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query observations: %w", err)
	}
	defer obsRows.Close()

	var observations []observability.Observation
	for obsRows.Next() {
		var obs observability.Observation
		var inputJSON, outputJSON, modelParamsJSON sql.NullString

		err := obsRows.Scan(
			&obs.ID,
			&obs.TraceID,
			&obs.ExternalObservationID,
			&obs.ParentObservationID,
			&obs.Type,
			&obs.Name,
			&obs.StartTime,
			&obs.EndTime,
			&obs.Level,
			&obs.StatusMessage,
			&obs.Version,
			&obs.Model,
			&obs.Provider,
			&inputJSON,
			&outputJSON,
			&modelParamsJSON,
			&obs.PromptTokens,
			&obs.CompletionTokens,
			&obs.TotalTokens,
			&obs.InputCost,
			&obs.OutputCost,
			&obs.TotalCost,
			&obs.LatencyMs,
			&obs.QualityScore,
			&obs.CreatedAt,
			&obs.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan observation: %w", err)
		}

		// Unmarshal JSON fields
		if inputJSON.Valid {
			if err := json.Unmarshal([]byte(inputJSON.String), &obs.Input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}
		}

		if outputJSON.Valid {
			if err := json.Unmarshal([]byte(outputJSON.String), &obs.Output); err != nil {
				return nil, fmt.Errorf("failed to unmarshal output: %w", err)
			}
		}

		if modelParamsJSON.Valid {
			if err := json.Unmarshal([]byte(modelParamsJSON.String), &obs.ModelParameters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal model parameters: %w", err)
			}
		}

		observations = append(observations, obs)
	}

	trace.Observations = observations
	return trace, nil
}

// GetTraceStats retrieves aggregated statistics for a trace
func (r *TraceRepository) GetTraceStats(ctx context.Context, id ulid.ULID) (*observability.TraceStats, error) {
	query := `
		SELECT
			COUNT(*) as total_observations,
			COALESCE(SUM(EXTRACT(MILLISECONDS FROM (end_time - start_time))), 0) as total_latency_ms,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(AVG(quality_score), 0) as avg_quality_score,
			COUNT(CASE WHEN status_message IS NOT NULL AND status_message != '' THEN 1 END) as error_count,
			COUNT(CASE WHEN type IN ('llm', 'generation') THEN 1 END) as llm_observation_count
		FROM llm_observations
		WHERE trace_id = $1
	`

	var stats observability.TraceStats
	var totalLatencyFloat float64

	row := r.db.WithContext(ctx).Raw(query, id).Row()
	err := row.Scan(
		&stats.TotalObservations,
		&totalLatencyFloat,
		&stats.TotalTokens,
		&stats.TotalCost,
		&stats.AverageQualityScore,
		&stats.ErrorCount,
		&stats.LLMObservationCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get trace stats: %w", err)
	}

	stats.TraceID = id
	stats.TotalLatencyMs = int(totalLatencyFloat)

	// Get provider and model distribution
	distQuery := `
		SELECT
			provider,
			model,
			COUNT(*) as count
		FROM llm_observations
		WHERE trace_id = $1 AND provider IS NOT NULL
		GROUP BY provider, model
	`

	distRows, err := r.db.WithContext(ctx).Raw(distQuery, id).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution stats: %w", err)
	}
	defer distRows.Close()

	stats.ProviderDistribution = make(map[string]int)
	stats.ModelDistribution = make(map[string]int)

	for distRows.Next() {
		var provider, model sql.NullString
		var count int

		err := distRows.Scan(&provider, &model, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}

		if provider.Valid {
			stats.ProviderDistribution[provider.String] += count
		}
		if model.Valid {
			stats.ModelDistribution[model.String] += count
		}
	}

	return &stats, nil
}

// CreateBatch creates multiple traces in a single transaction
func (r *TraceRepository) CreateBatch(ctx context.Context, traces []*observability.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, trace := range traces {
			if err := r.createWithTx(ctx, tx, trace); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatch updates multiple traces in a single transaction
func (r *TraceRepository) UpdateBatch(ctx context.Context, traces []*observability.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, trace := range traces {
			if err := r.updateWithTx(ctx, tx, trace); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple traces by their IDs
func (r *TraceRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	// Convert ULIDs to strings for PostgreSQL array
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	query := `DELETE FROM llm_traces WHERE id = ANY($1)`
	result := r.db.WithContext(ctx).Exec(query, pq.Array(idStrings))

	if result.Error != nil {
		return fmt.Errorf("failed to delete traces batch: %w", result.Error)
	}

	return nil
}

// GetTracesByTimeRange retrieves traces within a time range
func (r *TraceRepository) GetTracesByTimeRange(ctx context.Context, projectID ulid.ULID, startTime, endTime time.Time, limit, offset int) ([]*observability.Trace, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, startTime, endTime, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query traces by time range: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// CountTraces counts traces matching the filter criteria
func (r *TraceRepository) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to build where clause: %w", err)
	}

	query := "SELECT COUNT(*) FROM llm_traces" + whereClause
	var count int64

	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}

// GetRecentTraces retrieves the most recent traces for a project
func (r *TraceRepository) GetRecentTraces(ctx context.Context, projectID ulid.ULID, limit int) ([]*observability.Trace, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, project_id, session_id, external_trace_id, parent_trace_id,
			   name, user_id, tags, metadata, created_at, updated_at
		FROM llm_traces
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query recent traces: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// Helper methods

// createWithTx creates a trace within a transaction
func (r *TraceRepository) createWithTx(ctx context.Context, tx *gorm.DB, trace *observability.Trace) error {
	if trace.ID.IsZero() {
		trace.ID = ulid.New()
	}

	tagsJSON, err := json.Marshal(trace.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(trace.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO llm_traces (
			id, project_id, session_id, external_trace_id, parent_trace_id,
			name, user_id, tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	trace.CreatedAt = now
	trace.UpdatedAt = now

	return tx.WithContext(ctx).Exec(query,
		trace.ID,
		trace.ProjectID,
		trace.SessionID,
		trace.ExternalTraceID,
		trace.ParentTraceID,
		trace.Name,
		trace.UserID,
		string(tagsJSON),
		string(metadataJSON),
		trace.CreatedAt,
		trace.UpdatedAt,
	).Error
}

// updateWithTx updates a trace within a transaction
func (r *TraceRepository) updateWithTx(ctx context.Context, tx *gorm.DB, trace *observability.Trace) error {
	tagsJSON, err := json.Marshal(trace.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(trace.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE llm_traces
		SET session_id = $2, parent_trace_id = $3, name = $4, user_id = $5,
			tags = $6, metadata = $7, updated_at = $8
		WHERE id = $1
	`

	trace.UpdatedAt = time.Now()

	result := tx.WithContext(ctx).Exec(query,
		trace.ID,
		trace.SessionID,
		trace.ParentTraceID,
		trace.Name,
		trace.UserID,
		string(tagsJSON),
		string(metadataJSON),
		trace.UpdatedAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update trace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return observability.NewTraceNotFoundError(trace.ID.String())
	}

	return nil
}

// scanTraces scans multiple traces from SQL rows
func (r *TraceRepository) scanTraces(rows *sql.Rows) ([]*observability.Trace, error) {
	var traces []*observability.Trace

	for rows.Next() {
		var trace observability.Trace
		var tagsJSON, metadataJSON string

		err := rows.Scan(
			&trace.ID,
			&trace.ProjectID,
			&trace.SessionID,
			&trace.ExternalTraceID,
			&trace.ParentTraceID,
			&trace.Name,
			&trace.UserID,
			&tagsJSON,
			&metadataJSON,
			&trace.CreatedAt,
			&trace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(tagsJSON), &trace.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &trace.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		traces = append(traces, &trace)
	}

	return traces, nil
}

// buildWhereClause builds a WHERE clause based on filter criteria
func (r *TraceRepository) buildWhereClause(filter *observability.TraceFilter) (string, []interface{}, error) {
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

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.SessionID != nil {
		conditions = append(conditions, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, *filter.SessionID)
		argIndex++
	}

	if filter.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
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

	if len(filter.Tags) > 0 {
		for key, value := range filter.Tags {
			conditions = append(conditions, fmt.Sprintf("tags->>$%d = $%d", argIndex, argIndex+1))
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
func (r *TraceRepository) buildOrderClause(filter *observability.TraceFilter) string {
	if filter == nil || filter.SortBy == "" {
		return " ORDER BY created_at DESC"
	}

	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}

	switch filter.SortBy {
	case "created_at", "updated_at", "name":
		return fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	default:
		return " ORDER BY created_at DESC"
	}
}

// buildLimitClause builds a LIMIT clause based on filter criteria
func (r *TraceRepository) buildLimitClause(filter *observability.TraceFilter) string {
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

