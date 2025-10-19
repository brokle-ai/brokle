package observability

import (
	"brokle/internal/core/domain/observability"
	"github.com/sirupsen/logrus"
)

// ServiceRegistry holds all observability services
type ServiceRegistry struct {
	// New ClickHouse-first services
	TraceService       *TraceService
	ObservationService *ObservationService
	ScoreService       *ScoreService
	SessionService     *SessionService

	// Existing telemetry service (Redis Streams + async processing)
	TelemetryService observability.TelemetryService
}

// NewServiceRegistry creates a new service registry with all observability services
func NewServiceRegistry(
	// ClickHouse repositories
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	scoreRepo observability.ScoreRepository,
	sessionRepo observability.SessionRepository,

	// Telemetry system (keep existing)
	telemetryService observability.TelemetryService,

	logger *logrus.Logger,
) *ServiceRegistry {
	// Create new ClickHouse-first services
	traceService := NewTraceService(traceRepo, observationRepo, scoreRepo)
	observationService := NewObservationService(observationRepo, traceRepo, scoreRepo)
	scoreService := NewScoreService(scoreRepo, traceRepo, observationRepo, sessionRepo)
	sessionService := NewSessionService(sessionRepo, traceRepo, scoreRepo)

	return &ServiceRegistry{
		TraceService:       traceService,
		ObservationService: observationService,
		ScoreService:       scoreService,
		SessionService:     sessionService,
		TelemetryService:   telemetryService,
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

// GetSessionService returns the session service
func (r *ServiceRegistry) GetSessionService() *SessionService {
	return r.SessionService
}

// GetTelemetryService returns the telemetry service
func (r *ServiceRegistry) GetTelemetryService() observability.TelemetryService {
	return r.TelemetryService
}
