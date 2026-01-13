-- PostgreSQL Migration: create_annotation_queues (DOWN)
-- Drops annotation queues tables and indexes

-- Drop indexes first (some are partial indexes, drop explicitly)
DROP INDEX IF EXISTS idx_annotation_queue_assignments_queue;
DROP INDEX IF EXISTS idx_annotation_queue_assignments_user;
DROP INDEX IF EXISTS idx_annotation_queue_items_object;
DROP INDEX IF EXISTS idx_annotation_queue_items_priority;
DROP INDEX IF EXISTS idx_annotation_queue_items_locked;
DROP INDEX IF EXISTS idx_annotation_queue_items_queue_status;
DROP INDEX IF EXISTS idx_annotation_queues_project_created;
DROP INDEX IF EXISTS idx_annotation_queues_project_status;

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS annotation_queue_assignments;
DROP TABLE IF EXISTS annotation_queue_items;
DROP TABLE IF EXISTS annotation_queues;
