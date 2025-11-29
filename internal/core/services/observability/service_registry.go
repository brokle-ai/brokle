package observability

import (
	"brokle/internal/config"
	"brokle/internal/core/domain/analytics"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/storage"
	"brokle/internal/infrastructure/streams"

	"github.com/sirupsen/logrus"
)

// ServiceRegistry holds all observability services
type ServiceRegistry struct {
	// ClickHouse-first services
	TraceService       *TraceService
	SpanService        *SpanService
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

// NewServiceRegistry creates a new service registry with all observability services
func NewServiceRegistry(
	// ClickHouse repositories
	traceRepo observability.TraceRepository,
	spanRepo observability.SpanRepository,
	scoreRepo observability.ScoreRepository,
	metricsRepo observability.MetricsRepository,
	logsRepo observability.LogsRepository,
	genaiEventsRepo observability.GenAIEventsRepository,
	blobStorageRepo observability.BlobStorageRepository,

	// Blob storage dependencies
	s3Client *storage.S3Client,
	blobConfig *config.BlobStorageConfig,

	// Stream infrastructure (for OTLP)
	streamProducer *streams.TelemetryStreamProducer,
	deduplicationService observability.TelemetryDeduplicationService,

	// Telemetry system (keep existing)
	telemetryService observability.TelemetryService,

	// Pricing service (NEW)
	providerPricingService analytics.ProviderPricingService,

	// Observability configuration
	observabilityConfig *config.ObservabilityConfig,

	logger *logrus.Logger,
) *ServiceRegistry {
	// Create blob storage service (kept for future use: exports, media files, raw events)
	blobStorageService := NewBlobStorageService(blobStorageRepo, s3Client, blobConfig, logger)

	// Create OTLP converter services (with provider pricing service for cost analytics)
	otlpConverterService := NewOTLPConverterService(logger, providerPricingService, observabilityConfig)
	otlpMetricsConverterService := NewOTLPMetricsConverterService(logger)
	otlpLogsConverterService := NewOTLPLogsConverterService(logger)
	otlpEventsConverterService := NewOTLPEventsConverterService(logger)

	// Create ClickHouse-first services (no blob storage for LLM data - stored inline with ZSTD compression)
	traceService := NewTraceService(traceRepo, spanRepo, logger)
	spanService := NewSpanService(spanRepo, traceRepo, scoreRepo, logger)
	scoreService := NewScoreService(scoreRepo, traceRepo, spanRepo)
	metricsService := NewMetricsService(metricsRepo, logger)
	logsService := NewLogsService(logsRepo, logger)
	genaiEventsService := NewGenAIEventsService(genaiEventsRepo, logger)

	return &ServiceRegistry{
		TraceService:                traceService,
		SpanService:                 spanService,
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

// GetTraceService returns the trace service
func (r *ServiceRegistry) GetTraceService() *TraceService {
	return r.TraceService
}

// GetSpanService returns the span service
func (r *ServiceRegistry) GetSpanService() *SpanService {
	return r.SpanService
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
