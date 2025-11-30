package storage

import "context"

// BlobStorageService defines the interface for blob storage business operations
type BlobStorageService interface {
	CreateBlobReference(ctx context.Context, blob *BlobStorageFileLog) error
	GetBlobByID(ctx context.Context, id string) (*BlobStorageFileLog, error)
	GetBlobsByEntityID(ctx context.Context, entityType, entityID string) ([]*BlobStorageFileLog, error)
	GetBlobsByProjectID(ctx context.Context, projectID string, filter *BlobStorageFilter) ([]*BlobStorageFileLog, error)
	UpdateBlobReference(ctx context.Context, blob *BlobStorageFileLog) error
	DeleteBlobReference(ctx context.Context, id string) error
	ShouldOffload(content string) bool
	UploadToS3(ctx context.Context, content string, projectID, entityType, entityID, eventID string) (*BlobStorageFileLog, error)
	UploadToS3WithPreview(ctx context.Context, content string, projectID, entityType, entityID, eventID string) (*BlobStorageFileLog, string, error)
	DownloadFromS3(ctx context.Context, blobID string) (string, error)
	CountBlobs(ctx context.Context, filter *BlobStorageFilter) (int64, error)
}
