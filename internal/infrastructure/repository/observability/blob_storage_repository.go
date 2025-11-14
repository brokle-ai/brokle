package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type blobStorageRepository struct {
	db clickhouse.Conn
}

// NewBlobStorageRepository creates a new blob storage repository instance
func NewBlobStorageRepository(db clickhouse.Conn) observability.BlobStorageRepository {
	return &blobStorageRepository{db: db}
}

// Create inserts a new blob storage reference into ClickHouse
func (r *blobStorageRepository) Create(ctx context.Context, blob *observability.BlobStorageFileLog) error {
	// Set version and event_ts for new blob references
	// Version is now optional application version
	blob.UpdatedAt = time.Now()

	query := `
		INSERT INTO blob_storage_file_log (
			id, project_id, entity_type, entity_id, event_id,
			bucket_name, bucket_path,
			file_size_bytes, content_type, compression,
			created_at, updated_at,
			version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		blob.ID,
		blob.ProjectID,
		blob.EntityType,
		blob.EntityID,
		blob.EventID,
		blob.BucketName,
		blob.BucketPath,
		blob.FileSizeBytes,
		blob.ContentType,
		blob.Compression,
		blob.CreatedAt,
		blob.UpdatedAt,
		blob.Version,
		// Removed: event_ts, is_deleted
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *blobStorageRepository) Update(ctx context.Context, blob *observability.BlobStorageFileLog) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	// Version is now optional application version (not auto-incremented)
	blob.UpdatedAt = time.Now()

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, blob)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *blobStorageRepository) Delete(ctx context.Context, id string) error {
	query := `
		INSERT INTO blob_storage_file_log
		SELECT
			id, project_id, entity_type, entity_id, event_id,
			bucket_name, bucket_path,
			file_size_bytes, content_type, compression,
			created_at, updated_at,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM blob_storage_file_log
		WHERE id = ?		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a blob storage reference by its ID (returns latest version)
func (r *blobStorageRepository) GetByID(ctx context.Context, id string) (*observability.BlobStorageFileLog, error) {
	query := `
		SELECT
			id, project_id, entity_type, entity_id, event_id,
			bucket_name, bucket_path,
			file_size_bytes, content_type, compression,
			created_at, updated_at,
			version
		FROM blob_storage_file_log
		WHERE id = ?		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanBlobRow(row)
}

// GetByEntityID retrieves all blob storage references for an entity
func (r *blobStorageRepository) GetByEntityID(ctx context.Context, entityType, entityID string) ([]*observability.BlobStorageFileLog, error) {
	query := `
		SELECT
			id, project_id, entity_type, entity_id, event_id,
			bucket_name, bucket_path,
			file_size_bytes, content_type, compression,
			created_at, updated_at,
			version
		FROM blob_storage_file_log
		WHERE entity_type = ? AND entity_id = ?		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("query blobs by entity: %w", err)
	}
	defer rows.Close()

	return r.scanBlobs(rows)
}

// GetByProjectID retrieves blob storage references by project ID with optional filters
func (r *blobStorageRepository) GetByProjectID(ctx context.Context, projectID string, filter *observability.BlobStorageFilter) ([]*observability.BlobStorageFileLog, error) {
	query := `
		SELECT
			id, project_id, entity_type, entity_id, event_id,
			bucket_name, bucket_path,
			file_size_bytes, content_type, compression,
			created_at, updated_at,
			version
		FROM blob_storage_file_log
		WHERE project_id = ?	`

	args := []interface{}{projectID}

	// Apply filters
	if filter != nil {
		if filter.EntityType != nil {
			query += " AND entity_type = ?"
			args = append(args, *filter.EntityType)
		}
		if filter.StartTime != nil {
			query += " AND created_at >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND created_at <= ?"
			args = append(args, *filter.EndTime)
		}
		if filter.MinSizeBytes != nil {
			query += " AND file_size_bytes >= ?"
			args = append(args, *filter.MinSizeBytes)
		}
		if filter.MaxSizeBytes != nil {
			query += " AND file_size_bytes <= ?"
			args = append(args, *filter.MaxSizeBytes)
		}
	}

	// Order by created_at descending (most recent first)
	query += " ORDER BY created_at DESC"

	// Apply limit and offset
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query blobs by project: %w", err)
	}
	defer rows.Close()

	return r.scanBlobs(rows)
}

// Count returns the count of blob storage references matching the filter
func (r *blobStorageRepository) Count(ctx context.Context, filter *observability.BlobStorageFilter) (int64, error) {
	query := "SELECT count() FROM blob_storage_file_log WHERE is_deleted = 0"
	args := []interface{}{}

	if filter != nil {
		if filter.EntityType != nil {
			query += " AND entity_type = ?"
			args = append(args, *filter.EntityType)
		}
		if filter.StartTime != nil {
			query += " AND created_at >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND created_at <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a single blob from query row
func (r *blobStorageRepository) scanBlobRow(row driver.Row) (*observability.BlobStorageFileLog, error) {
	var blob observability.BlobStorageFileLog

	err := row.Scan(
		&blob.ID,
		&blob.ProjectID,
		&blob.EntityType,
		&blob.EntityID,
		&blob.EventID,
		&blob.BucketName,
		&blob.BucketPath,
		&blob.FileSizeBytes,
		&blob.ContentType,
		&blob.Compression,
		&blob.CreatedAt,
		&blob.UpdatedAt,
		&blob.Version,
		// Removed: event_ts, is_deleted
	)

	if err != nil {
		return nil, fmt.Errorf("scan blob: %w", err)
	}

	return &blob, nil
}

// Helper function to scan multiple blobs from rows
func (r *blobStorageRepository) scanBlobs(rows driver.Rows) ([]*observability.BlobStorageFileLog, error) {
	var blobs []*observability.BlobStorageFileLog

	for rows.Next() {
		var blob observability.BlobStorageFileLog

		err := rows.Scan(
			&blob.ID,
			&blob.ProjectID,
			&blob.EntityType,
			&blob.EntityID,
			&blob.EventID,
			&blob.BucketName,
			&blob.BucketPath,
			&blob.FileSizeBytes,
			&blob.ContentType,
			&blob.Compression,
			&blob.CreatedAt,
			&blob.UpdatedAt,
			&blob.Version,
			// Removed: event_ts, is_deleted
		)

		if err != nil {
			return nil, fmt.Errorf("scan blob row: %w", err)
		}

		blobs = append(blobs, &blob)
	}

	return blobs, rows.Err()
}
