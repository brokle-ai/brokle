package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
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

// Create inserts a new observation into ClickHouse
func (r *observationRepository) Create(ctx context.Context, obs *observability.Observation) error {
	// Set version and event_ts for new observations
	// Only set version to 1 if it's currently 0 (new record)
	// This allows Update() to increment version without being reset
	if obs.Version == 0 {
		obs.Version = 1
	}
	obs.EventTs = time.Now()

	query := `
		INSERT INTO observations (
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		obs.ID.String(),
		obs.TraceID.String(),
		ulidPtrToString(obs.ParentObservationID),
		obs.ProjectID.String(),
		string(obs.Type),
		obs.Name,
		obs.StartTime,
		obs.EndTime,
		obs.Model,
		obs.ModelParameters,
		obs.Input,
		obs.Output,
		obs.Metadata,
		obs.CostDetails,
		obs.UsageDetails,
		string(obs.Level),
		obs.StatusMessage,
		obs.CompletionStartTime,
		obs.TimeToFirstTokenMs,
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

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, obs)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *observationRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `
		INSERT INTO observations
		SELECT
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM observations
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id.String())
}

// GetByID retrieves an observation by its ID (returns latest version)
func (r *observationRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version, event_ts, is_deleted
		FROM observations
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id.String())
	return r.scanObservation(row)
}

// GetByTraceID retrieves all observations for a trace
func (r *observationRepository) GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version, event_ts, is_deleted
		FROM observations
		WHERE trace_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, traceID.String())
	if err != nil {
		return nil, fmt.Errorf("query observations by trace: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetChildren retrieves child observations of a parent observation
func (r *observationRepository) GetChildren(ctx context.Context, parentObservationID ulid.ULID) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version, event_ts, is_deleted
		FROM observations
		WHERE parent_observation_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, parentObservationID.String())
	if err != nil {
		return nil, fmt.Errorf("query child observations: %w", err)
	}
	defer rows.Close()

	return r.scanObservations(rows)
}

// GetTreeByTraceID retrieves all observations for a trace and builds a hierarchical tree
func (r *observationRepository) GetTreeByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	// Get all observations for the trace
	observations, err := r.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup
	obsMap := make(map[string]*observability.Observation)
	for _, obs := range observations {
		obsMap[obs.ID.String()] = obs
	}

	// Build the tree by linking children to parents
	var rootObservations []*observability.Observation
	for _, obs := range observations {
		if obs.ParentObservationID == nil {
			// This is a root observation
			rootObservations = append(rootObservations, obs)
		} else {
			// This is a child - link it to its parent
			parent, exists := obsMap[obs.ParentObservationID.String()]
			if exists {
				if parent.ChildObservations == nil {
					parent.ChildObservations = make([]*observability.Observation, 0)
				}
				parent.ChildObservations = append(parent.ChildObservations, obs)
			}
		}
	}

	return rootObservations, nil
}

// GetByFilter retrieves observations matching the filter
func (r *observationRepository) GetByFilter(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, error) {
	query := `
		SELECT
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
			version, event_ts, is_deleted
		FROM observations
		WHERE is_deleted = 0
	`

	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, filter.TraceID.String())
		}
		if filter.ParentID != nil {
			query += " AND parent_observation_id = ?"
			args = append(args, filter.ParentID.String())
		}
		if filter.Type != nil {
			query += " AND type = ?"
			args = append(args, string(*filter.Type))
		}
		if filter.Model != nil {
			query += " AND model = ?"
			args = append(args, *filter.Model)
		}
		if filter.Level != nil {
			query += " AND level = ?"
			args = append(args, string(*filter.Level))
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
		if filter.IsCompleted != nil {
			if *filter.IsCompleted {
				query += " AND end_time IS NOT NULL"
			} else {
				query += " AND end_time IS NULL"
			}
		}
	}

	// Order by start_time
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
			id, trace_id, parent_observation_id, project_id, type, name,
			start_time, end_time, model, model_parameters, input, output, metadata,
			cost_details, usage_details, level, status_message,
			completion_start_time, time_to_first_token_ms,
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

		err = batch.Append(
			obs.ID.String(),
			obs.TraceID.String(),
			ulidPtrToString(obs.ParentObservationID),
			obs.ProjectID.String(),
			string(obs.Type),
			obs.Name,
			obs.StartTime,
			obs.EndTime,
			obs.Model,
			obs.ModelParameters,
			obs.Input,
			obs.Output,
			obs.Metadata,
			obs.CostDetails,
			obs.UsageDetails,
			string(obs.Level),
			obs.StatusMessage,
			obs.CompletionStartTime,
			obs.TimeToFirstTokenMs,
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
			args = append(args, filter.TraceID.String())
		}
		if filter.Type != nil {
			query += " AND type = ?"
			args = append(args, string(*filter.Type))
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

// Helper function to scan a single observation
func (r *observationRepository) scanObservation(row driver.Row) (*observability.Observation, error) {
	var obs observability.Observation
	var (
		idStr, traceID, parentObsID, projectID, obsType, name           *string
		model, input, output, statusMessage                             *string
		modelParams, metadata                                           map[string]string
		costDetails                                                     map[string]float64
		usageDetails                                                    map[string]uint64
		level                                                           string
		startTime, eventTs                                              time.Time
		endTime, completionStartTime                                    *time.Time
		timeToFirstTokenMs                                              *uint32
		version, isDeleted                                              uint32
	)

	err := row.Scan(
		&idStr,
		&traceID,
		&parentObsID,
		&projectID,
		&obsType,
		&name,
		&startTime,
		&endTime,
		&model,
		&modelParams,
		&input,
		&output,
		&metadata,
		&costDetails,
		&usageDetails,
		&level,
		&statusMessage,
		&completionStartTime,
		&timeToFirstTokenMs,
		&version,
		&eventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("scan observation: %w", err)
	}

	// Parse ULIDs and strings
	if idStr != nil {
		parsedID, _ := ulid.Parse(*idStr)
		obs.ID = parsedID
	}
	if traceID != nil {
		parsedTraceID, _ := ulid.Parse(*traceID)
		obs.TraceID = parsedTraceID
	}
	if projectID != nil {
		parsedProjectID, _ := ulid.Parse(*projectID)
		obs.ProjectID = parsedProjectID
	}
	obs.ParentObservationID = stringToUlidPtr(parentObsID)
	if obsType != nil {
		obs.Type = observability.ObservationType(*obsType)
	}
	if name != nil {
		obs.Name = *name
	}
	obs.StartTime = startTime
	obs.EndTime = endTime
	obs.Model = model
	obs.ModelParameters = modelParams
	obs.Input = input
	obs.Output = output
	obs.Metadata = metadata
	obs.CostDetails = costDetails
	obs.UsageDetails = usageDetails
	obs.Level = observability.ObservationLevel(level)
	obs.StatusMessage = statusMessage
	obs.CompletionStartTime = completionStartTime
	obs.TimeToFirstTokenMs = timeToFirstTokenMs
	obs.Version = version
	obs.EventTs = eventTs
	obs.IsDeleted = isDeleted != 0

	return &obs, nil
}

// Helper function to scan multiple observations from rows
func (r *observationRepository) scanObservations(rows driver.Rows) ([]*observability.Observation, error) {
	var observations []*observability.Observation

	for rows.Next() {
		var obs observability.Observation
		var (
			idStr, traceID, parentObsID, projectID, obsType, name       *string
			model, input, output, statusMessage                         *string
			modelParams, metadata                                       map[string]string
			costDetails                                                 map[string]float64
			usageDetails                                                map[string]uint64
			level                                                       string
			startTime, eventTs                                          time.Time
			endTime, completionStartTime                                *time.Time
			timeToFirstTokenMs                                          *uint32
			version, isDeleted                                          uint32
		)

		err := rows.Scan(
			&idStr,
			&traceID,
			&parentObsID,
			&projectID,
			&obsType,
			&name,
			&startTime,
			&endTime,
			&model,
			&modelParams,
			&input,
			&output,
			&metadata,
			&costDetails,
			&usageDetails,
			&level,
			&statusMessage,
			&completionStartTime,
			&timeToFirstTokenMs,
			&version,
			&eventTs,
			&isDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("scan observation row: %w", err)
		}

		// Parse ULIDs and strings
		if idStr != nil {
			parsedID, _ := ulid.Parse(*idStr)
			obs.ID = parsedID
		}
		if traceID != nil {
			parsedTraceID, _ := ulid.Parse(*traceID)
			obs.TraceID = parsedTraceID
		}
		if projectID != nil {
			parsedProjectID, _ := ulid.Parse(*projectID)
			obs.ProjectID = parsedProjectID
		}
		obs.ParentObservationID = stringToUlidPtr(parentObsID)
		if obsType != nil {
			obs.Type = observability.ObservationType(*obsType)
		}
		if name != nil {
			obs.Name = *name
		}
		obs.StartTime = startTime
		obs.EndTime = endTime
		obs.Model = model
		obs.ModelParameters = modelParams
		obs.Input = input
		obs.Output = output
		obs.Metadata = metadata
		obs.CostDetails = costDetails
		obs.UsageDetails = usageDetails
		obs.Level = observability.ObservationLevel(level)
		obs.StatusMessage = statusMessage
		obs.CompletionStartTime = completionStartTime
		obs.TimeToFirstTokenMs = timeToFirstTokenMs
		obs.Version = version
		obs.EventTs = eventTs
		obs.IsDeleted = isDeleted != 0

		observations = append(observations, &obs)
	}

	return observations, rows.Err()
}
