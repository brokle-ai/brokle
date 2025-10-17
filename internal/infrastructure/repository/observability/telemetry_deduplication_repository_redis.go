package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
	"brokle/pkg/ulid"
)

// TelemetryDeduplicationRepositoryRedis implements Redis-only deduplication
// This is the simplified version for Phase 2 migration
type TelemetryDeduplicationRepositoryRedis struct {
	redis *database.RedisDB
}

// NewTelemetryDeduplicationRepositoryRedis creates a new Redis-only deduplication repository
func NewTelemetryDeduplicationRepositoryRedis(redis *database.RedisDB) *TelemetryDeduplicationRepositoryRedis {
	return &TelemetryDeduplicationRepositoryRedis{
		redis: redis,
	}
}

// Create creates a new telemetry event deduplication entry in Redis with auto-expiry
func (r *TelemetryDeduplicationRepositoryRedis) Create(ctx context.Context, dedup *observability.TelemetryEventDeduplication) error {
	if dedup == nil {
		return fmt.Errorf("dedup entry cannot be nil")
	}

	if dedup.EventID.IsZero() {
		return fmt.Errorf("event ID is required")
	}

	if dedup.BatchID.IsZero() {
		return fmt.Errorf("batch ID is required")
	}

	// Calculate TTL
	now := time.Now()
	ttl := dedup.ExpiresAt.Sub(now)

	if ttl <= 0 {
		return fmt.Errorf("deduplication entry already expired")
	}

	// Store in Redis with auto-expiry (SETEX)
	redisKey := r.buildRedisKey(dedup.EventID)
	err := r.redis.Set(ctx, redisKey, dedup.BatchID.String(), ttl)
	if err != nil {
		return fmt.Errorf("failed to create deduplication entry: %w", err)
	}

	return nil
}

// Exists checks if a deduplication entry exists for the given event ID
func (r *TelemetryDeduplicationRepositoryRedis) Exists(ctx context.Context, eventID ulid.ULID) (bool, error) {
	redisKey := r.buildRedisKey(eventID)

	exists, err := r.redis.Exists(ctx, redisKey)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists > 0, nil
}

// CheckBatchDuplicates checks for duplicate event IDs in a batch using Redis pipeline
func (r *TelemetryDeduplicationRepositoryRedis) CheckBatchDuplicates(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}

	// Use Redis pipeline for efficient batch checking
	pipe := r.redis.Client.Pipeline()

	// Create EXISTS commands for each event ID
	cmds := make([]*redis.IntCmd, len(eventIDs))
	for i, eventID := range eventIDs {
		redisKey := r.buildRedisKey(eventID)
		cmds[i] = pipe.Exists(ctx, redisKey)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check batch duplicates: %w", err)
	}

	// Collect duplicates
	var duplicates []ulid.ULID
	for i, cmd := range cmds {
		exists, err := cmd.Result()
		if err != nil {
			// Skip on error, treat as not found
			continue
		}

		if exists > 0 {
			duplicates = append(duplicates, eventIDs[i])
		}
	}

	return duplicates, nil
}

// CreateBatch creates multiple deduplication entries in Redis using pipeline
func (r *TelemetryDeduplicationRepositoryRedis) CreateBatch(ctx context.Context, entries []*observability.TelemetryEventDeduplication) error {
	if len(entries) == 0 {
		return nil
	}

	// Use Redis pipeline for efficient batch creation
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

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to batch create deduplication entries: %w", err)
	}

	return nil
}

// Delete deletes a deduplication entry from Redis
func (r *TelemetryDeduplicationRepositoryRedis) Delete(ctx context.Context, eventID ulid.ULID) error {
	redisKey := r.buildRedisKey(eventID)

	err := r.redis.Delete(ctx, redisKey)
	if err != nil {
		return fmt.Errorf("failed to delete deduplication entry: %w", err)
	}

	return nil
}

// GetByEventID retrieves batch ID for an event from Redis
func (r *TelemetryDeduplicationRepositoryRedis) GetByEventID(ctx context.Context, eventID ulid.ULID) (*observability.TelemetryEventDeduplication, error) {
	redisKey := r.buildRedisKey(eventID)

	batchIDStr, err := r.redis.Get(ctx, redisKey)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("deduplication entry not found for event %s", eventID.String())
		}
		return nil, fmt.Errorf("failed to get deduplication entry: %w", err)
	}

	batchID, err := ulid.Parse(batchIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse batch ID: %w", err)
	}

	// Get TTL to reconstruct expires_at
	ttl, err := r.redis.Client.TTL(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get TTL: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	return &observability.TelemetryEventDeduplication{
		EventID:     eventID,
		BatchID:     batchID,
		ProjectID:   ulid.ULID{}, // Not stored in Redis (optimization)
		FirstSeenAt: now.Add(-time.Hour), // Approximate (not critical for dedup)
		ExpiresAt:   expiresAt,
	}, nil
}

// CountByProjectID returns count (not supported in Redis-only, returns 0)
// This method is kept for interface compatibility but returns 0 as Redis doesn't store project-level counts
func (r *TelemetryDeduplicationRepositoryRedis) CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error) {
	// Not supported in Redis-only mode (would require scanning all keys)
	// Return 0 for compatibility
	return 0, nil
}

// Helper methods

// buildRedisKey builds a Redis key for deduplication
func (r *TelemetryDeduplicationRepositoryRedis) buildRedisKey(eventID ulid.ULID) string {
	return fmt.Sprintf("telemetry_dedup:%s", eventID.String())
}

// GetStats returns statistics about deduplication cache (approximate)
func (r *TelemetryDeduplicationRepositoryRedis) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Get approximate count using SCAN (non-blocking)
	var cursor uint64
	var count int64

	// Sample first 100 keys matching the pattern
	keys, _, err := r.redis.Client.Scan(ctx, cursor, "telemetry_dedup:*", 100).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	count = int64(len(keys))

	return map[string]interface{}{
		"approximate_count": count,
		"pattern":          "telemetry_dedup:*",
		"storage":          "redis",
		"auto_expiry":      true,
	}, nil
}

// Interface compatibility methods (not supported in Redis-only mode)

// ExistsInBatch checks if event exists in specific batch (not supported in Redis-only)
func (r *TelemetryDeduplicationRepositoryRedis) ExistsInBatch(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID) (bool, error) {
	// Not supported - would need to store batch_id in Redis value
	return false, fmt.Errorf("ExistsInBatch not supported in Redis-only mode")
}

// ExistsWithRedisCheck checks existence with Redis flag (Redis-only always uses Redis)
func (r *TelemetryDeduplicationRepositoryRedis) ExistsWithRedisCheck(ctx context.Context, eventID ulid.ULID) (bool, bool, error) {
	exists, err := r.Exists(ctx, eventID)
	// Second return value indicates "found in Redis" - always true for Redis-only
	return exists, exists, err
}

// StoreInRedis stores event in Redis (equivalent to Create)
func (r *TelemetryDeduplicationRepositoryRedis) StoreInRedis(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID, ttl time.Duration) error {
	dedup := &observability.TelemetryEventDeduplication{
		EventID:   eventID,
		BatchID:   batchID,
		ProjectID: ulid.ULID{}, // Not required for Redis storage
		ExpiresAt: time.Now().Add(ttl),
	}
	return r.Create(ctx, dedup)
}

// GetFromRedis retrieves batch ID from Redis
func (r *TelemetryDeduplicationRepositoryRedis) GetFromRedis(ctx context.Context, eventID ulid.ULID) (*ulid.ULID, error) {
	dedup, err := r.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return &dedup.BatchID, nil
}

// CleanupExpired removes expired entries (not needed - Redis handles auto-expiry)
func (r *TelemetryDeduplicationRepositoryRedis) CleanupExpired(ctx context.Context) (int64, error) {
	// Redis handles auto-expiry via TTL
	return 0, nil
}

// GetExpiredEntries returns expired entries (not supported - Redis auto-expires)
func (r *TelemetryDeduplicationRepositoryRedis) GetExpiredEntries(ctx context.Context, limit int) ([]*observability.TelemetryEventDeduplication, error) {
	// Not supported - Redis auto-expires entries
	return nil, nil
}

// BatchCleanup batch cleanup (not needed - Redis handles auto-expiry)
func (r *TelemetryDeduplicationRepositoryRedis) BatchCleanup(ctx context.Context, olderThan time.Time, batchSize int) (int64, error) {
	// Redis handles auto-expiry via TTL
	return 0, nil
}

// GetByProjectID retrieves entries by project (not supported efficiently in Redis-only)
func (r *TelemetryDeduplicationRepositoryRedis) GetByProjectID(ctx context.Context, projectID ulid.ULID, limit, offset int) ([]*observability.TelemetryEventDeduplication, error) {
	// Not supported - would require scanning all keys
	return nil, fmt.Errorf("GetByProjectID not supported in Redis-only mode")
}

// CleanupByProjectID cleanup by project (not needed - Redis handles auto-expiry)
func (r *TelemetryDeduplicationRepositoryRedis) CleanupByProjectID(ctx context.Context, projectID ulid.ULID, olderThan time.Time) (int64, error) {
	// Redis handles auto-expiry via TTL
	return 0, nil
}
