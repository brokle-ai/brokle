-- Migration: add_prompt_indexes
-- Created: 2025-12-13T08:07:34+05:30

-- ===================================
-- ADD PROMPT MANAGEMENT PERFORMANCE INDEXES
-- ===================================
-- Adds indexes for better query performance per design doc (db-schema.md)

-- Full-text search on prompts (name + description)
-- Enables fast text search with tsvector
CREATE INDEX IF NOT EXISTS idx_prompts_search
  ON prompts USING GIN (
    to_tsvector('english', name || ' ' || COALESCE(description, ''))
  )
  WHERE deleted_at IS NULL;

-- JSONB search indexes on prompt_versions
-- Enables efficient filtering by template content
CREATE INDEX IF NOT EXISTS idx_prompt_versions_template
  ON prompt_versions USING GIN (template);

-- Enables efficient filtering by model configuration
CREATE INDEX IF NOT EXISTS idx_prompt_versions_config
  ON prompt_versions USING GIN (config);

-- Enables efficient variable lookups
CREATE INDEX IF NOT EXISTS idx_prompt_versions_variables
  ON prompt_versions USING GIN (variables);
