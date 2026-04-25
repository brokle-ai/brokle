package app

import (
	"log/slog"

	"brokle/internal/config"
	observabilityService "brokle/internal/core/services/observability"
	storageService "brokle/internal/core/services/storage"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/storage"
	"brokle/internal/infrastructure/streams"
)

// ProvideObservabilityServices builds the observability ServiceRegistry —
// the shared bag of trace / score / span / OTLP-converter services
// consumed by both the HTTP plane and the worker plane.
//
// Wire ordering matters: deduplication + stream-producer are built
// first; telemetry service depends on both; S3 client + blob storage
// are optional and degrade to disabled when configuration is incomplete.
func ProvideObservabilityServices(
	observabilityRepos *ObservabilityRepositories,
	storageRepos *StorageRepositories,
	analyticsServices *AnalyticsServices,
	redisDB *database.RedisDB,
	cfg *config.Config,
	logger *slog.Logger,
) *observabilityService.ServiceRegistry {
	deduplicationService := observabilityService.NewTelemetryDeduplicationService(observabilityRepos.TelemetryDeduplication)
	streamProducer := streams.NewTelemetryStreamProducer(redisDB, logger)
	telemetryService := observabilityService.NewTelemetryService(
		deduplicationService,
		streamProducer,
		logger,
	)

	var s3Client *storage.S3Client
	if cfg.BlobStorage.Provider != "" && cfg.BlobStorage.BucketName != "" {
		var err error
		s3Client, err = storage.NewS3Client(&cfg.BlobStorage, logger)
		if err != nil {
			logger.Warn("Failed to initialize S3 client, blob storage will be disabled", "error", err)
		}
	}

	blobStorageSvc := storageService.NewBlobStorageService(
		storageRepos.BlobStorage,
		s3Client,
		&cfg.BlobStorage,
		logger,
	)

	return observabilityService.NewServiceRegistry(
		observabilityRepos.Trace,
		observabilityRepos.Score,
		observabilityRepos.ScoreAnalytics,
		observabilityRepos.Metrics,
		observabilityRepos.Logs,
		observabilityRepos.GenAIEvents,
		observabilityRepos.FilterPreset,
		blobStorageSvc,
		s3Client,
		&cfg.Archive, // Archive config for S3 raw telemetry archival
		streamProducer,
		deduplicationService,
		telemetryService,
		analyticsServices.ProviderPricing,
		&cfg.Observability,
		logger,
	)
}
