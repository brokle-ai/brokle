-- PostgreSQL Migration Rollback: create_filter_presets
-- Created: 2026-01-01
-- Purpose: Remove filter_presets table and related objects

DROP TRIGGER IF EXISTS trigger_filter_presets_updated_at ON filter_presets;
DROP FUNCTION IF EXISTS update_filter_presets_updated_at();
DROP TABLE IF EXISTS filter_presets;
