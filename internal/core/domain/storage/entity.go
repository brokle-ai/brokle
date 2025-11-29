package storage

import "time"

// BlobStorageFileLog represents a reference to S3-stored data
type BlobStorageFileLog struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	FileSizeBytes *uint64   `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	Compression   *string   `json:"compression,omitempty" db:"compression"`
	ContentType   *string   `json:"content_type,omitempty" db:"content_type"`
	EntityID      string    `json:"entity_id" db:"entity_id"`
	BucketPath    string    `json:"bucket_path" db:"bucket_path"`
	BucketName    string    `json:"bucket_name" db:"bucket_name"`
	EventID       string    `json:"event_id" db:"event_id"`
	ID            string    `json:"id" db:"id"`
	EntityType    string    `json:"entity_type" db:"entity_type"`
	ProjectID     string    `json:"project_id" db:"project_id"`
}

// GetS3URI returns the full S3 URI for this blob
func (b *BlobStorageFileLog) GetS3URI() string {
	return "s3://" + b.BucketName + "/" + b.BucketPath
}

// IsCompressed returns true if the blob has compression applied
func (b *BlobStorageFileLog) IsCompressed() bool {
	return b.Compression != nil && *b.Compression != ""
}
