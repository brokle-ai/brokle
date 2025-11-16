package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/streams"
	"brokle/pkg/ulid"
)

// Sentinel errors for message processing states
var (
	// ErrMovedToDLQ indicates message was successfully moved to Dead Letter Queue
	// This signals that the message can be safely acknowledged (data preserved in DLQ)
	ErrMovedToDLQ = errors.New("message moved to DLQ")
)

// Dead Letter Queue constants
const (
	// DLQ stream key prefix
	dlqStreamPrefix = "telemetry:dlq:batches"

	// DLQ retention period (7 days)
	dlqRetentionPeriod = 7 * 24 * time.Hour

	// DLQ max length (prevent unbounded growth)
	dlqMaxLength = 1000
)

// TelemetryStreamConsumer consumes telemetry batches from Redis Streams and writes to ClickHouse
type TelemetryStreamConsumer struct {
	deduplicationSvc    observability.TelemetryDeduplicationService
	traceService        observability.TraceService
	spanService         observability.SpanService
	scoreService        observability.ScoreService
	redis               *database.RedisDB
	logger              *logrus.Logger
	activeStreams       map[string]bool
	quit                chan struct{}
	consumerGroup       string
	consumerID          string
	wg                  sync.WaitGroup
	discoveryInterval   time.Duration
	batchesProcessed    int64
	maxStreamsPerRead   int
	running             int64
	maxRetries          int
	blockDuration       time.Duration
	maxDiscoveryBackoff time.Duration
	retryBackoff        time.Duration
	eventsProcessed     int64
	errorsCount         int64
	dlqMessagesCount    int64
	batchSize           int
	discoveryBackoff    time.Duration
	streamRotation      int
	streamsMutex        sync.RWMutex
	statsLock           sync.RWMutex
}

// TelemetryStreamConsumerConfig holds configuration for the consumer
type TelemetryStreamConsumerConfig struct {
	ConsumerGroup     string
	ConsumerID        string
	BatchSize         int
	BlockDuration     time.Duration
	MaxRetries        int
	RetryBackoff      time.Duration
	DiscoveryInterval time.Duration
	MaxStreamsPerRead int
}

// NewTelemetryStreamConsumer creates a new telemetry stream consumer
func NewTelemetryStreamConsumer(
	redis *database.RedisDB,
	deduplicationSvc observability.TelemetryDeduplicationService,
	logger *logrus.Logger,
	config *TelemetryStreamConsumerConfig,
	// Observability services for structured events
	traceService observability.TraceService,
	spanService observability.SpanService,
	scoreService observability.ScoreService,
) *TelemetryStreamConsumer {
	if config == nil {
		config = &TelemetryStreamConsumerConfig{
			ConsumerGroup:     "telemetry-workers",
			ConsumerID:        "worker-" + ulid.New().String(),
			BatchSize:         50, // Optimized for lower latency and better worker utilization
			BlockDuration:     time.Second,
			MaxRetries:        3,
			RetryBackoff:      500 * time.Millisecond,
			DiscoveryInterval: 30 * time.Second,
			MaxStreamsPerRead: 10,
		}
	}

	return &TelemetryStreamConsumer{
		redis:               redis,
		deduplicationSvc:    deduplicationSvc,
		logger:              logger,
		traceService:        traceService,
		spanService:         spanService,
		scoreService:        scoreService,
		consumerGroup:       config.ConsumerGroup,
		consumerID:          config.ConsumerID,
		batchSize:           config.BatchSize,
		blockDuration:       config.BlockDuration,
		maxRetries:          config.MaxRetries,
		retryBackoff:        config.RetryBackoff,
		discoveryInterval:   config.DiscoveryInterval,
		maxStreamsPerRead:   config.MaxStreamsPerRead,
		quit:                make(chan struct{}),
		activeStreams:       make(map[string]bool),
		discoveryBackoff:    time.Second,
		maxDiscoveryBackoff: 30 * time.Second,
	}
}

// Start begins consuming from Redis Streams
func (c *TelemetryStreamConsumer) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt64(&c.running, 0, 1) {
		return errors.New("consumer already running")
	}

	c.logger.WithFields(logrus.Fields{
		"consumer_group":     c.consumerGroup,
		"consumer_id":        c.consumerID,
		"batch_size":         c.batchSize,
		"discovery_interval": c.discoveryInterval,
		"max_streams":        c.maxStreamsPerRead,
	}).Info("Starting telemetry stream consumer")

	// Start consumption loop
	c.wg.Add(1)
	go c.consumeLoop(ctx)

	// Start stream discovery loop
	c.wg.Add(1)
	go c.discoveryLoop(ctx)

	return nil
}

// Stop gracefully stops the consumer
func (c *TelemetryStreamConsumer) Stop() {
	if !atomic.CompareAndSwapInt64(&c.running, 1, 0) {
		return
	}

	c.logger.Info("Stopping telemetry stream consumer")
	close(c.quit)
	c.wg.Wait()

	c.statsLock.RLock()
	c.logger.WithFields(logrus.Fields{
		"batches_processed": c.batchesProcessed,
		"events_processed":  c.eventsProcessed,
		"errors_count":      c.errorsCount,
	}).Info("Telemetry stream consumer stopped")
	c.statsLock.RUnlock()
}

// discoveryLoop periodically discovers and initializes new streams
func (c *TelemetryStreamConsumer) discoveryLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.discoveryInterval)
	defer ticker.Stop()

	// Initial discovery on startup
	if err := c.performDiscovery(ctx); err != nil {
		c.logger.WithError(err).Error("Initial stream discovery failed")
	}

	for {
		select {
		case <-c.quit:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.performDiscovery(ctx); err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"backoff": c.discoveryBackoff,
				}).Error("Stream discovery failed, backing off")

				// Exponential backoff on failure
				time.Sleep(c.discoveryBackoff)
				c.discoveryBackoff = minDuration(c.discoveryBackoff*2, c.maxDiscoveryBackoff)
			} else {
				// Reset backoff on success
				c.discoveryBackoff = time.Second
			}
		}
	}
}

// performDiscovery discovers streams and initializes consumer groups
func (c *TelemetryStreamConsumer) performDiscovery(ctx context.Context) error {
	streams, err := c.discoverStreams(ctx)
	if err != nil {
		return err
	}

	if len(streams) == 0 {
		c.logger.Debug("No telemetry streams discovered")
		return nil
	}

	// Cleanup inactive streams before adding new ones
	c.cleanupInactiveStreams(streams)

	return c.ensureConsumerGroups(ctx, streams)
}

// discoverStreams discovers active telemetry streams using Redis SCAN
func (c *TelemetryStreamConsumer) discoverStreams(ctx context.Context) ([]string, error) {
	var allStreams []string
	cursor := uint64(0)
	pattern := "telemetry:batches:*"

	// Use SCAN for production-safe iteration (non-blocking)
	for {
		keys, nextCursor, err := c.redis.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan streams: %w", err)
		}

		allStreams = append(allStreams, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break // Completed full scan
		}
	}

	c.logger.WithFields(logrus.Fields{
		"stream_count": len(allStreams),
		"pattern":      pattern,
	}).Debug("Discovered telemetry streams")

	return allStreams, nil
}

// ensureConsumerGroups creates consumer groups for discovered streams
func (c *TelemetryStreamConsumer) ensureConsumerGroups(ctx context.Context, streams []string) error {
	for _, streamKey := range streams {
		// Check if stream is new
		c.streamsMutex.RLock()
		exists := c.activeStreams[streamKey]
		c.streamsMutex.RUnlock()

		if exists {
			continue // Already initialized
		}

		// Create consumer group (idempotent operation)
		// Use "0" to read from beginning, ensuring we don't miss messages that arrived before consumer started
		err := c.redis.Client.XGroupCreateMkStream(ctx, streamKey, c.consumerGroup, "0").Err()
		if err != nil {
			// Ignore BUSYGROUP error (group already exists)
			if !strings.Contains(err.Error(), "BUSYGROUP") {
				c.logger.WithError(err).WithField("stream", streamKey).Warn("Failed to create consumer group")
				continue
			}
		}

		// Mark as active
		c.streamsMutex.Lock()
		c.activeStreams[streamKey] = true
		c.streamsMutex.Unlock()

		c.logger.WithFields(logrus.Fields{
			"stream":         streamKey,
			"consumer_group": c.consumerGroup,
		}).Debug("Consumer group initialized for stream")
	}

	return nil
}

// cleanupInactiveStreams removes streams that no longer exist in Redis
func (c *TelemetryStreamConsumer) cleanupInactiveStreams(discoveredStreams []string) {
	// Build set of current streams for O(1) lookup
	currentStreams := make(map[string]bool, len(discoveredStreams))
	for _, streamKey := range discoveredStreams {
		currentStreams[streamKey] = true
	}

	// Find streams that are active but no longer exist
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()

	var removedStreams []string
	for streamKey := range c.activeStreams {
		if !currentStreams[streamKey] {
			delete(c.activeStreams, streamKey)
			removedStreams = append(removedStreams, streamKey)
		}
	}

	if len(removedStreams) > 0 {
		c.logger.WithFields(logrus.Fields{
			"removed_count": len(removedStreams),
			"streams":       removedStreams,
		}).Info("Cleaned up inactive streams")
	}
}

// minDuration returns the minimum of two durations
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// consumeLoop is the main consumption loop
func (c *TelemetryStreamConsumer) consumeLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-c.quit:
			return
		case <-ctx.Done():
			return
		default:
			if err := c.consumeBatch(ctx); err != nil {
				if err != redis.Nil {
					c.logger.WithError(err).Error("Error consuming batch")
					c.incrementErrors()
				}
				// Brief pause on error to prevent tight error loops
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// consumeBatch reads and processes a batch of messages from discovered streams
func (c *TelemetryStreamConsumer) consumeBatch(ctx context.Context) error {
	// Get active streams with rotation
	c.streamsMutex.Lock()
	var allStreamKeys []string
	for streamKey := range c.activeStreams {
		allStreamKeys = append(allStreamKeys, streamKey)
	}

	// Round-robin rotation for fairness
	// Ensures all projects get processed even if total streams > maxStreamsPerRead
	if len(allStreamKeys) > 0 && c.streamRotation >= len(allStreamKeys) {
		c.streamRotation = 0 // Reset rotation
	}

	// Rotate streams for fair distribution
	if c.streamRotation > 0 && len(allStreamKeys) > c.streamRotation {
		allStreamKeys = append(allStreamKeys[c.streamRotation:], allStreamKeys[:c.streamRotation]...)
	}
	c.streamsMutex.Unlock()

	if len(allStreamKeys) == 0 {
		// No streams discovered yet - wait for discovery loop
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Limit streams per read (Redis best practice)
	streamKeys := allStreamKeys
	if len(streamKeys) > c.maxStreamsPerRead {
		streamKeys = streamKeys[:c.maxStreamsPerRead]
	}

	// Build XReadGroup arguments with ">" marker for each stream
	streamArgs := make([]string, 0, len(streamKeys)*2)
	for _, streamKey := range streamKeys {
		streamArgs = append(streamArgs, streamKey)
	}
	for range streamKeys {
		streamArgs = append(streamArgs, ">") // Read only new messages
	}

	// Read from multiple streams
	streams, err := c.redis.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.consumerGroup,
		Consumer: c.consumerID,
		Streams:  streamArgs,
		Count:    int64(c.batchSize),
		Block:    c.blockDuration,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			// No messages available - normal condition (this is expected most of the time)
			return nil
		}
		c.logger.WithError(err).Error("XReadGroup failed")
		return err
	}

	// Process messages from all streams
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			// Process message (may succeed, move to DLQ, or fail before DLQ)
			err := c.processMessage(ctx, stream.Stream, msg)

			// Determine if message should be acknowledged:
			// - Success (err == nil): Data in ClickHouse → Safe to ack
			// - In DLQ (ErrMovedToDLQ): Data preserved in DLQ → Safe to ack
			// - Failed before DLQ (other errors): No data preservation → Leave pending for retry
			shouldAck := err == nil || errors.Is(err, ErrMovedToDLQ)

			if shouldAck {
				// Acknowledge message - either processed successfully or safely in DLQ
				if ackErr := c.redis.Client.XAck(ctx, stream.Stream, c.consumerGroup, msg.ID).Err(); ackErr != nil {
					c.logger.WithError(ackErr).WithFields(logrus.Fields{
						"stream":     stream.Stream,
						"message_id": msg.ID,
					}).Warn("Failed to acknowledge message")
				}
			} else {
				// Leave message pending for retry (parse errors, DLQ write failures, etc.)
				c.logger.WithError(err).WithFields(logrus.Fields{
					"stream":     stream.Stream,
					"message_id": msg.ID,
				}).Error("Message processing failed - leaving pending for retry")
			}

			// Track errors for non-DLQ failures
			if err != nil && !errors.Is(err, ErrMovedToDLQ) {
				c.incrementErrors()
			}
		}
	}

	// Increment rotation for next read
	c.streamsMutex.Lock()
	c.streamRotation += c.maxStreamsPerRead
	c.streamsMutex.Unlock()

	return nil
}

// processMessage processes a single stream message
func (c *TelemetryStreamConsumer) processMessage(ctx context.Context, streamKey string, msg redis.XMessage) error {
	// Extract batch data from message
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		return errors.New("invalid message format: missing data field")
	}

	// Deserialize batch
	var batch streams.TelemetryStreamMessage
	if err := json.Unmarshal([]byte(dataStr), &batch); err != nil {
		return fmt.Errorf("failed to unmarshal batch data: %w", err)
	}

	// Process with retry logic
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * c.retryBackoff
			c.logger.WithFields(logrus.Fields{
				"batch_id": batch.BatchID.String(),
				"attempt":  attempt,
				"backoff":  backoff,
			}).Info("Retrying batch processing")
			time.Sleep(backoff)
		}

		if err := c.processBatch(ctx, &batch); err != nil {
			lastErr = err
			continue
		}

		// Success
		c.incrementStats(1, int64(len(batch.Events)))
		c.logger.WithFields(logrus.Fields{
			"batch_id":    batch.BatchID.String(),
			"project_id":  batch.ProjectID.String(),
			"event_count": len(batch.Events),
			"message_id":  msg.ID,
		}).Debug("Successfully processed batch from stream")
		return nil
	}

	// Max retries exceeded - move to DLQ
	if err := c.moveToDLQ(ctx, streamKey, msg, &batch, lastErr); err != nil {
		c.logger.WithError(err).Error("Failed to move message to DLQ")
		// DLQ write failed - keep claims held, message stays pending
		return fmt.Errorf("max retries exceeded AND failed to move to DLQ: %w", lastErr)
	}

	// DLQ write succeeded - release claims so client retries can proceed
	if len(batch.ClaimedSpanIDs) > 0 {
		if releaseErr := c.deduplicationSvc.ReleaseEvents(ctx, batch.ClaimedSpanIDs); releaseErr != nil {
			c.logger.WithError(releaseErr).WithFields(logrus.Fields{
				"batch_id":    batch.BatchID.String(),
				"event_count": len(batch.ClaimedSpanIDs),
			}).Error("Failed to release claims after DLQ write")
			// Don't fail - DLQ write succeeded, worst case: 24h TTL expires
		} else {
			c.logger.WithFields(logrus.Fields{
				"batch_id":    batch.BatchID.String(),
				"event_count": len(batch.ClaimedSpanIDs),
			}).Info("Released claims after moving batch to DLQ")
		}
	}

	// Successfully moved to DLQ and released claims
	return ErrMovedToDLQ
}

// sortEventsByDependency sorts events to ensure dependencies are processed first
// Order: trace → session → span → quality_score
// This prevents "parent not found" errors during batch processing
func (c *TelemetryStreamConsumer) sortEventsByDependency(events []streams.TelemetryEventData) []streams.TelemetryEventData {
	// Define processing priority (lower = processed first)
	eventPriority := map[observability.TelemetryEventType]int{
		observability.TelemetryEventTypeTrace:        1, // Traces first (parents)
		observability.TelemetryEventTypeSession:      2, // Sessions second
		observability.TelemetryEventTypeSpan:         3, // Spans third (require traces)
		observability.TelemetryEventTypeQualityScore: 4, // Scores last (require traces/spans)
	}

	// Create a copy to avoid modifying original slice
	sorted := make([]streams.TelemetryEventData, len(events))
	copy(sorted, events)

	// Stable sort by priority
	sort.SliceStable(sorted, func(i, j int) bool {
		typeI := observability.TelemetryEventType(sorted[i].EventType)
		typeJ := observability.TelemetryEventType(sorted[j].EventType)

		priorityI := eventPriority[typeI]
		priorityJ := eventPriority[typeJ]

		// Unknown types get lowest priority (processed last)
		if priorityI == 0 {
			priorityI = 999
		}
		if priorityJ == 0 {
			priorityJ = 999
		}

		return priorityI < priorityJ
	})

	return sorted
}

// safeExtractFromPayload safely extracts string from payload map (nil-safe)
func safeExtractFromPayload(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}

	val, ok := payload[key]
	if !ok {
		return ""
	}

	strVal, ok := val.(string)
	if !ok {
		return ""
	}

	return strVal
}

// processTraceEvent processes a trace event using TraceService
func (c *TelemetryStreamConsumer) processTraceEvent(ctx context.Context, eventData *streams.TelemetryEventData, projectID ulid.ULID) error {
	// Map event payload to Trace struct
	var trace observability.Trace
	if err := mapToStruct(eventData.EventPayload, &trace); err != nil {
		return fmt.Errorf("failed to unmarshal trace payload: %w", err)
	}

	// Set project_id from authentication context
	trace.ProjectID = projectID.String()

	// Use service layer (handles validation, business logic, and repository)
	if err := c.traceService.CreateTrace(ctx, &trace); err != nil {
		return fmt.Errorf("failed to create trace via service: %w", err)
	}

	return nil
}

// processSpanEvent processes an span event using SpanService
func (c *TelemetryStreamConsumer) processSpanEvent(ctx context.Context, eventData *streams.TelemetryEventData, projectID ulid.ULID) error {
	// Debug: Log payload structure to identify ULID format issues
	c.logger.WithFields(logrus.Fields{
		"event_id": eventData.EventID.String(),
		"payload_keys": func() []string {
			keys := make([]string, 0, len(eventData.EventPayload))
			for k := range eventData.EventPayload {
				keys = append(keys, k)
			}
			return keys
		}(),
	}).Debug("Processing span payload")

	// Map event payload to Span struct
	var span observability.Span
	if err := mapToStruct(eventData.EventPayload, &span); err != nil {
		// Log detailed error with sample payload data
		c.logger.WithFields(logrus.Fields{
			"event_id":     eventData.EventID.String(),
			"trace_id_raw": eventData.EventPayload["trace_id"],
			"id_raw":       eventData.EventPayload["id"],
			"error":        err.Error(),
		}).Error("Failed to unmarshal span - payload debug")
		return fmt.Errorf("failed to unmarshal span payload: %w", err)
	}

	// Set context from stream message
	span.ProjectID = projectID.String()

	// Use service layer
	if err := c.spanService.CreateSpan(ctx, &span); err != nil {
		return fmt.Errorf("failed to create span via service: %w", err)
	}

	return nil
}

// processScoreEvent processes a quality score event using ScoreService
func (c *TelemetryStreamConsumer) processScoreEvent(ctx context.Context, eventData *streams.TelemetryEventData, projectID ulid.ULID) error {
	// Map event payload to Score struct
	var score observability.Score
	if err := mapToStruct(eventData.EventPayload, &score); err != nil {
		return fmt.Errorf("failed to unmarshal score payload: %w", err)
	}

	// Set context from stream message
	score.ProjectID = projectID.String()

	// Use service layer
	if err := c.scoreService.CreateScore(ctx, &score); err != nil {
		return fmt.Errorf("failed to create score via service: %w", err)
	}

	return nil
}

// mapToStruct converts map[string]interface{} to a struct using JSON marshaling
// This is a type-safe way to convert event payloads to domain types
func mapToStruct(input map[string]interface{}, output interface{}) error {
	// Marshal map to JSON
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// Unmarshal JSON to struct
	if err := json.Unmarshal(jsonData, output); err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return nil
}

// processBatch processes a batch of telemetry events by routing to appropriate services based on event type
// This is the main orchestration method that decides how to handle each event
func (c *TelemetryStreamConsumer) processBatch(ctx context.Context, batch *streams.TelemetryStreamMessage) error {
	// batch.ProjectID is already ulid.ULID type
	projectID := batch.ProjectID

	// Track processing stats
	var (
		processedCount int
		failedCount    int
		lastError      error
	)

	// Sort events by dependency order: traces → sessions → spans → scores
	// This ensures parent entities exist before children are created
	sortedEvents := c.sortEventsByDependency(batch.Events)

	// Process each event based on its type
	for _, event := range sortedEvents {
		var err error

		// Route event to appropriate service based on event_type
		switch observability.TelemetryEventType(event.EventType) {
		case observability.TelemetryEventTypeTrace:
			// Structured trace event → TraceService → traces table
			err = c.processTraceEvent(ctx, &event, projectID)

		// Session events removed - sessions are now virtual groupings via session_id attribute

		case observability.TelemetryEventTypeSpan:
			// Structured span event → SpanService → spans table
			err = c.processSpanEvent(ctx, &event, projectID)

		case observability.TelemetryEventTypeQualityScore:
			// Structured score event → ScoreService → scores table
			err = c.processScoreEvent(ctx, &event, projectID)

		default:
			// Unknown event type - log warning and skip
			c.logger.WithFields(logrus.Fields{
				"event_id":   event.EventID.String(),
				"event_type": event.EventType,
				"batch_id":   batch.BatchID.String(),
			}).Warn("Unknown event type, skipping")
			failedCount++
			continue
		}

		if err != nil {
			// Log error but continue processing other events (partial success model)
			c.logger.WithError(err).WithFields(logrus.Fields{
				"event_id":   event.EventID.String(),
				"event_type": event.EventType,
				"batch_id":   batch.BatchID.String(),
			}).Error("Failed to process event")
			failedCount++
			lastError = err
			continue
		}

		processedCount++
	}

	// Determine success: At least one event processed successfully
	if processedCount == 0 && failedCount > 0 {
		return fmt.Errorf("batch processing failed: 0/%d events processed, last error: %w", len(batch.Events), lastError)
	}

	// Log partial failures
	if failedCount > 0 {
		c.logger.WithFields(logrus.Fields{
			"batch_id":        batch.BatchID.String(),
			"total_events":    len(batch.Events),
			"processed_count": processedCount,
			"failed_count":    failedCount,
		}).Warn("Batch processed with partial failures")
	}

	// ✅ Deduplication already handled synchronously in HTTP handler via ClaimEvents
	// Events are claimed atomically before publishing to stream, so no async registration needed
	// This eliminates the race condition where duplicate check happens before async registration completes

	return nil
}

// ptrTime is a helper to create a pointer to a time.Time value
func ptrTime(t time.Time) *time.Time {
	return &t
}

// serializeMetadata converts metadata map to JSON string
func (c *TelemetryStreamConsumer) serializeMetadata(metadata map[string]interface{}) string {
	if metadata == nil {
		return "{}"
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to serialize metadata")
		return "{}"
	}

	return string(data)
}

// incrementStats atomically increments processing statistics
func (c *TelemetryStreamConsumer) incrementStats(batches, events int64) {
	atomic.AddInt64(&c.batchesProcessed, batches)
	atomic.AddInt64(&c.eventsProcessed, events)
}

// incrementErrors atomically increments error count
func (c *TelemetryStreamConsumer) incrementErrors() {
	atomic.AddInt64(&c.errorsCount, 1)
}

// moveToDLQ moves a failed message to the Dead Letter Queue
func (c *TelemetryStreamConsumer) moveToDLQ(ctx context.Context, streamKey string, msg redis.XMessage, batch *streams.TelemetryStreamMessage, err error) error {
	dlqKey := fmt.Sprintf("%s:%s", dlqStreamPrefix, batch.ProjectID.String())

	// Serialize DLQ entry with error metadata
	dlqData := map[string]interface{}{
		"original_stream": streamKey,
		"original_msg_id": msg.ID,
		"batch_id":        batch.BatchID.String(),
		"project_id":      batch.ProjectID.String(),
		"event_count":     len(batch.Events),
		"error_message":   err.Error(),
		"failed_at":       time.Now().Unix(),
		"retry_count":     c.maxRetries,
		"original_data":   msg.Values["data"], // Preserve original message data
	}

	// Add to DLQ stream with trimming and TTL
	result, addErr := c.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqKey,
		MaxLen: dlqMaxLength, // Prevent unbounded growth
		Approx: true,
		Values: dlqData,
	}).Result()

	if addErr != nil {
		return fmt.Errorf("failed to add message to DLQ: %w", addErr)
	}

	// Set TTL on DLQ stream (7 days retention)
	if err := c.redis.Client.Expire(ctx, dlqKey, dlqRetentionPeriod).Err(); err != nil {
		c.logger.WithError(err).Warn("Failed to set DLQ TTL")
	}

	// Increment DLQ counter
	atomic.AddInt64(&c.dlqMessagesCount, 1)

	c.logger.WithFields(logrus.Fields{
		"dlq_id":      result,
		"dlq_key":     dlqKey,
		"batch_id":    batch.BatchID.String(),
		"project_id":  batch.ProjectID.String(),
		"error":       err.Error(),
		"retry_count": c.maxRetries,
	}).Warn("Moved failed batch to Dead Letter Queue")

	return nil
}

// GetStats returns current consumer statistics
func (c *TelemetryStreamConsumer) GetStats() map[string]int64 {
	c.streamsMutex.RLock()
	activeStreamCount := int64(len(c.activeStreams))
	c.streamsMutex.RUnlock()

	return map[string]int64{
		"batches_processed": atomic.LoadInt64(&c.batchesProcessed),
		"events_processed":  atomic.LoadInt64(&c.eventsProcessed),
		"errors_count":      atomic.LoadInt64(&c.errorsCount),
		"dlq_messages":      atomic.LoadInt64(&c.dlqMessagesCount),
		"active_streams":    activeStreamCount,
	}
}

// GetDLQMessages retrieves messages from the Dead Letter Queue for a project
func (c *TelemetryStreamConsumer) GetDLQMessages(ctx context.Context, projectID ulid.ULID, count int64) ([]redis.XMessage, error) {
	dlqKey := fmt.Sprintf("%s:%s", dlqStreamPrefix, projectID.String())

	// Read messages from DLQ
	messages, err := c.redis.Client.XRevRange(ctx, dlqKey, "+", "-").Result()
	if err != nil {
		if err == redis.Nil {
			return []redis.XMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read DLQ messages: %w", err)
	}

	// Limit results
	if count > 0 && int64(len(messages)) > count {
		messages = messages[:count]
	}

	return messages, nil
}

// RetryDLQMessage attempts to reprocess a message from the DLQ
func (c *TelemetryStreamConsumer) RetryDLQMessage(ctx context.Context, projectID ulid.ULID, messageID string) error {
	dlqKey := fmt.Sprintf("%s:%s", dlqStreamPrefix, projectID.String())

	// Read the message
	messages, err := c.redis.Client.XRange(ctx, dlqKey, messageID, messageID).Result()
	if err != nil || len(messages) == 0 {
		return fmt.Errorf("DLQ message not found: %s", messageID)
	}

	msg := messages[0]

	// Extract original data
	originalData, ok := msg.Values["original_data"].(string)
	if !ok {
		return errors.New("invalid DLQ message format: missing original_data")
	}

	// Deserialize batch
	var batch streams.TelemetryStreamMessage
	if err := json.Unmarshal([]byte(originalData), &batch); err != nil {
		return fmt.Errorf("failed to unmarshal DLQ batch data: %w", err)
	}

	// Attempt reprocessing
	if err := c.processBatch(ctx, &batch); err != nil {
		return fmt.Errorf("retry failed: %w", err)
	}

	// Remove from DLQ on success
	if err := c.redis.Client.XDel(ctx, dlqKey, messageID).Err(); err != nil {
		c.logger.WithError(err).Warn("Failed to remove message from DLQ after successful retry")
	}

	c.logger.WithFields(logrus.Fields{
		"message_id": messageID,
		"batch_id":   batch.BatchID.String(),
		"project_id": projectID.String(),
	}).Info("Successfully retried DLQ message")

	return nil
}
