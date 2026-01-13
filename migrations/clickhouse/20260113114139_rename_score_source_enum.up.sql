-- ClickHouse Migration: rename_score_source_enum
-- Created: 2026-01-13T11:41:39+05:30

-- Rename score source enum values to match industry patterns
-- Old: code=1, llm=2, human=3
-- New: api=1, eval=2, annotation=3
-- ClickHouse allows modifying enum values while preserving data (same numeric IDs)
ALTER TABLE scores MODIFY COLUMN source Enum8('api' = 1, 'eval' = 2, 'annotation' = 3);
