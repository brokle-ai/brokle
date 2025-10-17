package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// telemetryDeduplicationService implements the TelemetryDeduplicationService interface
type telemetryDeduplicationService struct {
	deduplicationRepo observability.TelemetryDeduplicationRepository
}

// NewTelemetryDeduplicationService creates a new telemetry deduplication service
func NewTelemetryDeduplicationService(
	deduplicationRepo observability.TelemetryDeduplicationRepository,
) observability.TelemetryDeduplicationService {
	return &telemetryDeduplicationService{
		deduplicationRepo: deduplicationRepo,
	}
}

// CheckDuplicate checks if an event ID is a duplicate using Redis-only approach
func (s *telemetryDeduplicationService) CheckDuplicate(ctx context.Context, eventID ulid.ULID) (bool, error) {
	if eventID.IsZero() {
		return false, fmt.Errorf("event ID cannot be zero")
	}

	// Use Redis-only exists check (auto-expiry handles cleanup)
	exists, err := s.deduplicationRepo.Exists(ctx, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return exists, nil
}

// CheckBatchDuplicates efficiently checks multiple event IDs for duplicates
func (s *telemetryDeduplicationService) CheckBatchDuplicates(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}

	// Validate event IDs
	for i, eventID := range eventIDs {
		if eventID.IsZero() {
			return nil, fmt.Errorf("event ID at index %d cannot be zero", i)
		}
	}

	// Use repository's optimized batch duplicate checking
	duplicates, err := s.deduplicationRepo.CheckBatchDuplicates(ctx, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check batch duplicates: %w", err)
	}

	return duplicates, nil
}

// ClaimEvents atomically claims event IDs for processing
// Returns: (claimedIDs, duplicateIDs, error)
func (s *telemetryDeduplicationService) ClaimEvents(ctx context.Context, projectID, batchID ulid.ULID, eventIDs []ulid.ULID, ttl time.Duration) ([]ulid.ULID, []ulid.ULID, error) {
	if len(eventIDs) == 0 {
		return nil, nil, nil
	}

	if projectID.IsZero() {
		return nil, nil, fmt.Errorf("project ID cannot be zero")
	}

	if batchID.IsZero() {
		return nil, nil, fmt.Errorf("batch ID cannot be zero")
	}

	// Validate event IDs
	for i, eventID := range eventIDs {
		if eventID.IsZero() {
			return nil, nil, fmt.Errorf("event ID at index %d cannot be zero", i)
		}
	}

	// Delegate to repository for atomic claim operation
	claimedIDs, duplicateIDs, err := s.deduplicationRepo.ClaimEvents(ctx, projectID, batchID, eventIDs, ttl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to claim events: %w", err)
	}

	return claimedIDs, duplicateIDs, nil
}

// ReleaseEvents removes claimed event IDs (for rollback on publish failure)
func (s *telemetryDeduplicationService) ReleaseEvents(ctx context.Context, eventIDs []ulid.ULID) error {
	if len(eventIDs) == 0 {
		return nil
	}

	// Delegate to repository for batch deletion
	if err := s.deduplicationRepo.ReleaseEvents(ctx, eventIDs); err != nil {
		return fmt.Errorf("failed to release events: %w", err)
	}

	return nil
}

// RegisterEvent registers a new event for deduplication with ULID-based TTL
func (s *telemetryDeduplicationService) RegisterEvent(ctx context.Context, eventID ulid.ULID, batchID ulid.ULID, projectID ulid.ULID, ttl time.Duration) error {
	if eventID.IsZero() {
		return fmt.Errorf("event ID cannot be zero")
	}
	if batchID.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}
	if projectID.IsZero() {
		return fmt.Errorf("project ID cannot be zero")
	}

	// Calculate optimal expiration time based on ULID timestamp and TTL
	expiresAt := s.GetExpirationTime(eventID, ttl)

	// Create deduplication entry
	dedup := &observability.TelemetryEventDeduplication{
		EventID:     eventID,
		BatchID:     batchID,
		ProjectID:   projectID,
		FirstSeenAt: time.Now(),
		ExpiresAt:   expiresAt,
	}

	// Validate the deduplication entry
	if validationErrors := dedup.Validate(); len(validationErrors) > 0 {
		return fmt.Errorf("deduplication entry validation failed: %v", validationErrors)
	}

	// Create in repository (this handles both database and Redis)
	if err := s.deduplicationRepo.Create(ctx, dedup); err != nil {
		return fmt.Errorf("failed to register event for deduplication: %w", err)
	}

	return nil
}

// RegisterProcessedEventsBatch registers multiple processed events for deduplication
// Uses the default TTL from configuration (typically 24 hours for telemetry events)
func (s *telemetryDeduplicationService) RegisterProcessedEventsBatch(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, eventIDs []ulid.ULID) error {
	if projectID.IsZero() {
		return fmt.Errorf("project ID cannot be zero")
	}
	if batchID.IsZero() {
		return fmt.Errorf("batch ID cannot be zero")
	}
	if len(eventIDs) == 0 {
		return nil // Nothing to register
	}

	// Default TTL for telemetry events (24 hours is typical for deduplication)
	defaultTTL := 24 * time.Hour

	// Create deduplication entries for all processed events
	dedupEntries := make([]*observability.TelemetryEventDeduplication, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		if eventID.IsZero() {
			continue // Skip invalid event IDs
		}

		// Calculate optimal expiration time based on ULID timestamp and TTL
		expiresAt := s.GetExpirationTime(eventID, defaultTTL)

		dedup := &observability.TelemetryEventDeduplication{
			EventID:     eventID,
			BatchID:     batchID,
			ProjectID:   projectID,
			FirstSeenAt: time.Now(),
			ExpiresAt:   expiresAt,
		}

		// Validate the deduplication entry
		if validationErrors := dedup.Validate(); len(validationErrors) > 0 {
			continue // Skip invalid entries rather than failing the entire batch
		}

		dedupEntries = append(dedupEntries, dedup)
	}

	if len(dedupEntries) == 0 {
		return nil // No valid entries to register
	}

	// Use batch create for better performance
	if err := s.deduplicationRepo.CreateBatch(ctx, dedupEntries); err != nil {
		return fmt.Errorf("failed to register processed events batch for deduplication: %w", err)
	}

	return nil
}

// CalculateOptimalTTL calculates the optimal TTL based on ULID timestamp and default TTL
func (s *telemetryDeduplicationService) CalculateOptimalTTL(ctx context.Context, eventID ulid.ULID, defaultTTL time.Duration) (time.Duration, error) {
	if eventID.IsZero() {
		return 0, fmt.Errorf("event ID cannot be zero")
	}

	// Extract timestamp from ULID
	eventTime := eventID.Time()
	now := time.Now()

	// If the event is very old, use a shorter TTL
	eventAge := now.Sub(eventTime)
	if eventAge > defaultTTL {
		// Event is older than default TTL, use a minimal TTL
		return time.Minute * 5, nil
	}

	// For recent events, use the full default TTL
	// This ensures we don't get false positives for legitimate retries
	return defaultTTL, nil
}

// GetExpirationTime calculates expiration time based on ULID timestamp and base TTL
func (s *telemetryDeduplicationService) GetExpirationTime(eventID ulid.ULID, baseTTL time.Duration) time.Time {
	// Extract timestamp from ULID
	eventTime := eventID.Time()

	// Add the base TTL to the event time
	// This ensures consistent expiration regardless of when we process the event
	return eventTime.Add(baseTTL)
}

// CleanupExpired removes expired deduplication entries
// Note: With Redis-only approach, TTL handles auto-expiry, so manual cleanup is not needed
func (s *telemetryDeduplicationService) CleanupExpired(ctx context.Context) (int64, error) {
	// Redis automatically expires entries based on TTL
	// Return 0 to indicate no manual cleanup performed
	return 0, nil
}

// CleanupByProject removes expired entries for a specific project
// Note: With Redis-only approach, TTL handles auto-expiry per key
func (s *telemetryDeduplicationService) CleanupByProject(ctx context.Context, projectID ulid.ULID, olderThan time.Time) (int64, error) {
	if projectID.IsZero() {
		return 0, fmt.Errorf("project ID cannot be zero")
	}

	// Redis automatically expires entries based on TTL
	// No project-scoped cleanup needed
	return 0, nil
}

// BatchCleanup removes expired entries in batches for better performance
// Note: With Redis-only approach, TTL handles auto-expiry, no batch cleanup needed
func (s *telemetryDeduplicationService) BatchCleanup(ctx context.Context, olderThan time.Time, batchSize int) (int64, error) {
	// Redis automatically expires entries based on TTL
	// No batch cleanup needed
	return 0, nil
}

// SyncToRedis synchronizes deduplication entries to Redis cache
// Note: With Redis-only approach, entries are already in Redis via Create/CreateBatch
func (s *telemetryDeduplicationService) SyncToRedis(ctx context.Context, entries []*observability.TelemetryEventDeduplication) error {
	// Redis-only approach - entries are already in Redis
	// No sync needed
	return nil
}

// ValidateRedisHealth checks Redis connectivity and performance
func (s *telemetryDeduplicationService) ValidateRedisHealth(ctx context.Context) (*observability.RedisHealthStatus, error) {
	// Create a test key to measure latency
	testKey := ulid.New()
	testBatchID := ulid.New()
	testProjectID := ulid.New()

	start := time.Now()

	// Test Redis operations using Create
	testEntry := &observability.TelemetryEventDeduplication{
		EventID:   testKey,
		BatchID:   testBatchID,
		ProjectID: testProjectID,
		ExpiresAt: time.Now().Add(time.Minute),
	}

	err := s.deduplicationRepo.Create(ctx, testEntry)
	if err != nil {
		errMsg := err.Error()
		return &observability.RedisHealthStatus{
			Available: false,
			LastError: &errMsg,
		}, nil
	}

	// Measure latency
	latency := time.Since(start)

	// Test retrieval using Exists
	exists, getErr := s.deduplicationRepo.Exists(ctx, testKey)
	if getErr != nil {
		errMsg := getErr.Error()
		return &observability.RedisHealthStatus{
			Available: false,
			LastError: &errMsg,
		}, nil
	}

	if !exists {
		errMsg := "test key not found after creation"
		return &observability.RedisHealthStatus{
			Available: false,
			LastError: &errMsg,
		}, nil
	}

	// Cleanup test key
	_ = s.deduplicationRepo.Delete(ctx, testKey)

	return &observability.RedisHealthStatus{
		Available:   true,
		LatencyMs:   float64(latency.Nanoseconds()) / 1_000_000,
		Connections: 1, // This would be actual connection count in production
		Uptime:      time.Hour * 24, // This would be actual uptime in production
	}, nil
}

// GetDeduplicationStats retrieves deduplication performance statistics
func (s *telemetryDeduplicationService) GetDeduplicationStats(ctx context.Context, projectID ulid.ULID) (*observability.DeduplicationStats, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("project ID cannot be zero")
	}

	// Get count of entries for the project
	totalEntries, err := s.deduplicationRepo.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deduplication stats: %w", err)
	}

	// In a real implementation, these metrics would be collected from actual operations
	// For now, we'll return basic statistics
	stats := &observability.DeduplicationStats{
		ProjectID:         projectID,
		TotalChecks:       totalEntries * 2, // Assuming 2 checks per entry on average
		CacheHits:         totalEntries * 8 / 10, // Assuming 80% cache hit rate
		CacheMisses:       totalEntries * 2 / 10, // Assuming 20% cache miss rate
		DatabaseFallbacks: totalEntries * 1 / 10, // Assuming 10% database fallback rate
		DuplicatesFound:   totalEntries * 1 / 20, // Assuming 5% duplicate rate
	}

	// Calculate derived metrics
	if stats.TotalChecks > 0 {
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalChecks) * 100
		stats.FallbackRate = float64(stats.DatabaseFallbacks) / float64(stats.TotalChecks) * 100
	}

	// Average latency would be measured from actual operations
	stats.AverageLatencyMs = 2.5 // Placeholder for actual measured latency

	return stats, nil
}

// GetCacheHitRate calculates cache hit rate over a time window
func (s *telemetryDeduplicationService) GetCacheHitRate(ctx context.Context, timeWindow time.Duration) (float64, error) {
	// In a real implementation, this would query actual metrics
	// For now, return a simulated cache hit rate

	// Simulate different hit rates based on time window
	if timeWindow < time.Hour {
		return 85.0, nil // Higher hit rate for recent data
	} else if timeWindow < time.Hour*24 {
		return 78.0, nil // Moderate hit rate for daily data
	}

	return 72.0, nil // Lower hit rate for older data
}

// GetFallbackRate calculates database fallback rate over a time window
func (s *telemetryDeduplicationService) GetFallbackRate(ctx context.Context, timeWindow time.Duration) (float64, error) {
	// In a real implementation, this would query actual metrics
	// For now, return a simulated fallback rate

	// Simulate different fallback rates based on time window
	if timeWindow < time.Hour {
		return 5.0, nil // Lower fallback rate for recent data
	} else if timeWindow < time.Hour*24 {
		return 12.0, nil // Moderate fallback rate for daily data
	}

	return 18.0, nil // Higher fallback rate for older data (cache eviction)
}