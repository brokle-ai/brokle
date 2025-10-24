package observability

import (
	"brokle/internal/config"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/storage"
	"github.com/sirupsen/logrus"
)

// ServiceRegistry holds all observability services
type ServiceRegistry struct {
	// ClickHouse-first services
	TraceService       *TraceService
	ObservationService *ObservationService
	ScoreService       *ScoreService
	BlobStorageService *BlobStorageService

	// OTLP conversion service
	OTLPConverterService *OTLPConverterService

	// Existing telemetry service (Redis Streams + async processing)
	TelemetryService observability.TelemetryService
}

// NewServiceRegistry creates a new service registry with all observability services
func NewServiceRegistry(
	// ClickHouse repositories
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	scoreRepo observability.ScoreRepository,
	blobStorageRepo observability.BlobStorageRepository,

	// Blob storage dependencies
	s3Client *storage.S3Client,
	blobConfig *config.BlobStorageConfig,

	// Telemetry system (keep existing)
	telemetryService observability.TelemetryService,

	logger *logrus.Logger,
) *ServiceRegistry {
	// Create blob storage service first (needed by trace and observation services)
	blobStorageService := NewBlobStorageService(blobStorageRepo, s3Client, blobConfig, logger)

	// Create OTLP converter service
	otlpConverterService := NewOTLPConverterService(logger)

	// Create ClickHouse-first services with blob storage dependency
	traceService := NewTraceService(traceRepo, observationRepo, scoreRepo, blobStorageService, logger)
	observationService := NewObservationService(observationRepo, traceRepo, scoreRepo, blobStorageService, logger)
	scoreService := NewScoreService(scoreRepo, traceRepo, observationRepo)

	return &ServiceRegistry{
		TraceService:         traceService,
		ObservationService:   observationService,
		ScoreService:         scoreService,
		BlobStorageService:   blobStorageService,
		OTLPConverterService: otlpConverterService,
		TelemetryService:     telemetryService,
	}
}

// GetTraceService returns the trace service
func (r *ServiceRegistry) GetTraceService() *TraceService {
	return r.TraceService
}

// GetObservationService returns the observation service
func (r *ServiceRegistry) GetObservationService() *ObservationService {
	return r.ObservationService
}

// GetScoreService returns the score service
func (r *ServiceRegistry) GetScoreService() *ScoreService {
	return r.ScoreService
}

// GetBlobStorageService returns the blob storage service
func (r *ServiceRegistry) GetBlobStorageService() *BlobStorageService {
	return r.BlobStorageService
}

// GetOTLPConverterService returns the OTLP converter service
func (r *ServiceRegistry) GetOTLPConverterService() *OTLPConverterService {
	return r.OTLPConverterService
}

// GetTelemetryService returns the telemetry service
func (r *ServiceRegistry) GetTelemetryService() observability.TelemetryService {
	return r.TelemetryService
}
