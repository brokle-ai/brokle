package observability

import (
	"brokle/internal/config"
	"brokle/internal/core/domain/analytics"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/storage"
	"brokle/internal/infrastructure/streams"

	"github.com/sirupsen/logrus"
)

type ServiceRegistry struct {
	TraceService       *TraceService
	ScoreService       *ScoreService
	MetricsService     *MetricsService
	LogsService        *LogsService
	GenAIEventsService *GenAIEventsService
	BlobStorageService *BlobStorageService

	// OTLP conversion services
	OTLPConverterService        *OTLPConverterService
	OTLPMetricsConverterService *OTLPMetricsConverterService
	OTLPLogsConverterService    *OTLPLogsConverterService
	OTLPEventsConverterService  *OTLPEventsConverterService

	// Stream infrastructure (for OTLP direct access)
	StreamProducer       *streams.TelemetryStreamProducer
	DeduplicationService observability.TelemetryDeduplicationService

	// Existing telemetry service (Redis Streams + async processing)
	TelemetryService observability.TelemetryService
}

func NewServiceRegistry(
	traceRepo observability.TraceRepository,
	scoreRepo observability.ScoreRepository,
	metricsRepo observability.MetricsRepository,
	logsRepo observability.LogsRepository,
	genaiEventsRepo observability.GenAIEventsRepository,
	blobStorageRepo observability.BlobStorageRepository,

	s3Client *storage.S3Client,
	blobConfig *config.BlobStorageConfig,

	streamProducer *streams.TelemetryStreamProducer,
	deduplicationService observability.TelemetryDeduplicationService,

	telemetryService observability.TelemetryService,
	providerPricingService analytics.ProviderPricingService,
	observabilityConfig *config.ObservabilityConfig,

	logger *logrus.Logger,
) *ServiceRegistry {
	blobStorageService := NewBlobStorageService(blobStorageRepo, s3Client, blobConfig, logger)
	otlpConverterService := NewOTLPConverterService(logger, providerPricingService, observabilityConfig)
	otlpMetricsConverterService := NewOTLPMetricsConverterService(logger)
	otlpLogsConverterService := NewOTLPLogsConverterService(logger)
	otlpEventsConverterService := NewOTLPEventsConverterService(logger)
	traceService := NewTraceService(traceRepo, logger)
	scoreService := NewScoreService(scoreRepo, traceRepo)
	metricsService := NewMetricsService(metricsRepo, logger)
	logsService := NewLogsService(logsRepo, logger)
	genaiEventsService := NewGenAIEventsService(genaiEventsRepo, logger)

	return &ServiceRegistry{
		TraceService:                traceService,
		ScoreService:                scoreService,
		MetricsService:              metricsService,
		LogsService:                 logsService,
		GenAIEventsService:          genaiEventsService,
		BlobStorageService:          blobStorageService,
		OTLPConverterService:        otlpConverterService,
		OTLPMetricsConverterService: otlpMetricsConverterService,
		OTLPLogsConverterService:    otlpLogsConverterService,
		OTLPEventsConverterService:  otlpEventsConverterService,
		StreamProducer:              streamProducer,
		DeduplicationService:        deduplicationService,
		TelemetryService:            telemetryService,
	}
}

func (r *ServiceRegistry) GetTraceService() *TraceService {
	return r.TraceService
}

func (r *ServiceRegistry) GetScoreService() *ScoreService {
	return r.ScoreService
}

func (r *ServiceRegistry) GetBlobStorageService() *BlobStorageService {
	return r.BlobStorageService
}

func (r *ServiceRegistry) GetOTLPConverterService() *OTLPConverterService {
	return r.OTLPConverterService
}

func (r *ServiceRegistry) GetTelemetryService() observability.TelemetryService {
	return r.TelemetryService
}
