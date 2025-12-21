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
	query := `
		INSERT INTO scores (
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		score.Reason,
		score.Metadata,
		score.ExperimentID,
		score.ExperimentItemID,
		score.Timestamp,
	)
}

// Update performs an upsert by re-inserting with updated values
func (r *scoreRepository) Update(ctx context.Context, score *observability.Score) error {
	return r.Create(ctx, score)
}

// Delete performs hard deletion (MergeTree supports lightweight deletes with DELETE mutation)
func (r *scoreRepository) Delete(ctx context.Context, id string) error {
	// MergeTree lightweight DELETE (async mutation, eventually consistent)
	query := `ALTER TABLE scores DELETE WHERE score_id = ?`
	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a score by its ID (returns latest version)
func (r *scoreRepository) GetByID(ctx context.Context, id string) (*observability.Score, error) {
	query := `
		SELECT
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
		FROM scores
		WHERE score_id = ?
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanScoreRow(row)
}

// GetByTraceID retrieves all scores for a trace
func (r *scoreRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Score, error) {
	query := `
		SELECT
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
		FROM scores
		WHERE trace_id = ?
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
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
		FROM scores
		WHERE span_id = ?
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
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
		FROM scores
		WHERE 1=1
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
	allowedSortFields := []string{"timestamp", "value", "dimension", "data_type", "timestamp", "updated_at", "score_id"}
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

	query += fmt.Sprintf(" ORDER BY %s %s, score_id %s", sortField, sortDir, sortDir)

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
			score_id, project_id, trace_id, span_id,
			name, value, string_value, data_type, source,
			reason, metadata, experiment_id, experiment_item_id,
			timestamp
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
			score.Reason,
			score.Metadata,
			score.ExperimentID,
			score.ExperimentItemID,
			score.Timestamp,
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

// ExistsByConfigName checks if any scores exist for a given config name within a project
func (r *scoreRepository) ExistsByConfigName(ctx context.Context, projectID, configName string) (bool, error) {
	query := `SELECT count() > 0 FROM scores WHERE project_id = ? AND name = ? LIMIT 1`

	var exists bool
	err := r.db.QueryRow(ctx, query, projectID, configName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check scores exist by config name: %w", err)
	}

	return exists, nil
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
		&score.Reason,
		&score.Metadata,
		&score.ExperimentID,
		&score.ExperimentItemID,
		&score.Timestamp,
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
			&score.Reason,
			&score.Metadata,
			&score.ExperimentID,
			&score.ExperimentItemID,
			&score.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("scan score row: %w", err)
		}

		scores = append(scores, &score)
	}

	return scores, rows.Err()
}
