-- Rollback: add_dataset_item_source
-- Created: 2025-12-28T21:34:34+05:30

-- Remove constraint first
ALTER TABLE dataset_items DROP CONSTRAINT IF EXISTS chk_dataset_items_source;

-- Remove indexes
DROP INDEX IF EXISTS idx_dataset_items_source_trace;
DROP INDEX IF EXISTS idx_dataset_items_content_hash;

-- Remove columns
ALTER TABLE dataset_items DROP COLUMN IF EXISTS content_hash;
ALTER TABLE dataset_items DROP COLUMN IF EXISTS source_span_id;
ALTER TABLE dataset_items DROP COLUMN IF EXISTS source_trace_id;
ALTER TABLE dataset_items DROP COLUMN IF EXISTS source;
