package observability

import (
	"context"
	"fmt"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/pagination"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type scoreRepository struct {
	db clickhouse.Conn
}

// NewScoreRepository creates a new score repository instance
func NewScoreRepository(db clickhouse.Conn) observability.ScoreRepository {
	return &scoreRepository{db: db}
}

// Create inserts a new score into ClickHouse
func (r *scoreRepository) Create(ctx context.Context, score *observability.Score) error {
	// Set version and event_ts for new scores
	// Version is now optional application version (not row version)

	query := `
		INSERT INTO scores (
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		score.ID,
		score.ProjectID,
		score.TraceID,
		score.SpanID,
		score.Name,
		score.Value,
		score.StringValue,
		score.DataType,
		score.Source,
		score.Comment,
		score.EvaluatorName,
		score.EvaluatorVersion,
		score.EvaluatorConfig,
		score.AuthorUserID,
		score.Timestamp,
		score.Version,
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *scoreRepository) Update(ctx context.Context, score *observability.Score) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	// Version is now optional application version (not auto-incremented)

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, score)
}

// Delete performs hard deletion (MergeTree supports lightweight deletes with DELETE mutation)
func (r *scoreRepository) Delete(ctx context.Context, id string) error {
	// MergeTree lightweight DELETE (async mutation, eventually consistent)
	query := `ALTER TABLE scores DELETE WHERE id = ?`
	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a score by its ID (returns latest version)
func (r *scoreRepository) GetByID(ctx context.Context, id string) (*observability.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		FROM scores
		WHERE id = ?
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanScoreRow(row)
}

// GetByTraceID retrieves all scores for a trace
func (r *scoreRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		FROM scores
		WHERE trace_id = ?		ORDER BY timestamp DESC
	`

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query scores by trace: %w", err)
	}
	defer rows.Close()

	return r.scanScores(rows)
}

// GetBySpanID retrieves all scores for a span
func (r *scoreRepository) GetBySpanID(ctx context.Context, spanID string) ([]*observability.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		FROM scores
		WHERE span_id = ?		ORDER BY timestamp DESC
	`

	rows, err := r.db.Query(ctx, query, spanID)
	if err != nil {
		return nil, fmt.Errorf("query scores by span: %w", err)
	}
	defer rows.Close()

	return r.scanScores(rows)
}

// GetByFilter retrieves scores matching the filter
func (r *scoreRepository) GetByFilter(ctx context.Context, filter *observability.ScoreFilter) ([]*observability.Score, error) {
	query := `
		SELECT
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		FROM scores
		WHERE is_deleted = 0
	`

	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.SpanID != nil {
			query += " AND span_id = ?"
			args = append(args, *filter.SpanID)
		}
		if filter.Name != nil {
			query += " AND name = ?"
			args = append(args, *filter.Name)
		}
		if filter.Source != nil {
			query += " AND source = ?"
			args = append(args, *filter.Source)
		}
		if filter.DataType != nil {
			query += " AND data_type = ?"
			args = append(args, *filter.DataType)
		}
		if filter.EvaluatorName != nil {
			query += " AND evaluator_name = ?"
			args = append(args, *filter.EvaluatorName)
		}
		if filter.MinValue != nil {
			query += " AND value >= ?"
			args = append(args, *filter.MinValue)
		}
		if filter.MaxValue != nil {
			query += " AND value <= ?"
			args = append(args, *filter.MaxValue)
		}
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}

	}

	// Determine sort field and direction with SQL injection protection
	allowedSortFields := []string{"timestamp", "value", "dimension", "data_type", "created_at", "updated_at", "id"}
	sortField := "timestamp" // default
	sortDir := "DESC"

	if filter != nil {
		// Validate sort field against whitelist
		if filter.Params.SortBy != "" {
			validated, err := pagination.ValidateSortField(filter.Params.SortBy, allowedSortFields)
			if err != nil {
				return nil, fmt.Errorf("invalid sort field: %w", err)
			}
			if validated != "" {
				sortField = validated
			}
		}
		if filter.Params.SortDir == "asc" {
			sortDir = "ASC"
		}
	}

	query += fmt.Sprintf(" ORDER BY %s %s, id %s", sortField, sortDir, sortDir)

	// Apply limit and offset for pagination
	limit := pagination.DefaultPageSize
	offset := 0
	if filter != nil {
		if filter.Params.Limit > 0 {
			limit = filter.Params.Limit
		}
		offset = filter.Params.GetOffset()
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query scores by filter: %w", err)
	}
	defer rows.Close()

	return r.scanScores(rows)
}

// CreateBatch inserts multiple scores in a single batch
func (r *scoreRepository) CreateBatch(ctx context.Context, scores []*observability.Score) error {
	if len(scores) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO scores (
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, score := range scores {
		err = batch.Append(
			score.ID,
			score.ProjectID,
			score.TraceID,
			score.SpanID,
			score.Name,
			score.Value,
			score.StringValue,
			score.DataType,
			score.Source,
			score.Comment,
			score.EvaluatorName,
			score.EvaluatorVersion,
			score.EvaluatorConfig,
			score.AuthorUserID,
			score.Timestamp,
			score.Version,
			// Removed: event_ts, is_deleted
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of scores matching the filter
func (r *scoreRepository) Count(ctx context.Context, filter *observability.ScoreFilter) (int64, error) {
	query := "SELECT count() FROM scores"
	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.SpanID != nil {
			query += " AND span_id = ?"
			args = append(args, *filter.SpanID)
		}
		if filter.Name != nil {
			query += " AND name = ?"
			args = append(args, *filter.Name)
		}
		if filter.Source != nil {
			query += " AND source = ?"
			args = append(args, *filter.Source)
		}
		if filter.DataType != nil {
			query += " AND data_type = ?"
			args = append(args, *filter.DataType)
		}
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return int64(count), err
}

// Helper function to scan a single score from query row
func (r *scoreRepository) scanScoreRow(row driver.Row) (*observability.Score, error) {
	var score observability.Score

	err := row.Scan(
		&score.ID,
		&score.ProjectID,
		&score.TraceID,
		&score.SpanID,
		&score.Name,
		&score.Value,
		&score.StringValue,
		&score.DataType,
		&score.Source,
		&score.Comment,
		&score.EvaluatorName,
		&score.EvaluatorVersion,
		&score.EvaluatorConfig,
		&score.AuthorUserID,
		&score.Timestamp,
		&score.Version,
		// Removed: event_ts, is_deleted
	)

	if err != nil {
		return nil, fmt.Errorf("scan score: %w", err)
	}

	return &score, nil
}

// Helper function to scan multiple scores from rows
func (r *scoreRepository) scanScores(rows driver.Rows) ([]*observability.Score, error) {
	scores := make([]*observability.Score, 0) // Initialize empty slice to return [] instead of nil

	for rows.Next() {
		var score observability.Score

		err := rows.Scan(
			&score.ID,
			&score.ProjectID,
			&score.TraceID,
			&score.SpanID,
			&score.Name,
			&score.Value,
			&score.StringValue,
			&score.DataType,
			&score.Source,
			&score.Comment,
			&score.EvaluatorName,
			&score.EvaluatorVersion,
			&score.EvaluatorConfig,
			&score.AuthorUserID,
			&score.Timestamp,
			&score.Version,
			// Removed: event_ts, is_deleted
		)

		if err != nil {
			return nil, fmt.Errorf("scan score row: %w", err)
		}

		scores = append(scores, &score)
	}

	return scores, rows.Err()
}
