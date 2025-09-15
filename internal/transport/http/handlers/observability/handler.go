package observability

import (
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	obsServices "brokle/internal/services/observability"
)

// Handler contains all observability-related HTTP handlers
type Handler struct {
	config   *config.Config
	logger   *logrus.Logger
	services *obsServices.ServiceRegistry
}

// NewHandler creates a new observability handler
func NewHandler(
	cfg *config.Config,
	logger *logrus.Logger,
	services *obsServices.ServiceRegistry,
) *Handler {
	return &Handler{
		config:   cfg,
		logger:   logger,
		services: services,
	}
}