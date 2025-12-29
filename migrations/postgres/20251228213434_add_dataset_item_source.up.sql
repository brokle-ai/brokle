-- Migration: add_dataset_item_source
-- Created: 2025-12-28T21:34:34+05:30
-- Purpose: Add source tracking, trace linking, and deduplication to dataset_items

-- Add source column to track item origin
ALTER TABLE dataset_items ADD COLUMN source VARCHAR(20) NOT NULL DEFAULT 'manual';

-- Add source trace/span linking for OTEL-native workflow
ALTER TABLE dataset_items ADD COLUMN source_trace_id VARCHAR(32);
ALTER TABLE dataset_items ADD COLUMN source_span_id VARCHAR(16);

-- Add content hash for deduplication
ALTER TABLE dataset_items ADD COLUMN content_hash VARCHAR(64);

-- Index for efficient deduplication lookups
CREATE INDEX idx_dataset_items_content_hash ON dataset_items(dataset_id, content_hash)
WHERE content_hash IS NOT NULL;

-- Index for source trace queries
CREATE INDEX idx_dataset_items_source_trace ON dataset_items(source_trace_id)
WHERE source_trace_id IS NOT NULL;

-- Add check constraint for valid source values
ALTER TABLE dataset_items ADD CONSTRAINT chk_dataset_items_source
CHECK (source IN ('manual', 'trace', 'span', 'csv', 'json', 'sdk'));

COMMENT ON COLUMN dataset_items.source IS 'Origin of the dataset item: manual, trace, span, csv, json, sdk';
COMMENT ON COLUMN dataset_items.source_trace_id IS 'Trace ID if item was created from production trace data';
COMMENT ON COLUMN dataset_items.source_span_id IS 'Span ID if item was created from production span data';
COMMENT ON COLUMN dataset_items.content_hash IS 'SHA256 hash of input+expected for deduplication';
