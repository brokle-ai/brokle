-- Migration: add_prompt_constraints
-- Created: 2025-12-13T08:06:25+05:30

-- ===================================
-- ADD PROMPT MANAGEMENT CHECK CONSTRAINTS
-- ===================================
-- Adds validation constraints to ensure data integrity per design doc (db-schema.md)

-- Prompts table: name format validation
-- Ensures prompt names start with a letter and contain only alphanumeric, underscore, or hyphen
ALTER TABLE prompts
  ADD CONSTRAINT prompts_name_format
  CHECK (name ~ '^[a-zA-Z][a-zA-Z0-9_-]*$');

-- Prompt versions: version number must be positive
ALTER TABLE prompt_versions
  ADD CONSTRAINT prompt_versions_version_positive
  CHECK (version > 0);

-- Prompt versions: commit message length limit
ALTER TABLE prompt_versions
  ADD CONSTRAINT prompt_versions_commit_message_length
  CHECK (commit_message IS NULL OR char_length(commit_message) <= 500);

-- Prompt labels: name format validation (lowercase with dots, dashes, underscores)
ALTER TABLE prompt_labels
  ADD CONSTRAINT prompt_labels_name_format
  CHECK (name ~ '^[a-z0-9_.-]+$');

-- Prompt labels: name length between 1 and 36 characters
ALTER TABLE prompt_labels
  ADD CONSTRAINT prompt_labels_name_length
  CHECK (char_length(name) BETWEEN 1 AND 36);

-- Protected labels: format validation
ALTER TABLE prompt_protected_labels
  ADD CONSTRAINT prompt_protected_labels_format
  CHECK (label_name ~ '^[a-z0-9_.-]+$');
