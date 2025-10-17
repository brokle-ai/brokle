package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	redis               *database.RedisDB
	clickhouseRepo      observability.TelemetryAnalyticsRepository
	deduplicationSvc    observability.TelemetryDeduplicationService
	logger              *logrus.Logger
	consumerGroup       string
	consumerID          string
	batchSize           int
	blockDuration       time.Duration
	maxRetries          int
	retryBackoff        time.Duration
	discoveryInterval   time.Duration
	maxStreamsPerRead   int
	running             int64
	wg                  sync.WaitGroup
	quit                chan struct{}
	statsLock           sync.RWMutex
	batchesProcessed    int64
	eventsProcessed     int64
	errorsCount         int64
	dlqMessagesCount    int64
	activeStreams       map[string]bool
	streamsMutex        sync.RWMutex
	streamRotation      int
	discoveryBackoff    time.Duration
	maxDiscoveryBackoff time.Duration
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
	clickhouseRepo observability.TelemetryAnalyticsRepository,
	deduplicationSvc observability.TelemetryDeduplicationService,
	logger *logrus.Logger,
	config *TelemetryStreamConsumerConfig,
) *TelemetryStreamConsumer {
	if config == nil {
		config = &TelemetryStreamConsumerConfig{
			ConsumerGroup:     "telemetry-workers",
			ConsumerID:        fmt.Sprintf("worker-%s", ulid.New().String()),
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
		clickhouseRepo:      clickhouseRepo,
		deduplicationSvc:    deduplicationSvc,
		logger:              logger,
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
		return fmt.Errorf("consumer already running")
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
		err := c.redis.Client.XGroupCreateMkStream(ctx, streamKey, c.consumerGroup, "$").Err()
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
			// No messages available - normal condition
			return nil
		}
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
		return fmt.Errorf("invalid message format: missing data field")
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
		// DLQ write failed - return error so message stays pending for retry
		return fmt.Errorf("max retries exceeded AND failed to move to DLQ: %w", lastErr)
	}

	// Successfully moved to DLQ - return sentinel so message can be safely acknowledged
	return ErrMovedToDLQ
}

// processBatch writes batch data to ClickHouse using domain types
func (c *TelemetryStreamConsumer) processBatch(ctx context.Context, batch *streams.TelemetryStreamMessage) error {
	// Convert stream events to domain TelemetryEvent types WITH context from stream message
	domainEvents := make([]*observability.TelemetryEvent, len(batch.Events))
	for i, event := range batch.Events {
		domainEvents[i] = &observability.TelemetryEvent{
			ID:           event.EventID,
			BatchID:      batch.BatchID,
			ProjectID:    batch.ProjectID,    // From stream message
			Environment:  batch.Environment,  // From stream message
			EventType:    observability.TelemetryEventType(event.EventType),
			EventPayload: event.EventPayload,
			CreatedAt:    batch.Timestamp,
			RetryCount:   0,
			ProcessedAt:  ptrTime(time.Now()),
		}
	}

	// Batch insert events to ClickHouse using domain types (context carried in structs)
	if err := c.clickhouseRepo.InsertTelemetryEventsBatch(ctx, domainEvents); err != nil {
		return fmt.Errorf("failed to insert telemetry events: %w", err)
	}

	// Create batch record in ClickHouse using domain types WITH context from stream message
	processingTimeMs := 0 // TODO: Calculate actual processing time
	domainBatch := &observability.TelemetryBatch{
		ID:               batch.BatchID,
		ProjectID:        batch.ProjectID,
		Environment:      batch.Environment,  // From stream message
		BatchMetadata:    batch.Metadata,
		TotalEvents:      len(batch.Events),
		ProcessedEvents:  len(batch.Events),
		FailedEvents:     0,
		Status:           observability.BatchStatusCompleted,
		ProcessingTimeMs: &processingTimeMs,
		CreatedAt:        batch.Timestamp,
		CompletedAt:      ptrTime(time.Now()),
	}

	// Insert batch (context carried in domain struct)
	if err := c.clickhouseRepo.InsertTelemetryBatch(ctx, domainBatch); err != nil {
		return fmt.Errorf("failed to insert telemetry batch: %w", err)
	}

	// ✅ Register events as processed ONLY after successful ClickHouse persistence
	// This prevents data loss if the worker fails before processing completes
	eventIDs := make([]ulid.ULID, len(batch.Events))
	for i, event := range batch.Events {
		eventIDs[i] = event.EventID
	}

	if err := c.deduplicationSvc.RegisterProcessedEventsBatch(ctx, batch.ProjectID, batch.BatchID, eventIDs); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"batch_id":    batch.BatchID.String(),
			"project_id":  batch.ProjectID.String(),
			"event_count": len(eventIDs),
		}).Warn("Failed to register events for deduplication - duplicates may be accepted on retry")
		// Don't fail the batch - data is already safely persisted in ClickHouse
		// Worst case: duplicate events might be accepted if SDK retries
	}

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
		"original_stream":  streamKey,
		"original_msg_id":  msg.ID,
		"batch_id":         batch.BatchID.String(),
		"project_id":       batch.ProjectID.String(),
		"environment":      batch.Environment,
		"event_count":      len(batch.Events),
		"error_message":    err.Error(),
		"failed_at":        time.Now().Unix(),
		"retry_count":      c.maxRetries,
		"original_data":    msg.Values["data"], // Preserve original message data
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
		"dlq_id":        result,
		"dlq_key":       dlqKey,
		"batch_id":      batch.BatchID.String(),
		"project_id":    batch.ProjectID.String(),
		"error":         err.Error(),
		"retry_count":   c.maxRetries,
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
		return fmt.Errorf("invalid DLQ message format: missing original_data")
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
