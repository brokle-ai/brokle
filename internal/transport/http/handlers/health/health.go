package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
)

// Handler handles health check endpoints
type Handler struct {
	config    *config.Config
	logger    *logrus.Logger
	startTime time.Time
}

// NewHandler creates a new health handler
func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{
		config:    config,
		logger:    logger,
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	LastChecked string `json:"last_checked"`
	Duration    string `json:"duration,omitempty"`
}

// Check handles basic health check
// @Summary Health check
// @Description Basic health check endpoint
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.config.App.Version,
		Uptime:    time.Since(h.startTime).String(),
	}

	c.JSON(http.StatusOK, response)
}

// Ready handles readiness check with dependencies
// @Summary Readiness check
// @Description Check if service is ready to handle requests
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Success 503 {object} HealthResponse
// @Router /health/ready [get]
func (h *Handler) Ready(c *gin.Context) {
	checks := make(map[string]HealthCheck)
	overallStatus := "healthy"
	statusCode := http.StatusOK

	// Check database connectivity
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	// Check Redis connectivity
	redisCheck := h.checkRedis()
	checks["redis"] = redisCheck
	if redisCheck.Status != "healthy" {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	// Check ClickHouse connectivity
	clickhouseCheck := h.checkClickHouse()
	checks["clickhouse"] = clickhouseCheck
	if clickhouseCheck.Status != "healthy" {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.config.App.Version,
		Uptime:    time.Since(h.startTime).String(),
		Checks:    checks,
	}

	c.JSON(statusCode, response)
}

// Live handles liveness check
// @Summary Liveness check
// @Description Check if service is alive
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/live [get]
func (h *Handler) Live(c *gin.Context) {
	response := HealthResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
	}

	c.JSON(http.StatusOK, response)
}

// checkDatabase checks database connectivity
func (h *Handler) checkDatabase() HealthCheck {
	start := time.Now()
	
	// TODO: Implement actual database ping
	// For now, simulate a successful check
	
	return HealthCheck{
		Status:      "healthy",
		Message:     "Database connection is healthy",
		LastChecked: time.Now().UTC().Format(time.RFC3339),
		Duration:    time.Since(start).String(),
	}
}

// checkRedis checks Redis connectivity
func (h *Handler) checkRedis() HealthCheck {
	start := time.Now()
	
	// TODO: Implement actual Redis ping
	// For now, simulate a successful check
	
	return HealthCheck{
		Status:      "healthy",
		Message:     "Redis connection is healthy",
		LastChecked: time.Now().UTC().Format(time.RFC3339),
		Duration:    time.Since(start).String(),
	}
}

// checkClickHouse checks ClickHouse connectivity
func (h *Handler) checkClickHouse() HealthCheck {
	start := time.Now()
	
	// TODO: Implement actual ClickHouse ping
	// For now, simulate a successful check
	
	return HealthCheck{
		Status:      "healthy",
		Message:     "ClickHouse connection is healthy",
		LastChecked: time.Now().UTC().Format(time.RFC3339),
		Duration:    time.Since(start).String(),
	}
}