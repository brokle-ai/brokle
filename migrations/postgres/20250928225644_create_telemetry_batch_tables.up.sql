-- Migration: create_telemetry_batch_tables
-- Created: 2025-09-28T22:56:44+05:30
--
-- Create telemetry batch processing tables for ULID-optimized bulk endpoint
-- Following Brokle patterns with proper indexing for 10k events/sec throughput

-- Create telemetry batches table for batch processing tracking
CREATE TABLE telemetry_batches (
    id CHAR(26) PRIMARY KEY,
    project_id CHAR(26) NOT NULL,
    batch_metadata JSONB NOT NULL DEFAULT '{}',
    total_events INTEGER NOT NULL DEFAULT 0 CHECK (total_events >= 0),
    processed_events INTEGER NOT NULL DEFAULT 0 CHECK (processed_events >= 0),
    failed_events INTEGER NOT NULL DEFAULT 0 CHECK (failed_events >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed', 'partial')),
    processing_time_ms INTEGER NULL CHECK (processing_time_ms >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create telemetry events table for event envelope pattern
CREATE TABLE telemetry_events (
    id CHAR(26) PRIMARY KEY,
    batch_id CHAR(26) NOT NULL REFERENCES telemetry_batches(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'trace_create', 'trace_update', 'observation_create',
        'observation_update', 'observation_complete', 'quality_score_create'
    )),
    event_payload JSONB NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NULL,
    error_message TEXT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Create telemetry event deduplication table with ULID-based TTL
CREATE TABLE telemetry_event_deduplication (
    event_id CHAR(26) PRIMARY KEY,
    batch_id CHAR(26) NOT NULL,
    project_id CHAR(26) NOT NULL,
    first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create optimized indexes for performance (separate migration files for CONCURRENTLY)
-- Basic indexes that can be created normally
CREATE INDEX idx_telemetry_batches_project_id ON telemetry_batches(project_id);
CREATE INDEX idx_telemetry_batches_status ON telemetry_batches(status);
CREATE INDEX idx_telemetry_batches_created_at ON telemetry_batches(created_at DESC);

CREATE INDEX idx_telemetry_events_batch_id ON telemetry_events(batch_id);
CREATE INDEX idx_telemetry_events_event_type ON telemetry_events(event_type);
CREATE INDEX idx_telemetry_events_created_at ON telemetry_events(created_at DESC);

CREATE INDEX idx_telemetry_dedup_expires_at ON telemetry_event_deduplication(expires_at);
CREATE INDEX idx_telemetry_dedup_project_id ON telemetry_event_deduplication(project_id);

-- GIN indexes for JSONB columns
CREATE INDEX idx_telemetry_batches_metadata ON telemetry_batches USING GIN (batch_metadata);
CREATE INDEX idx_telemetry_events_payload ON telemetry_events USING GIN (event_payload);

-- Conditional indexes for performance optimization
CREATE INDEX idx_telemetry_batches_processing
ON telemetry_batches(created_at DESC)
WHERE status IN ('processing', 'failed');

CREATE INDEX idx_telemetry_events_unprocessed
ON telemetry_events(created_at DESC)
WHERE processed_at IS NULL;

CREATE INDEX idx_telemetry_events_failed
ON telemetry_events(retry_count, created_at DESC)
WHERE retry_count > 0;

-- Create function for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_telemetry_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.completed_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic completion timestamp
CREATE TRIGGER trigger_telemetry_batches_completion
    BEFORE UPDATE ON telemetry_batches
    FOR EACH ROW
    WHEN (OLD.status != NEW.status AND NEW.status IN ('completed', 'failed', 'partial'))
    EXECUTE FUNCTION update_telemetry_updated_at();

-- Add comments for documentation
COMMENT ON TABLE telemetry_batches IS 'Batch processing tracking for telemetry events with ULID optimization';
COMMENT ON TABLE telemetry_events IS 'Individual telemetry events within batches using envelope pattern';
COMMENT ON TABLE telemetry_event_deduplication IS 'ULID-based event deduplication with smart TTL';

COMMENT ON COLUMN telemetry_batches.batch_metadata IS 'SDK metadata including name, version, environment, and custom fields';
COMMENT ON COLUMN telemetry_batches.processing_time_ms IS 'Total processing time in milliseconds for performance tracking';
COMMENT ON COLUMN telemetry_events.event_payload IS 'Complete event data for processing by respective services';
COMMENT ON COLUMN telemetry_event_deduplication.expires_at IS 'ULID timestamp-based expiry for efficient cleanup';

