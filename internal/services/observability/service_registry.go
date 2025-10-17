package observability

import (
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/streams"
	"brokle/internal/workers"
	"github.com/sirupsen/logrus"
)

// ServiceRegistry holds all observability services
type ServiceRegistry struct {
	TraceService       observability.TraceService
	ObservationService observability.ObservationService
	QualityService     observability.QualityService
	TelemetryService   observability.TelemetryService
}

// NewServiceRegistry creates a new service registry with all observability services
func NewServiceRegistry(
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	qualityScoreRepo observability.QualityScoreRepository,
	eventPublisher observability.EventPublisher,
	// Telemetry - Redis-only (no PostgreSQL)
	telemetryDeduplicationRepo observability.TelemetryDeduplicationRepository,
	// Redis Streams producer
	streamProducer *streams.TelemetryStreamProducer,
	// Analytics worker
	analyticsWorker *workers.TelemetryAnalyticsWorker,
	logger *logrus.Logger,
) *ServiceRegistry {
	// Create trace service
	traceService := NewTraceService(traceRepo, observationRepo, eventPublisher)

	// Create observation service
	observationService := NewObservationService(observationRepo, traceRepo, eventPublisher)

	// Create quality service
	qualityService := NewQualityService(qualityScoreRepo, traceRepo, observationRepo, eventPublisher)

	// Create deduplication service (Redis-only)
	telemetryDeduplicationService := NewTelemetryDeduplicationService(telemetryDeduplicationRepo)

	// Create main telemetry service with Redis Streams support (no PostgreSQL)
	telemetryService := NewTelemetryService(
		telemetryDeduplicationService,
		streamProducer,
		analyticsWorker,
		logger,
	)

	return &ServiceRegistry{
		TraceService:       traceService,
		ObservationService: observationService,
		QualityService:     qualityService,
		TelemetryService:   telemetryService,
	}
}

// GetTraceService returns the trace service
func (r *ServiceRegistry) GetTraceService() observability.TraceService {
	return r.TraceService
}

// GetObservationService returns the observation service
func (r *ServiceRegistry) GetObservationService() observability.ObservationService {
	return r.ObservationService
}

// GetQualityService returns the quality service
func (r *ServiceRegistry) GetQualityService() observability.QualityService {
	return r.QualityService
}

// GetTelemetryService returns the telemetry service
func (r *ServiceRegistry) GetTelemetryService() observability.TelemetryService {
	return r.TelemetryService
}