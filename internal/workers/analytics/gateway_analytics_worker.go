package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/pkg/ulid"
)

// GatewayAnalyticsWorker processes gateway usage data for analytics and billing
type GatewayAnalyticsWorker struct {
	logger              *logrus.Logger
	batchSize           int
	flushInterval       time.Duration
	analyticsRepository AnalyticsRepository
	billingService      BillingService
	
	// Internal state
	mutex         sync.RWMutex
	requestBuffer []*RequestMetric
	usageBuffer   []*UsageMetric
	costBuffer    []*CostMetric
	running       bool
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// RequestMetric represents a processed gateway request
type RequestMetric struct {
	ID             ulid.ULID                  `json:"id"`
	RequestID      ulid.ULID                  `json:"request_id"`
	OrganizationID ulid.ULID                  `json:"organization_id"`
	UserID         *ulid.ULID                 `json:"user_id,omitempty"`
	ProviderID     ulid.ULID                  `json:"provider_id"`
	ProviderName   string                     `json:"provider_name"`
	ModelID        ulid.ULID                  `json:"model_id"`
	ModelName      string                     `json:"model_name"`
	RequestType    gateway.RequestType        `json:"request_type"`
	Method         string                     `json:"method"`
	Endpoint       string                     `json:"endpoint"`
	Status         string                     `json:"status"`
	StatusCode     int                        `json:"status_code"`
	Duration       time.Duration              `json:"duration"`
	InputTokens    int32                      `json:"input_tokens"`
	OutputTokens   int32                      `json:"output_tokens"`
	TotalTokens    int32                      `json:"total_tokens"`
	EstimatedCost  float64                    `json:"estimated_cost"`
	ActualCost     float64                    `json:"actual_cost"`
	Currency       string                     `json:"currency"`
	RoutingReason  string                     `json:"routing_reason"`
	CacheHit       bool                       `json:"cache_hit"`
	Error          string                     `json:"error,omitempty"`
	Metadata       map[string]interface{}     `json:"metadata,omitempty"`
	Timestamp      time.Time                  `json:"timestamp"`
}

// UsageMetric represents aggregated usage data
type UsageMetric struct {
	ID               ulid.ULID           `json:"id"`
	OrganizationID   ulid.ULID           `json:"organization_id"`
	ProviderID       ulid.ULID           `json:"provider_id"`
	ModelID          ulid.ULID           `json:"model_id"`
	RequestType      gateway.RequestType `json:"request_type"`
	Period           string              `json:"period"` // hourly, daily, monthly
	PeriodStart      time.Time           `json:"period_start"`
	PeriodEnd        time.Time           `json:"period_end"`
	RequestCount     int64               `json:"request_count"`
	SuccessCount     int64               `json:"success_count"`
	ErrorCount       int64               `json:"error_count"`
	TotalInputTokens int64               `json:"total_input_tokens"`
	TotalOutputTokens int64              `json:"total_output_tokens"`
	TotalTokens      int64               `json:"total_tokens"`
	TotalCost        float64             `json:"total_cost"`
	Currency         string              `json:"currency"`
	AvgDuration      float64             `json:"avg_duration"`
	MinDuration      time.Duration       `json:"min_duration"`
	MaxDuration      time.Duration       `json:"max_duration"`
	CacheHitRate     float64             `json:"cache_hit_rate"`
	Timestamp        time.Time           `json:"timestamp"`
}

// CostMetric represents cost tracking data
type CostMetric struct {
	ID               ulid.ULID           `json:"id"`
	RequestID        ulid.ULID           `json:"request_id"`
	OrganizationID   ulid.ULID           `json:"organization_id"`
	ProviderID       ulid.ULID           `json:"provider_id"`
	ModelID          ulid.ULID           `json:"model_id"`
	RequestType      gateway.RequestType `json:"request_type"`
	InputTokens      int32               `json:"input_tokens"`
	OutputTokens     int32               `json:"output_tokens"`
	TotalTokens      int32               `json:"total_tokens"`
	InputCost        float64             `json:"input_cost"`
	OutputCost       float64             `json:"output_cost"`
	TotalCost        float64             `json:"total_cost"`
	EstimatedCost    float64             `json:"estimated_cost"`
	CostDifference   float64             `json:"cost_difference"`
	Currency         string              `json:"currency"`
	BillingTier      string              `json:"billing_tier"`
	DiscountApplied  float64             `json:"discount_applied"`
	Timestamp        time.Time           `json:"timestamp"`
}

// AnalyticsRepository defines the interface for storing analytics data
type AnalyticsRepository interface {
	BatchInsertRequestMetrics(ctx context.Context, metrics []*RequestMetric) error
	BatchInsertUsageMetrics(ctx context.Context, metrics []*UsageMetric) error
	BatchInsertCostMetrics(ctx context.Context, metrics []*CostMetric) error
	GetUsageStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*UsageMetric, error)
	GetCostStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*CostMetric, error)
}

// BillingService defines the interface for billing operations
type BillingService interface {
	RecordUsage(ctx context.Context, usage *CostMetric) error
	CalculateBill(ctx context.Context, orgID ulid.ULID, period string) (*BillingSummary, error)
	GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*BillingRecord, error)
}

// BillingSummary represents a billing period summary
type BillingSummary struct {
	ID                ulid.ULID           `json:"id"`
	OrganizationID    ulid.ULID           `json:"organization_id"`
	Period            string              `json:"period"`
	PeriodStart       time.Time           `json:"period_start"`
	PeriodEnd         time.Time           `json:"period_end"`
	TotalRequests     int64               `json:"total_requests"`
	TotalTokens       int64               `json:"total_tokens"`
	TotalCost         float64             `json:"total_cost"`
	Currency          string              `json:"currency"`
	ProviderBreakdown map[string]float64  `json:"provider_breakdown"`
	ModelBreakdown    map[string]float64  `json:"model_breakdown"`
	Discounts         float64             `json:"discounts"`
	NetCost           float64             `json:"net_cost"`
	Status            string              `json:"status"`
	GeneratedAt       time.Time           `json:"generated_at"`
}

// BillingRecord represents a billing transaction
type BillingRecord struct {
	ID               ulid.ULID  `json:"id"`
	OrganizationID   ulid.ULID  `json:"organization_id"`
	Period           string     `json:"period"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	TransactionID    *string    `json:"transaction_id,omitempty"`
	PaymentMethod    *string    `json:"payment_method,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
}

// WorkerConfig holds configuration for the analytics worker
type WorkerConfig struct {
	BatchSize     int
	FlushInterval time.Duration
	BufferSize    int
	WorkerCount   int
}

// DefaultWorkerConfig returns default configuration
func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		BatchSize:     1000,
		FlushInterval: 30 * time.Second,
		BufferSize:    10000,
		WorkerCount:   3,
	}
}

// NewGatewayAnalyticsWorker creates a new analytics worker instance
func NewGatewayAnalyticsWorker(
	logger *logrus.Logger,
	config *WorkerConfig,
	analyticsRepo AnalyticsRepository,
	billingService BillingService,
) *GatewayAnalyticsWorker {
	if config == nil {
		config = DefaultWorkerConfig()
	}

	return &GatewayAnalyticsWorker{
		logger:              logger,
		batchSize:           config.BatchSize,
		flushInterval:       config.FlushInterval,
		analyticsRepository: analyticsRepo,
		billingService:      billingService,
		requestBuffer:       make([]*RequestMetric, 0, config.BufferSize),
		usageBuffer:         make([]*UsageMetric, 0, config.BufferSize),
		costBuffer:          make([]*CostMetric, 0, config.BufferSize),
		stopCh:              make(chan struct{}),
	}
}

// Start starts the analytics worker without context
func (w *GatewayAnalyticsWorker) Start() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.running {
		w.logger.Warn("analytics worker is already running")
		return
	}

	w.running = true
	
	// Use background context for workers
	ctx := context.Background()
	
	// Start flush timer
	w.wg.Add(1)
	go w.flushWorker(ctx)

	// Start data processing workers
	w.wg.Add(1)
	go w.processWorker(ctx)

	w.logger.Info("Gateway analytics worker started")
}

// StartWithContext starts the analytics worker with context
func (w *GatewayAnalyticsWorker) StartWithContext(ctx context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.running {
		return fmt.Errorf("analytics worker is already running")
	}

	w.running = true
	
	// Start flush timer
	w.wg.Add(1)
	go w.flushWorker(ctx)

	// Start data processing workers
	w.wg.Add(1)
	go w.processWorker(ctx)

	w.logger.Info("Gateway analytics worker started")
	return nil
}

// Stop gracefully stops the analytics worker
func (w *GatewayAnalyticsWorker) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	close(w.stopCh)
	
	// Wait for workers to finish
	w.wg.Wait()

	// Flush remaining data
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	w.flushAllBuffers(ctx)

	w.logger.Info("Gateway analytics worker stopped")
	return nil
}

// RecordRequest records a gateway request for analytics
func (w *GatewayAnalyticsWorker) RecordRequest(ctx context.Context, metric *gateway.RequestMetrics) error {
	if !w.running {
		return fmt.Errorf("analytics worker is not running")
	}

	// Parse RequestID from string to ULID
	requestID, err := ulid.Parse(metric.RequestID)
	if err != nil {
		w.logger.WithError(err).WithField("request_id", metric.RequestID).Error("Failed to parse request ID")
		requestID = ulid.New() // Use new ULID if parsing fails
	}

	// Transform to analytics metric
	requestMetric := &RequestMetric{
		ID:             ulid.New(),
		RequestID:      requestID,
		OrganizationID: metric.ProjectID, // Using ProjectID as OrganizationID
		// UserID is not available in RequestMetrics, set to nil
		// ProviderID needs to be derived from provider name - use new ULID for now
		// TODO: Add provider lookup service to get ProviderID from provider name
		ProviderName:   metric.Provider,
		// ModelID needs to be derived from model name - use new ULID for now  
		// TODO: Add model lookup service to get ModelID from model name
		ModelName:      metric.Model,
		// RequestType is not available in current RequestMetrics - infer from context
		Method:         "POST", // Most AI requests are POST
		Endpoint:       "/v1/chat/completions", // Default endpoint - TODO: get actual endpoint
		Status:         func() string { if metric.Success { return "success" } else { return "error" } }(),
		StatusCode:     func() int { if metric.Success { return 200 } else { return 500 } }(),
		Duration:       time.Duration(metric.LatencyMs) * time.Millisecond,
		InputTokens:    int32(metric.InputTokens),
		OutputTokens:   int32(metric.OutputTokens),
		TotalTokens:    int32(metric.TotalTokens),
		EstimatedCost:  metric.CostUSD,
		ActualCost:     metric.CostUSD,
		Currency:       "USD",
		RoutingReason:  metric.RoutingStrategy,
		CacheHit:       metric.CacheHit,
		Error:          func() string { if metric.ErrorMessage != nil { return *metric.ErrorMessage } else { return "" } }(),
		Timestamp:      metric.Timestamp,
	}

	// Set default placeholder values for missing IDs (to be resolved by lookup services)
	providerID := ulid.New() // TODO: lookup actual provider ID
	modelID := ulid.New()    // TODO: lookup actual model ID

	// Set provider and model IDs on request metric
	requestMetric.ProviderID = providerID
	requestMetric.ModelID = modelID
	requestMetric.RequestType = gateway.RequestTypeChatCompletion // Default, TODO: infer from endpoint

	// Buffer the metric
	w.mutex.Lock()
	w.requestBuffer = append(w.requestBuffer, requestMetric)
	shouldFlush := len(w.requestBuffer) >= w.batchSize
	w.mutex.Unlock()

	// Create cost metric for billing (using available cost data)
	costMetric := &CostMetric{
		ID:               ulid.New(),
		RequestID:        requestID,
		OrganizationID:   metric.ProjectID,
		ProviderID:       providerID,
		ModelID:          modelID,
		RequestType:      gateway.RequestTypeChatCompletion, // Default
		InputTokens:      int32(metric.InputTokens),
		OutputTokens:     int32(metric.OutputTokens),
		TotalTokens:      int32(metric.TotalTokens),
		InputCost:        metric.CostUSD * 0.5,  // Estimate 50% for input cost
		OutputCost:       metric.CostUSD * 0.5,  // Estimate 50% for output cost
		TotalCost:        metric.CostUSD,
		EstimatedCost:    metric.CostUSD,
		CostDifference:   0.0, // No difference since we only have one cost value
		Currency:         "USD",
		BillingTier:      "standard", // Default tier
		DiscountApplied:  0.0,
		Timestamp:        metric.Timestamp,
	}

	w.mutex.Lock()
	w.costBuffer = append(w.costBuffer, costMetric)
	w.mutex.Unlock()

	// Immediate flush if buffer is full
	if shouldFlush {
		go w.flushRequestMetrics(ctx)
	}

	return nil
}

// GetUsageStats retrieves usage statistics
func (w *GatewayAnalyticsWorker) GetUsageStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*UsageMetric, error) {
	return w.analyticsRepository.GetUsageStats(ctx, orgID, period, start, end)
}

// GetCostStats retrieves cost statistics
func (w *GatewayAnalyticsWorker) GetCostStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*CostMetric, error) {
	return w.analyticsRepository.GetCostStats(ctx, orgID, period, start, end)
}

// GenerateBill generates a billing summary for an organization
func (w *GatewayAnalyticsWorker) GenerateBill(ctx context.Context, orgID ulid.ULID, period string) (*BillingSummary, error) {
	return w.billingService.CalculateBill(ctx, orgID, period)
}

// Internal worker methods

func (w *GatewayAnalyticsWorker) flushWorker(ctx context.Context) {
	defer w.wg.Done()
	
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.flushAllBuffers(ctx)
		}
	}
}

func (w *GatewayAnalyticsWorker) processWorker(ctx context.Context) {
	defer w.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute) // Process aggregations every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processAggregations(ctx)
		}
	}
}

func (w *GatewayAnalyticsWorker) flushAllBuffers(ctx context.Context) {
	w.flushRequestMetrics(ctx)
	w.flushUsageMetrics(ctx)
	w.flushCostMetrics(ctx)
}

func (w *GatewayAnalyticsWorker) flushRequestMetrics(ctx context.Context) {
	w.mutex.Lock()
	if len(w.requestBuffer) == 0 {
		w.mutex.Unlock()
		return
	}
	
	batch := make([]*RequestMetric, len(w.requestBuffer))
	copy(batch, w.requestBuffer)
	w.requestBuffer = w.requestBuffer[:0]
	w.mutex.Unlock()

	if err := w.analyticsRepository.BatchInsertRequestMetrics(ctx, batch); err != nil {
		w.logger.WithError(err).WithField("batch_size", len(batch)).Error("Failed to flush request metrics")
		// TODO: Implement retry logic or dead letter queue
	} else {
		w.logger.WithField("batch_size", len(batch)).Debug("Flushed request metrics")
	}
}

func (w *GatewayAnalyticsWorker) flushUsageMetrics(ctx context.Context) {
	w.mutex.Lock()
	if len(w.usageBuffer) == 0 {
		w.mutex.Unlock()
		return
	}
	
	batch := make([]*UsageMetric, len(w.usageBuffer))
	copy(batch, w.usageBuffer)
	w.usageBuffer = w.usageBuffer[:0]
	w.mutex.Unlock()

	if err := w.analyticsRepository.BatchInsertUsageMetrics(ctx, batch); err != nil {
		w.logger.WithError(err).WithField("batch_size", len(batch)).Error("Failed to flush usage metrics")
	} else {
		w.logger.WithField("batch_size", len(batch)).Debug("Flushed usage metrics")
	}
}

func (w *GatewayAnalyticsWorker) flushCostMetrics(ctx context.Context) {
	w.mutex.Lock()
	if len(w.costBuffer) == 0 {
		w.mutex.Unlock()
		return
	}
	
	batch := make([]*CostMetric, len(w.costBuffer))
	copy(batch, w.costBuffer)
	w.costBuffer = w.costBuffer[:0]
	w.mutex.Unlock()

	// Send to billing service first
	for _, metric := range batch {
		if err := w.billingService.RecordUsage(ctx, metric); err != nil {
			w.logger.WithError(err).WithField("request_id", metric.RequestID).Error("Failed to record usage for billing")
		}
	}

	// Then store in analytics
	if err := w.analyticsRepository.BatchInsertCostMetrics(ctx, batch); err != nil {
		w.logger.WithError(err).WithField("batch_size", len(batch)).Error("Failed to flush cost metrics")
	} else {
		w.logger.WithField("batch_size", len(batch)).Debug("Flushed cost metrics")
	}
}

func (w *GatewayAnalyticsWorker) processAggregations(ctx context.Context) {
	// TODO: Implement usage aggregation logic
	// This would create hourly/daily/monthly usage summaries
	// from the raw request metrics for faster reporting
	
	w.logger.Debug("Processing usage aggregations")
	
	// Example aggregation logic:
	// 1. Group request metrics by organization, provider, model, time period
	// 2. Calculate totals, averages, percentiles
	// 3. Store aggregated usage metrics
	// 4. Clean up old raw metrics if needed
}

// IsHealthy returns true if the worker is healthy
func (w *GatewayAnalyticsWorker) IsHealthy() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.running {
		return false
	}

	// Check buffer sizes - consider unhealthy if buffers are too full
	maxBufferSize := w.batchSize * 5 // Allow 5 batches worth of buffering
	if len(w.requestBuffer) > maxBufferSize || 
	   len(w.usageBuffer) > maxBufferSize || 
	   len(w.costBuffer) > maxBufferSize {
		return false
	}

	return true
}

// Health check method
func (w *GatewayAnalyticsWorker) GetHealth() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return map[string]interface{}{
		"running":                w.running,
		"request_buffer_size":    len(w.requestBuffer),
		"usage_buffer_size":      len(w.usageBuffer),
		"cost_buffer_size":       len(w.costBuffer),
		"batch_size":             w.batchSize,
		"flush_interval_seconds": w.flushInterval.Seconds(),
		"healthy":                w.IsHealthy(),
	}
}

// Metrics for monitoring
func (w *GatewayAnalyticsWorker) GetMetrics() map[string]interface{} {
	health := w.GetHealth()
	
	// Additional metrics would be added here
	// e.g., processed message counts, error rates, processing latency
	
	return health
}