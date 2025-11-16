package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/storage"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/preview"
	"brokle/pkg/ulid"
)

// BlobStorageService implements business logic for blob storage management
type BlobStorageService struct {
	blobRepo observability.BlobStorageRepository
	s3Client *storage.S3Client
	config   *config.BlobStorageConfig
	logger   *logrus.Logger
}

// NewBlobStorageService creates a new blob storage service instance
func NewBlobStorageService(
	blobRepo observability.BlobStorageRepository,
	s3Client *storage.S3Client,
	cfg *config.BlobStorageConfig,
	logger *logrus.Logger,
) *BlobStorageService {
	return &BlobStorageService{
		blobRepo: blobRepo,
		s3Client: s3Client,
		config:   cfg,
		logger:   logger,
	}
}

// CreateBlobReference creates a new blob storage reference
func (s *BlobStorageService) CreateBlobReference(ctx context.Context, blob *observability.BlobStorageFileLog) error {
	// Validate required fields
	if blob.ProjectID == "" {
		return appErrors.NewValidationError("project_id is required", "blob must have a valid project_id")
	}
	if blob.EntityType == "" {
		return appErrors.NewValidationError("entity_type is required", "blob must have an entity_type")
	}
	if blob.EntityID == "" {
		return appErrors.NewValidationError("entity_id is required", "blob must have an entity_id")
	}
	if blob.BucketName == "" {
		return appErrors.NewValidationError("bucket_name is required", "blob must have a bucket_name")
	}
	if blob.BucketPath == "" {
		return appErrors.NewValidationError("bucket_path is required", "blob must have a bucket_path")
	}

	// Generate new ID if not provided
	if blob.ID == "" {
		blob.ID = ulid.New().String()
	}

	// Generate event ID if not provided
	if blob.EventID == "" {
		blob.EventID = ulid.New().String()
	}

	// Set timestamps
	if blob.CreatedAt.IsZero() {
		blob.CreatedAt = time.Now()
	}
	blob.UpdatedAt = time.Now()

	// Create blob reference
	if err := s.blobRepo.Create(ctx, blob); err != nil {
		return appErrors.NewInternalError("failed to create blob reference", err)
	}

	return nil
}

// UpdateBlobReference updates an existing blob storage reference
func (s *BlobStorageService) UpdateBlobReference(ctx context.Context, blob *observability.BlobStorageFileLog) error {
	// Validate blob exists
	existing, err := s.blobRepo.GetByID(ctx, blob.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("blob " + blob.ID)
		}
		return appErrors.NewInternalError("failed to get blob", err)
	}

	// Merge fields
	mergeBlobFields(existing, blob)

	// Update blob
	if err := s.blobRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update blob reference", err)
	}

	return nil
}

// DeleteBlobReference soft deletes a blob storage reference
func (s *BlobStorageService) DeleteBlobReference(ctx context.Context, id string) error {
	// Get blob to retrieve S3 path
	blob, err := s.blobRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("blob " + id)
		}
		return appErrors.NewInternalError("failed to get blob", err)
	}

	// Delete from S3 first
	if s.s3Client != nil {
		if err := s.s3Client.Delete(ctx, blob.BucketPath); err != nil {
			// Log error but continue to delete reference
			// S3 deletion is best-effort
			s.logger.WithError(err).Warn("Failed to delete from S3, continuing with reference deletion")
		}
	}

	// Delete blob reference from ClickHouse
	if err := s.blobRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete blob reference", err)
	}

	return nil
}

// GetBlobByID retrieves a blob storage reference by ID
func (s *BlobStorageService) GetBlobByID(ctx context.Context, id string) (*observability.BlobStorageFileLog, error) {
	blob, err := s.blobRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("blob " + id)
		}
		return nil, appErrors.NewInternalError("failed to get blob", err)
	}

	return blob, nil
}

// GetBlobsByEntityID retrieves all blob references for an entity
func (s *BlobStorageService) GetBlobsByEntityID(ctx context.Context, entityType, entityID string) ([]*observability.BlobStorageFileLog, error) {
	blobs, err := s.blobRepo.GetByEntityID(ctx, entityType, entityID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get blobs by entity", err)
	}

	return blobs, nil
}

// GetBlobsByProjectID retrieves blobs by project ID with optional filters
func (s *BlobStorageService) GetBlobsByProjectID(ctx context.Context, projectID string, filter *observability.BlobStorageFilter) ([]*observability.BlobStorageFileLog, error) {
	blobs, err := s.blobRepo.GetByProjectID(ctx, projectID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get blobs by project", err)
	}

	return blobs, nil
}

// ShouldOffload determines if content should be offloaded to S3 based on size
// Threshold is configured via BLOB_STORAGE_THRESHOLD env var (default: 10KB)
func (s *BlobStorageService) ShouldOffload(content string) bool {
	// Config always has default from viper.SetDefault (10_000 bytes = 10KB)
	return len(content) > s.config.Threshold
}

// UploadToS3 uploads content to S3 and creates a blob reference
func (s *BlobStorageService) UploadToS3(ctx context.Context, content string, projectID, entityType, entityID, eventID string) (*observability.BlobStorageFileLog, error) {
	// Check if S3 client is initialized
	if s.s3Client == nil {
		return nil, errors.New("S3 client not initialized - check BLOB_STORAGE configuration in environment")
	}

	// Generate blob ID
	blobID := ulid.New().String()

	// Generate S3 key: {entity_type}/{entity_id}/{blob_id}.json
	s3Key := fmt.Sprintf("%s/%s/%s.json", entityType, entityID, blobID)

	// Upload to S3
	contentBytes := []byte(content)
	if err := s.s3Client.Upload(ctx, s3Key, contentBytes, "application/json"); err != nil {
		return nil, appErrors.NewInternalError("failed to upload to S3", err)
	}

	// Create blob reference
	blob := &observability.BlobStorageFileLog{
		ID:         blobID,
		ProjectID:  projectID,
		EntityType: entityType,
		EntityID:   entityID,
		EventID:    eventID,
		BucketName: s.config.BucketName,
		BucketPath: s3Key,
		FileSizeBytes: func() *uint64 {
			size := uint64(len(contentBytes))
			return &size
		}(),
		ContentType: func() *string {
			ct := "application/json"
			return &ct
		}(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store blob reference in ClickHouse
	if err := s.CreateBlobReference(ctx, blob); err != nil {
		// Try to cleanup S3 object on ClickHouse failure
		_ = s.s3Client.Delete(ctx, s3Key)
		return nil, err
	}

	return blob, nil
}

// UploadToS3WithPreview uploads content to S3 and returns blob info + preview
// This combines S3 upload with type-aware preview generation for efficiency
func (s *BlobStorageService) UploadToS3WithPreview(ctx context.Context, content string, projectID, entityType, entityID, eventID string) (*observability.BlobStorageFileLog, string, error) {
	// Upload to S3
	blob, err := s.UploadToS3(ctx, content, projectID, entityType, entityID, eventID)
	if err != nil {
		return nil, "", err
	}

	// Generate type-aware preview
	previewText := preview.GeneratePreview(content)

	return blob, previewText, nil
}

// DownloadFromS3 downloads content from S3 using blob reference
func (s *BlobStorageService) DownloadFromS3(ctx context.Context, blobID string) (string, error) {
	// Check if S3 client is initialized
	if s.s3Client == nil {
		return "", errors.New("S3 client not initialized - check BLOB_STORAGE configuration in environment")
	}

	// Get blob reference from ClickHouse
	blob, err := s.blobRepo.GetByID(ctx, blobID)
	if err != nil {
		return "", appErrors.NewNotFoundError("blob " + blobID)
	}

	// Download from S3
	contentBytes, err := s.s3Client.Download(ctx, blob.BucketPath)
	if err != nil {
		return "", appErrors.NewInternalError("failed to download from S3", err)
	}

	return string(contentBytes), nil
}

// CountBlobs returns the count of blob references matching the filter
func (s *BlobStorageService) CountBlobs(ctx context.Context, filter *observability.BlobStorageFilter) (int64, error) {
	count, err := s.blobRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count blobs", err)
	}

	return count, nil
}

// Helper function to merge blob fields
func mergeBlobFields(dst *observability.BlobStorageFileLog, src *observability.BlobStorageFileLog) {
	// Update mutable fields
	if src.FileSizeBytes != nil {
		dst.FileSizeBytes = src.FileSizeBytes
	}
	if src.ContentType != nil {
		dst.ContentType = src.ContentType
	}
	if src.Compression != nil {
		dst.Compression = src.Compression
	}
}
