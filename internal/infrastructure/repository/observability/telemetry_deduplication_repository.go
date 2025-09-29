package observability

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
	"brokle/pkg/ulid"
)

// TelemetryDeduplicationRepository implements the observability.TelemetryDeduplicationRepository interface
// with Redis fallback strategies for high-performance deduplication
type TelemetryDeduplicationRepository struct {
	db    *gorm.DB
	redis *database.RedisDB
}

// NewTelemetryDeduplicationRepository creates a new telemetry deduplication repository instance
func NewTelemetryDeduplicationRepository(db *gorm.DB, redis *database.RedisDB) *TelemetryDeduplicationRepository {
	return &TelemetryDeduplicationRepository{
		db:    db,
		redis: redis,
	}
}

// Create creates a new telemetry event deduplication entry in the database
func (r *TelemetryDeduplicationRepository) Create(ctx context.Context, dedup *observability.TelemetryEventDeduplication) error {
	query := `
		INSERT INTO telemetry_event_deduplication (
			event_id, batch_id, project_id, first_seen_at, expires_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id) DO NOTHING
	`

	now := time.Now()
	dedup.FirstSeenAt = now

	result := r.db.WithContext(ctx).Exec(query,
		dedup.EventID,
		dedup.BatchID,
		dedup.ProjectID,
		dedup.FirstSeenAt,
		dedup.ExpiresAt,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to create telemetry event deduplication: %w", result.Error)
	}

	// Also store in Redis for fast lookups with TTL
	ttl := dedup.ExpiresAt.Sub(now)
	if ttl > 0 {
		redisKey := r.buildRedisKey(dedup.EventID)
		err := r.redis.Set(ctx, redisKey, dedup.BatchID.String(), ttl)
		if err != nil {
			// Log Redis error but don't fail the operation
			// The database entry is the source of truth
			fmt.Printf("Warning: Failed to store deduplication entry in Redis: %v\n", err)
		}
	}

	return nil
}

// GetByEventID retrieves a telemetry event deduplication entry by event ID
func (r *TelemetryDeduplicationRepository) GetByEventID(ctx context.Context, eventID ulid.ULID) (*observability.TelemetryEventDeduplication, error) {
	var dedup observability.TelemetryEventDeduplication

	query := `
		SELECT event_id, batch_id, project_id, first_seen_at, expires_at
		FROM telemetry_event_deduplication
		WHERE event_id = $1
	`

	row := r.db.WithContext(ctx).Raw(query, eventID).Row()

	err := row.Scan(
		&dedup.EventID,
		&dedup.BatchID,
		&dedup.ProjectID,
		&dedup.FirstSeenAt,
		&dedup.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("telemetry event deduplication entry with event ID %s not found", eventID.String())
		}
		return nil, fmt.Errorf("failed to get telemetry event deduplication by ID: %w", err)
	}

	return &dedup, nil
}

// Delete deletes a telemetry event deduplication entry by event ID
func (r *TelemetryDeduplicationRepository) Delete(ctx context.Context, eventID ulid.ULID) error {
	// Delete from database
	query := `DELETE FROM telemetry_event_deduplication WHERE event_id = $1`

	result := r.db.WithContext(ctx).Exec(query, eventID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete telemetry event deduplication: %w", result.Error)
	}

	// Also delete from Redis
	redisKey := r.buildRedisKey(eventID)
	err := r.redis.Delete(ctx, redisKey)
	if err != nil {
		// Log Redis error but don't fail the operation
		fmt.Printf("Warning: Failed to delete deduplication entry from Redis: %v\n", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telemetry event deduplication entry with event ID %s not found", eventID.String())
	}

	return nil
}

// Exists checks if a telemetry event deduplication entry exists for the given event ID
func (r *TelemetryDeduplicationRepository) Exists(ctx context.Context, eventID ulid.ULID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM telemetry_event_deduplication WHERE event_id = $1)`

	var exists bool
	err := r.db.WithContext(ctx).Raw(query, eventID).Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if telemetry event deduplication exists: %w", err)
	}

	return exists, nil
}

// ExistsInBatch checks if a telemetry event deduplication entry exists for the given event ID and batch ID
func (r *TelemetryDeduplicationRepository) ExistsInBatch(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM telemetry_event_deduplication WHERE event_id = $1 AND batch_id = $2)`

	var exists bool
	err := r.db.WithContext(ctx).Raw(query, eventID, batchID).Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if telemetry event deduplication exists in batch: %w", err)
	}

	return exists, nil
}

// ExistsWithRedisCheck performs a fast Redis check first, then falls back to database
// Returns (exists, foundInRedis, error)
func (r *TelemetryDeduplicationRepository) ExistsWithRedisCheck(ctx context.Context, eventID ulid.ULID) (bool, bool, error) {
	redisKey := r.buildRedisKey(eventID)

	// First check Redis for fast lookup
	exists, err := r.redis.Exists(ctx, redisKey)
	if err == nil && exists > 0 {
		// Found in Redis, return immediately
		return true, true, nil
	}

	// Redis miss or error, fallback to database
	dbExists, err := r.Exists(ctx, eventID)
	if err != nil {
		return false, false, err
	}

	// If found in database but not in Redis, update Redis cache
	if dbExists {
		// Get the entry to extract batch ID and TTL
		dedup, err := r.GetByEventID(ctx, eventID)
		if err == nil && !dedup.IsExpired() {
			// Update Redis with remaining TTL
			ttl := dedup.TimeUntilExpiry()
			if ttl > 0 {
				_ = r.redis.Set(ctx, redisKey, dedup.BatchID.String(), ttl)
			}
		}
	}

	return dbExists, false, nil
}

// StoreInRedis stores a deduplication entry in Redis with TTL
func (r *TelemetryDeduplicationRepository) StoreInRedis(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID, ttl time.Duration) error {
	redisKey := r.buildRedisKey(eventID)

	err := r.redis.Set(ctx, redisKey, batchID.String(), ttl)
	if err != nil {
		return fmt.Errorf("failed to store deduplication entry in Redis: %w", err)
	}

	return nil
}

// GetFromRedis retrieves the batch ID for an event from Redis
func (r *TelemetryDeduplicationRepository) GetFromRedis(ctx context.Context, eventID ulid.ULID) (*ulid.ULID, error) {
	redisKey := r.buildRedisKey(eventID)

	batchIDStr, err := r.redis.Get(ctx, redisKey)
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get deduplication entry from Redis: %w", err)
	}

	batchID, err := ulid.Parse(batchIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse batch ID from Redis: %w", err)
	}

	return &batchID, nil
}

// CleanupExpired removes expired telemetry event deduplication entries
func (r *TelemetryDeduplicationRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM telemetry_event_deduplication
		WHERE expires_at < NOW()
	`

	result := r.db.WithContext(ctx).Exec(query)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired telemetry event deduplication entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetExpiredEntries retrieves expired telemetry event deduplication entries
func (r *TelemetryDeduplicationRepository) GetExpiredEntries(ctx context.Context, limit int) ([]*observability.TelemetryEventDeduplication, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT event_id, batch_id, project_id, first_seen_at, expires_at
		FROM telemetry_event_deduplication
		WHERE expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.db.WithContext(ctx).Raw(query, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query expired telemetry event deduplication entries: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEventDeduplication(rows)
}

// BatchCleanup removes expired entries in batches for better performance
func (r *TelemetryDeduplicationRepository) BatchCleanup(ctx context.Context, olderThan time.Time, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	if batchSize > 10000 {
		batchSize = 10000
	}

	var totalDeleted int64

	for {
		query := `
			DELETE FROM telemetry_event_deduplication
			WHERE event_id IN (
				SELECT event_id
				FROM telemetry_event_deduplication
				WHERE expires_at < $1
				LIMIT $2
			)
		`

		result := r.db.WithContext(ctx).Exec(query, olderThan, batchSize)
		if result.Error != nil {
			return totalDeleted, fmt.Errorf("failed to batch cleanup telemetry event deduplication: %w", result.Error)
		}

		deleted := result.RowsAffected
		totalDeleted += deleted

		// If we deleted fewer than the batch size, we're done
		if deleted < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

// GetByProjectID retrieves telemetry event deduplication entries by project ID with pagination
func (r *TelemetryDeduplicationRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryEventDeduplication, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT event_id, batch_id, project_id, first_seen_at, expires_at
		FROM telemetry_event_deduplication
		WHERE project_id = $1
		ORDER BY first_seen_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.WithContext(ctx).Raw(query, projectID, limit, offset).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry event deduplication by project ID: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryEventDeduplication(rows)
}

// CountByProjectID counts telemetry event deduplication entries by project ID
func (r *TelemetryDeduplicationRepository) CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error) {
	query := `SELECT COUNT(*) FROM telemetry_event_deduplication WHERE project_id = $1`

	var count int64
	err := r.db.WithContext(ctx).Raw(query, projectID).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count telemetry event deduplication by project ID: %w", err)
	}

	return count, nil
}

// CleanupByProjectID removes expired entries for a specific project
func (r *TelemetryDeduplicationRepository) CleanupByProjectID(ctx context.Context, projectID ulid.ULID, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM telemetry_event_deduplication
		WHERE project_id = $1 AND expires_at < $2
	`

	result := r.db.WithContext(ctx).Exec(query, projectID, olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup telemetry event deduplication by project ID: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// CheckBatchDuplicates checks for duplicate event IDs in a batch and returns the duplicates
func (r *TelemetryDeduplicationRepository) CheckBatchDuplicates(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}

	// For small batches, use Redis pipeline for fast lookup
	if len(eventIDs) <= 100 {
		return r.checkBatchDuplicatesRedis(ctx, eventIDs)
	}

	// For larger batches, use database query
	return r.checkBatchDuplicatesDB(ctx, eventIDs)
}

// CreateBatch creates multiple telemetry event deduplication entries in a single transaction
func (r *TelemetryDeduplicationRepository) CreateBatch(ctx context.Context, entries []*observability.TelemetryEventDeduplication) error {
	if len(entries) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Prepare batch insert
		valueStrings := make([]string, 0, len(entries))
		valueArgs := make([]interface{}, 0, len(entries)*5)

		now := time.Now()

		for i, entry := range entries {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
				i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))

			entry.FirstSeenAt = now

			valueArgs = append(valueArgs,
				entry.EventID,
				entry.BatchID,
				entry.ProjectID,
				entry.FirstSeenAt,
				entry.ExpiresAt,
			)
		}

		query := fmt.Sprintf(`
			INSERT INTO telemetry_event_deduplication (
				event_id, batch_id, project_id, first_seen_at, expires_at
			) VALUES %s
			ON CONFLICT (event_id) DO NOTHING
		`, strings.Join(valueStrings, ", "))

		err := tx.WithContext(ctx).Exec(query, valueArgs...).Error
		if err != nil {
			return fmt.Errorf("failed to batch create telemetry event deduplication: %w", err)
		}

		// Also store in Redis for fast lookups
		go r.storeBatchInRedis(context.Background(), entries)

		return nil
	})
}

// Helper methods

// buildRedisKey builds a Redis key for deduplication
func (r *TelemetryDeduplicationRepository) buildRedisKey(eventID ulid.ULID) string {
	return fmt.Sprintf("telemetry_dedup:%s", eventID.String())
}

// checkBatchDuplicatesRedis checks for duplicates using Redis pipeline
func (r *TelemetryDeduplicationRepository) checkBatchDuplicatesRedis(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	pipe := r.redis.Client.Pipeline()

	// Create commands for each event ID
	cmds := make([]*redis.IntCmd, len(eventIDs))
	for i, eventID := range eventIDs {
		redisKey := r.buildRedisKey(eventID)
		cmds[i] = pipe.Exists(ctx, redisKey)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Fallback to database check on Redis error
		return r.checkBatchDuplicatesDB(ctx, eventIDs)
	}

	// Collect duplicates
	var duplicates []ulid.ULID
	var redisMisses []ulid.ULID

	for i, cmd := range cmds {
		exists, err := cmd.Result()
		if err != nil {
			// On individual command error, add to misses for DB check
			redisMisses = append(redisMisses, eventIDs[i])
			continue
		}

		if exists > 0 {
			duplicates = append(duplicates, eventIDs[i])
		} else {
			redisMisses = append(redisMisses, eventIDs[i])
		}
	}

	// Check Redis misses in database
	if len(redisMisses) > 0 {
		dbDuplicates, err := r.checkBatchDuplicatesDB(ctx, redisMisses)
		if err != nil {
			return duplicates, err // Return Redis results even if DB check fails
		}
		duplicates = append(duplicates, dbDuplicates...)
	}

	return duplicates, nil
}

// checkBatchDuplicatesDB checks for duplicates using database query
func (r *TelemetryDeduplicationRepository) checkBatchDuplicatesDB(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	// Convert ULIDs to strings for PostgreSQL array
	idStrings := make([]string, len(eventIDs))
	for i, id := range eventIDs {
		idStrings[i] = id.String()
	}

	query := `
		SELECT event_id
		FROM telemetry_event_deduplication
		WHERE event_id = ANY($1)
	`

	rows, err := r.db.WithContext(ctx).Raw(query, pq.Array(idStrings)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to check batch duplicates: %w", err)
	}
	defer rows.Close()

	var duplicates []ulid.ULID
	for rows.Next() {
		var eventIDStr string
		err := rows.Scan(&eventIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate event ID: %w", err)
		}

		eventID, err := ulid.Parse(eventIDStr)
		if err != nil {
			continue // Skip invalid ULIDs
		}

		duplicates = append(duplicates, eventID)
	}

	return duplicates, nil
}

// storeBatchInRedis stores multiple entries in Redis asynchronously
func (r *TelemetryDeduplicationRepository) storeBatchInRedis(ctx context.Context, entries []*observability.TelemetryEventDeduplication) {
	pipe := r.redis.Client.Pipeline()

	now := time.Now()
	for _, entry := range entries {
		if entry.IsExpired() {
			continue // Skip expired entries
		}

		ttl := entry.ExpiresAt.Sub(now)
		if ttl <= 0 {
			continue // Skip entries that will expire immediately
		}

		redisKey := r.buildRedisKey(entry.EventID)
		pipe.Set(ctx, redisKey, entry.BatchID.String(), ttl)
	}

	// Execute pipeline asynchronously - errors are not critical
	_, _ = pipe.Exec(ctx)
}

// scanTelemetryEventDeduplication scans multiple telemetry event deduplication entries from SQL rows
func (r *TelemetryDeduplicationRepository) scanTelemetryEventDeduplication(rows *sql.Rows) ([]*observability.TelemetryEventDeduplication, error) {
	var entries []*observability.TelemetryEventDeduplication

	for rows.Next() {
		var entry observability.TelemetryEventDeduplication

		err := rows.Scan(
			&entry.EventID,
			&entry.BatchID,
			&entry.ProjectID,
			&entry.FirstSeenAt,
			&entry.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry event deduplication: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}