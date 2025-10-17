package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"brokle/internal/infrastructure/database"
	"brokle/pkg/ulid"
)

// TelemetryStreamMessage represents a telemetry batch message in Redis Stream
type TelemetryStreamMessage struct {
	BatchID     ulid.ULID              `json:"batch_id"`
	ProjectID   ulid.ULID              `json:"project_id"`
	Environment string                 `json:"environment,omitempty"`
	Events      []TelemetryEventData   `json:"events"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// TelemetryEventData represents individual event data in the stream message
type TelemetryEventData struct {
	EventID      ulid.ULID              `json:"event_id"`
	EventType    string                 `json:"event_type"`
	EventPayload map[string]interface{} `json:"event_payload"`
}

// TelemetryStreamProducer handles publishing telemetry data to Redis Streams
type TelemetryStreamProducer struct {
	redis  *database.RedisDB
	logger *logrus.Logger
}

// NewTelemetryStreamProducer creates a new telemetry stream producer
func NewTelemetryStreamProducer(redis *database.RedisDB, logger *logrus.Logger) *TelemetryStreamProducer {
	return &TelemetryStreamProducer{
		redis:  redis,
		logger: logger,
	}
}

// PublishBatch publishes a telemetry batch to Redis Stream
// Returns the stream message ID for tracking
func (p *TelemetryStreamProducer) PublishBatch(ctx context.Context, batch *TelemetryStreamMessage) (string, error) {
	if batch == nil {
		return "", fmt.Errorf("batch cannot be nil")
	}

	if batch.BatchID.IsZero() {
		return "", fmt.Errorf("batch ID is required")
	}

	if batch.ProjectID.IsZero() {
		return "", fmt.Errorf("project ID is required")
	}

	// Use project-specific stream for better distribution and scalability
	streamKey := fmt.Sprintf("telemetry:batches:%s", batch.ProjectID.String())

	// Serialize batch data to JSON
	eventData, err := json.Marshal(batch)
	if err != nil {
		return "", fmt.Errorf("failed to marshal batch data: %w", err)
	}

	// Add to Redis Stream without MaxLen to prevent data loss
	// Stream cleanup is handled by TTL (30 days) which safely expires after all messages processed
	// MaxLen would trim unprocessed pending messages during high load or consumer outages
	result, err := p.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"batch_id":    batch.BatchID.String(),
			"project_id":  batch.ProjectID.String(),
			"environment": batch.Environment,
			"event_count": len(batch.Events),
			"data":        string(eventData),
			"timestamp":   batch.Timestamp.Unix(),
		},
	}).Result()

	if err != nil {
		return "", fmt.Errorf("failed to add batch to stream: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"stream_id":   result,
		"batch_id":    batch.BatchID.String(),
		"project_id":  batch.ProjectID.String(),
		"event_count": len(batch.Events),
		"stream_key":  streamKey,
	}).Debug("Published telemetry batch to Redis Stream")

	// Set stream TTL for GDPR compliance (30 days)
	if err := p.SetStreamTTL(ctx, batch.ProjectID, 30*24*time.Hour); err != nil {
		p.logger.WithError(err).Warn("Failed to set stream TTL (GDPR compliance)")
		// Don't return error - TTL is best-effort for compliance
	}

	return result, nil
}

// GetStreamInfo retrieves information about a stream
func (p *TelemetryStreamProducer) GetStreamInfo(ctx context.Context, projectID ulid.ULID) (*redis.XInfoStream, error) {
	streamKey := fmt.Sprintf("telemetry:batches:%s", projectID.String())

	info, err := p.redis.Client.XInfoStream(ctx, streamKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("stream not found for project %s", projectID.String())
		}
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	return info, nil
}

// GetStreamLength returns the number of messages in a stream
func (p *TelemetryStreamProducer) GetStreamLength(ctx context.Context, projectID ulid.ULID) (int64, error) {
	streamKey := fmt.Sprintf("telemetry:batches:%s", projectID.String())

	length, err := p.redis.Client.XLen(ctx, streamKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get stream length: %w", err)
	}

	return length, nil
}

// SetStreamTTL sets a TTL on the stream for automatic expiration (GDPR compliance)
// Default: 30 days (720 hours) for GDPR Article 17 compliance (right to erasure)
func (p *TelemetryStreamProducer) SetStreamTTL(ctx context.Context, projectID ulid.ULID, ttl time.Duration) error {
	streamKey := fmt.Sprintf("telemetry:batches:%s", projectID.String())

	// Set TTL on stream key
	err := p.redis.Client.Expire(ctx, streamKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set stream TTL: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"project_id": projectID.String(),
		"stream_key": streamKey,
		"ttl_hours":  ttl.Hours(),
	}).Debug("Set stream TTL for GDPR compliance")

	return nil
}

// DeleteStream removes a stream (use with caution)
func (p *TelemetryStreamProducer) DeleteStream(ctx context.Context, projectID ulid.ULID) error {
	streamKey := fmt.Sprintf("telemetry:batches:%s", projectID.String())

	err := p.redis.Client.Del(ctx, streamKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"project_id": projectID.String(),
		"stream_key": streamKey,
	}).Warn("Deleted telemetry stream")

	return nil
}
