package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"

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
	score.EventTs = time.Now()

	query := `
		INSERT INTO scores (
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version, event_ts, is_deleted
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
		score.EventTs,
		boolToUint8(score.IsDeleted),
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *scoreRepository) Update(ctx context.Context, score *observability.Score) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	// Version is now optional application version (not auto-incremented)
	score.EventTs = time.Now()

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, score)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *scoreRepository) Delete(ctx context.Context, id string) error {
	query := `
		INSERT INTO scores
		SELECT
			id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source, comment,
			evaluator_name, evaluator_version, evaluator_config,
			author_user_id, timestamp,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM scores
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

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
			version, event_ts, is_deleted
		FROM scores
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
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
			version, event_ts, is_deleted
		FROM scores
		WHERE trace_id = ? AND is_deleted = 0
		ORDER BY timestamp DESC
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
			version, event_ts, is_deleted
		FROM scores
		WHERE span_id = ? AND is_deleted = 0
		ORDER BY timestamp DESC
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
			version, event_ts, is_deleted
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

	// Order by timestamp descending (most recent first)
	query += " ORDER BY timestamp DESC"

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
			version, event_ts, is_deleted
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, score := range scores {
		// Set version and event_ts for new scores
		// Version is now optional application version
		if score.EventTs.IsZero() {
			score.EventTs = time.Now()
		}

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
			score.EventTs,
			boolToUint8(score.IsDeleted),
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of scores matching the filter
func (r *scoreRepository) Count(ctx context.Context, filter *observability.ScoreFilter) (int64, error) {
	query := "SELECT count() FROM scores WHERE is_deleted = 0"
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

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a single score from query row
func (r *scoreRepository) scanScoreRow(row driver.Row) (*observability.Score, error) {
	var score observability.Score
	var isDeleted uint8

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
		&score.EventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("scan score: %w", err)
	}

	score.IsDeleted = isDeleted != 0

	return &score, nil
}

// Helper function to scan multiple scores from rows
func (r *scoreRepository) scanScores(rows driver.Rows) ([]*observability.Score, error) {
	var scores []*observability.Score

	for rows.Next() {
		var score observability.Score
		var isDeleted uint8

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
			&score.EventTs,
			&isDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("scan score row: %w", err)
		}

		score.IsDeleted = isDeleted != 0

		scores = append(scores, &score)
	}

	return scores, rows.Err()
}
