package observability

import (
	"brokle/internal/core/domain/observability"
)

// ServiceRegistry holds all observability services
type ServiceRegistry struct {
	TraceService       observability.TraceService
	ObservationService observability.ObservationService
	QualityService     observability.QualityService
}

// NewServiceRegistry creates a new service registry with all observability services
func NewServiceRegistry(
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	qualityScoreRepo observability.QualityScoreRepository,
	eventPublisher observability.EventPublisher,
) *ServiceRegistry {
	// Create trace service
	traceService := NewTraceService(traceRepo, observationRepo, eventPublisher)

	// Create observation service
	observationService := NewObservationService(observationRepo, traceRepo, eventPublisher)

	// Create quality service
	qualityService := NewQualityService(qualityScoreRepo, traceRepo, observationRepo, eventPublisher)

	return &ServiceRegistry{
		TraceService:       traceService,
		ObservationService: observationService,
		QualityService:     qualityService,
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