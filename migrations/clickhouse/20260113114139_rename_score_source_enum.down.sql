-- ClickHouse Rollback: rename_score_source_enum
-- Created: 2026-01-13T11:41:39+05:30

-- Revert to original enum values
ALTER TABLE scores MODIFY COLUMN source Enum8('code' = 1, 'llm' = 2, 'human' = 3);
