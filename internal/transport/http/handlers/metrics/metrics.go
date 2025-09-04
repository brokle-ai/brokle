package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
)

// Handler handles metrics endpoints
type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

// NewHandler creates a new metrics handler
func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{
		config: config,
		logger: logger,
	}
}

// Handler handles Prometheus metrics endpoint
func (h *Handler) Handler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}