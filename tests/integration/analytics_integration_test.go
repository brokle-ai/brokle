//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"brokle/internal/config"
	"brokle/internal/core/domain/gateway"
	"brokle/internal/workers/analytics"
	"brokle/pkg/database"
	"brokle/pkg/ulid"
)

// AnalyticsIntegrationTestSuite provides tests for analytics operations
type AnalyticsIntegrationTestSuite struct {
	suite.Suite
	cfg                 *config.Config
	chDB                *sql.DB
	analyticsRepo       analytics.AnalyticsRepository
	billingService      analytics.BillingService
	gatewayWorker       *analytics.GatewayAnalyticsWorker
	logger              *logrus.Logger
	ctx                 context.Context
	testOrgID          ulid.ULID
	testProviderID     ulid.ULID
	testModelID        ulid.ULID
}

// MockAnalyticsRepository implements a mock analytics repository for testing
type MockAnalyticsRepository struct {
	requestMetrics []*analytics.RequestMetric
	usageMetrics   []*analytics.UsageMetric
	costMetrics    []*analytics.CostMetric
}

func (m *MockAnalyticsRepository) BatchInsertRequestMetrics(ctx context.Context, metrics []*analytics.RequestMetric) error {
	m.requestMetrics = append(m.requestMetrics, metrics...)
	return nil
}

func (m *MockAnalyticsRepository) BatchInsertUsageMetrics(ctx context.Context, metrics []*analytics.UsageMetric) error {
	m.usageMetrics = append(m.usageMetrics, metrics...)
	return nil
}

func (m *MockAnalyticsRepository) BatchInsertCostMetrics(ctx context.Context, metrics []*analytics.CostMetric) error {
	m.costMetrics = append(m.costMetrics, metrics...)
	return nil
}

func (m *MockAnalyticsRepository) GetUsageStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*analytics.UsageMetric, error) {
	var filtered []*analytics.UsageMetric
	for _, metric := range m.usageMetrics {
		if metric.OrganizationID == orgID &&
			metric.PeriodStart.After(start.Add(-time.Second)) &&
			metric.PeriodEnd.Before(end.Add(time.Second)) {
			filtered = append(filtered, metric)
		}
	}
	return filtered, nil
}

func (m *MockAnalyticsRepository) GetCostStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*analytics.CostMetric, error) {
	var filtered []*analytics.CostMetric
	for _, metric := range m.costMetrics {
		if metric.OrganizationID == orgID &&
			metric.Timestamp.After(start.Add(-time.Second)) &&
			metric.Timestamp.Before(end.Add(time.Second)) {
			filtered = append(filtered, metric)
		}
	}
	return filtered, nil
}

// MockBillingService implements a mock billing service for testing
type MockBillingService struct {
	usageRecords    []*analytics.CostMetric
	billingSummary  *analytics.BillingSummary
	billingRecords  []*analytics.BillingRecord
}

func (m *MockBillingService) RecordUsage(ctx context.Context, usage *analytics.CostMetric) error {
	m.usageRecords = append(m.usageRecords, usage)
	return nil
}

func (m *MockBillingService) CalculateBill(ctx context.Context, orgID ulid.ULID, period string) (*analytics.BillingSummary, error) {
	if m.billingSummary != nil {
		return m.billingSummary, nil
	}

	// Calculate from recorded usage
	var totalCost float64
	var totalRequests int64
	var totalTokens int64
	
	for _, usage := range m.usageRecords {
		if usage.OrganizationID == orgID {
			totalCost += usage.TotalCost
			totalRequests++
			totalTokens += int64(usage.TotalTokens)
		}
	}

	return &analytics.BillingSummary{
		OrganizationID: orgID,
		Period:         period,
		PeriodStart:    time.Now().AddDate(0, 0, -30),
		PeriodEnd:      time.Now(),
		TotalRequests:  totalRequests,
		TotalTokens:    totalTokens,
		TotalCost:      totalCost,
		Currency:       "USD",
		NetCost:        totalCost,
		Status:         "calculated",
		GeneratedAt:    time.Now(),
	}, nil
}

func (m *MockBillingService) GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*analytics.BillingRecord, error) {
	var filtered []*analytics.BillingRecord
	for _, record := range m.billingRecords {
		if record.OrganizationID == orgID &&
			record.CreatedAt.After(start.Add(-time.Second)) &&
			record.CreatedAt.Before(end.Add(time.Second)) {
			filtered = append(filtered, record)
		}
	}
	return filtered, nil
}

// SetupSuite sets up the test suite with analytics infrastructure
func (suite *AnalyticsIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Load test configuration
	cfg, err := config.Load()
	require.NoError(suite.T(), err)

	// Override configuration for testing
	cfg.Database.ClickHouse.Database = cfg.Database.ClickHouse.Database + "_test"
	suite.cfg = cfg

	// Initialize ClickHouse connection
	suite.chDB, err = database.NewClickHouse(cfg.Database.ClickHouse)
	require.NoError(suite.T(), err)

	// Initialize logger
	suite.logger = logrus.New()
	suite.logger.SetLevel(logrus.DebugLevel)

	// Initialize mock repositories and services
	suite.analyticsRepo = &MockAnalyticsRepository{}
	suite.billingService = &MockBillingService{}

	// Create analytics worker
	workerConfig := analytics.DefaultWorkerConfig()
	workerConfig.BatchSize = 10 // Smaller batch size for testing
	workerConfig.FlushInterval = 100 * time.Millisecond // Faster flush for testing

	suite.gatewayWorker = analytics.NewGatewayAnalyticsWorker(
		suite.logger,
		workerConfig,
		suite.analyticsRepo,
		suite.billingService,
	)

	// Setup test data
	suite.setupTestData()

	// Verify ClickHouse connectivity
	err = suite.chDB.Ping()
	require.NoError(suite.T(), err, "Failed to connect to ClickHouse")
}

// TearDownSuite cleans up after the test suite
func (suite *AnalyticsIntegrationTestSuite) TearDownSuite() {
	// Stop the analytics worker if running
	if suite.gatewayWorker != nil {
		_ = suite.gatewayWorker.Stop()
	}

	// Close database connection
	if suite.chDB != nil {
		_ = suite.chDB.Close()
	}
}

// setupTestData creates test IDs for integration tests
func (suite *AnalyticsIntegrationTestSuite) setupTestData() {
	suite.testOrgID = ulid.New()
	suite.testProviderID = ulid.New()
	suite.testModelID = ulid.New()
}

// TestAnalyticsWorkerLifecycle tests worker start/stop lifecycle
func (suite *AnalyticsIntegrationTestSuite) TestAnalyticsWorkerLifecycle() {
	// Test initial state
	assert.False(suite.T(), suite.gatewayWorker.IsHealthy(), "Worker should not be healthy initially")

	// Start worker
	suite.gatewayWorker.Start()
	time.Sleep(50 * time.Millisecond) // Give worker time to start

	assert.True(suite.T(), suite.gatewayWorker.IsHealthy(), "Worker should be healthy after start")

	// Check health status
	health := suite.gatewayWorker.GetHealth()
	assert.True(suite.T(), health["running"].(bool))
	assert.Equal(suite.T(), 0, health["request_buffer_size"])
	assert.Equal(suite.T(), 0, health["usage_buffer_size"])
	assert.Equal(suite.T(), 0, health["cost_buffer_size"])

	// Stop worker
	err := suite.gatewayWorker.Stop()
	require.NoError(suite.T(), err)

	assert.False(suite.T(), suite.gatewayWorker.IsHealthy(), "Worker should not be healthy after stop")
}

// TestRequestMetricsProcessing tests processing of gateway request metrics
func (suite *AnalyticsIntegrationTestSuite) TestRequestMetricsProcessing() {
	// Start worker
	suite.gatewayWorker.Start()
	defer suite.gatewayWorker.Stop()

	// Create test request metrics
	requestMetrics := []*gateway.RequestMetrics{
		{
			RequestID:       ulid.New().String(),
			ProjectID:       suite.testOrgID,
			Environment:     "production",
			Provider:        "openai",
			Model:           "gpt-3.5-turbo",
			InputTokens:     100,
			OutputTokens:    50,
			TotalTokens:     150,
			CostUSD:         0.003,
			LatencyMs:       250,
			Success:         true,
			CacheHit:        false,
			RoutingStrategy: "cost_optimization",
			Timestamp:       time.Now(),
		},
		{
			RequestID:       ulid.New().String(),
			ProjectID:       suite.testOrgID,
			Environment:     "staging",
			Provider:        "anthropic",
			Model:           "claude-3-haiku",
			InputTokens:     80,
			OutputTokens:    40,
			TotalTokens:     120,
			CostUSD:         0.002,
			LatencyMs:       180,
			Success:         true,
			CacheHit:        true,
			RoutingStrategy: "latency_optimization",
			Timestamp:       time.Now(),
		},
		{
			RequestID:    ulid.New().String(),
			ProjectID:    suite.testOrgID,
			Environment:  "production",
			Provider:     "openai",
			Model:        "gpt-4",
			InputTokens:  200,
			OutputTokens: 100,
			TotalTokens:  300,
			CostUSD:      0.012,
			LatencyMs:    500,
			Success:      false,
			CacheHit:     false,
			ErrorMessage: func() *string { s := "Rate limit exceeded"; return &s }(),
			Timestamp:    time.Now(),
		},
	}

	// Record metrics
	for _, metric := range requestMetrics {
		err := suite.gatewayWorker.RecordRequest(suite.ctx, metric)
		require.NoError(suite.T(), err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify metrics were processed
	mockRepo := suite.analyticsRepo.(*MockAnalyticsRepository)
	assert.Len(suite.T(), mockRepo.requestMetrics, 3, "Should have 3 request metrics")
	assert.Len(suite.T(), mockRepo.costMetrics, 3, "Should have 3 cost metrics")

	// Verify billing service received usage records
	mockBilling := suite.billingService.(*MockBillingService)
	assert.Len(suite.T(), mockBilling.usageRecords, 3, "Should have 3 usage records")

	// Verify request metric data
	firstRequest := mockRepo.requestMetrics[0]
	assert.Equal(suite.T(), suite.testOrgID, firstRequest.OrganizationID)
	assert.Equal(suite.T(), gateway.Environment("production"), firstRequest.Environment)
	assert.Equal(suite.T(), "openai", firstRequest.ProviderName)
	assert.Equal(suite.T(), "gpt-3.5-turbo", firstRequest.ModelName)
	assert.Equal(suite.T(), int32(100), firstRequest.InputTokens)
	assert.Equal(suite.T(), int32(50), firstRequest.OutputTokens)
	assert.Equal(suite.T(), "success", firstRequest.Status)
	assert.False(suite.T(), firstRequest.CacheHit)

	// Verify error handling
	thirdRequest := mockRepo.requestMetrics[2]
	assert.Equal(suite.T(), "error", thirdRequest.Status)
	assert.Equal(suite.T(), "Rate limit exceeded", thirdRequest.Error)

	// Verify cost metric data
	firstCost := mockRepo.costMetrics[0]
	assert.Equal(suite.T(), suite.testOrgID, firstCost.OrganizationID)
	assert.Equal(suite.T(), 0.003, firstCost.TotalCost)
	assert.Equal(suite.T(), "USD", firstCost.Currency)
	assert.Equal(suite.T(), int32(150), firstCost.TotalTokens)
}

// TestBatchProcessing tests batch processing and buffer management
func (suite *AnalyticsIntegrationTestSuite) TestBatchProcessing() {
	// Start worker with small batch size
	suite.gatewayWorker.Start()
	defer suite.gatewayWorker.Stop()

	const numRequests = 25
	batchSize := 10

	// Create multiple requests to trigger batching
	for i := 0; i < numRequests; i++ {
		metric := &gateway.RequestMetrics{
			RequestID:       ulid.New().String(),
			ProjectID:       suite.testOrgID,
			Environment:     "production",
			Provider:        "openai",
			Model:           "gpt-3.5-turbo",
			InputTokens:     50,
			OutputTokens:    25,
			TotalTokens:     75,
			CostUSD:         0.0015,
			LatencyMs:       100,
			Success:         true,
			CacheHit:        i%3 == 0, // Every third request is a cache hit
			RoutingStrategy: "cost_optimization",
			Timestamp:       time.Now(),
		}

		err := suite.gatewayWorker.RecordRequest(suite.ctx, metric)
		require.NoError(suite.T(), err)

		// Check buffer size occasionally
		if i == batchSize-1 {
			health := suite.gatewayWorker.GetHealth()
			bufferSize := health["request_buffer_size"].(int)
			assert.LessOrEqual(suite.T(), bufferSize, batchSize, "Buffer should not exceed batch size")
		}
	}

	// Wait for all processing to complete
	time.Sleep(300 * time.Millisecond)

	// Verify all requests were processed
	mockRepo := suite.analyticsRepo.(*MockAnalyticsRepository)
	assert.Len(suite.T(), mockRepo.requestMetrics, numRequests, "All requests should be processed")
	assert.Len(suite.T(), mockRepo.costMetrics, numRequests, "All cost metrics should be processed")

	// Verify buffer is empty after processing
	health := suite.gatewayWorker.GetHealth()
	assert.Equal(suite.T(), 0, health["request_buffer_size"], "Request buffer should be empty")
	assert.Equal(suite.T(), 0, health["cost_buffer_size"], "Cost buffer should be empty")

	// Verify cache hit statistics
	cacheHits := 0
	for _, metric := range mockRepo.requestMetrics {
		if metric.CacheHit {
			cacheHits++
		}
	}
	expectedCacheHits := (numRequests + 2) / 3 // Every third request
	assert.Equal(suite.T(), expectedCacheHits, cacheHits, "Cache hit count should match expected")
}

// TestUsageStatsRetrieval tests retrieval of usage statistics
func (suite *AnalyticsIntegrationTestSuite) TestUsageStatsRetrieval() {
	// Add some mock usage metrics to the repository
	mockRepo := suite.analyticsRepo.(*MockAnalyticsRepository)
	now := time.Now()
	
	usageMetrics := []*analytics.UsageMetric{
		{
			ID:                ulid.New(),
			OrganizationID:    suite.testOrgID,
			Environment:       gateway.EnvironmentProduction,
			ProviderID:        suite.testProviderID,
			ModelID:           suite.testModelID,
			RequestType:       gateway.RequestTypeChatCompletion,
			Period:            "daily",
			PeriodStart:       now.AddDate(0, 0, -1),
			PeriodEnd:         now,
			RequestCount:      100,
			SuccessCount:      95,
			ErrorCount:        5,
			TotalInputTokens:  5000,
			TotalOutputTokens: 2500,
			TotalTokens:       7500,
			TotalCost:         0.15,
			Currency:          "USD",
			AvgDuration:       250.5,
			CacheHitRate:      0.15,
			Timestamp:         now,
		},
		{
			ID:                ulid.New(),
			OrganizationID:    ulid.New(), // Different org
			Environment:       gateway.EnvironmentProduction,
			ProviderID:        suite.testProviderID,
			ModelID:           suite.testModelID,
			RequestType:       gateway.RequestTypeChatCompletion,
			Period:            "daily",
			PeriodStart:       now.AddDate(0, 0, -1),
			PeriodEnd:         now,
			RequestCount:      50,
			SuccessCount:      48,
			ErrorCount:        2,
			TotalInputTokens:  2000,
			TotalOutputTokens: 1000,
			TotalTokens:       3000,
			TotalCost:         0.06,
			Currency:          "USD",
			AvgDuration:       200.0,
			CacheHitRate:      0.10,
			Timestamp:         now,
		},
	}

	mockRepo.usageMetrics = usageMetrics

	// Test retrieval for our test organization
	stats, err := suite.gatewayWorker.GetUsageStats(
		suite.ctx,
		suite.testOrgID,
		"daily",
		now.AddDate(0, 0, -2),
		now.Add(time.Hour),
	)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), stats, 1, "Should return 1 usage metric for test org")

	stat := stats[0]
	assert.Equal(suite.T(), suite.testOrgID, stat.OrganizationID)
	assert.Equal(suite.T(), int64(100), stat.RequestCount)
	assert.Equal(suite.T(), int64(95), stat.SuccessCount)
	assert.Equal(suite.T(), int64(5), stat.ErrorCount)
	assert.Equal(suite.T(), 0.15, stat.TotalCost)
	assert.Equal(suite.T(), 0.15, stat.CacheHitRate)
}

// TestCostStatsRetrieval tests retrieval of cost statistics
func (suite *AnalyticsIntegrationTestSuite) TestCostStatsRetrieval() {
	// Add some mock cost metrics to the repository
	mockRepo := suite.analyticsRepo.(*MockAnalyticsRepository)
	now := time.Now()

	costMetrics := []*analytics.CostMetric{
		{
			ID:             ulid.New(),
			RequestID:      ulid.New(),
			OrganizationID: suite.testOrgID,
			Environment:    gateway.EnvironmentProduction,
			ProviderID:     suite.testProviderID,
			ModelID:        suite.testModelID,
			RequestType:    gateway.RequestTypeChatCompletion,
			InputTokens:    100,
			OutputTokens:   50,
			TotalTokens:    150,
			InputCost:      0.001,
			OutputCost:     0.002,
			TotalCost:      0.003,
			EstimatedCost:  0.0025,
			CostDifference: 0.0005,
			Currency:       "USD",
			BillingTier:    "standard",
			DiscountApplied: 0.0,
			Timestamp:      now,
		},
		{
			ID:             ulid.New(),
			RequestID:      ulid.New(),
			OrganizationID: suite.testOrgID,
			Environment:    gateway.EnvironmentStaging,
			ProviderID:     suite.testProviderID,
			ModelID:        suite.testModelID,
			RequestType:    gateway.RequestTypeCompletion,
			InputTokens:    80,
			OutputTokens:   40,
			TotalTokens:    120,
			InputCost:      0.0008,
			OutputCost:     0.0016,
			TotalCost:      0.0024,
			EstimatedCost:  0.002,
			CostDifference: 0.0004,
			Currency:       "USD",
			BillingTier:    "premium",
			DiscountApplied: 0.1,
			Timestamp:      now.Add(-time.Hour),
		},
	}

	mockRepo.costMetrics = costMetrics

	// Test retrieval for our test organization
	stats, err := suite.gatewayWorker.GetCostStats(
		suite.ctx,
		suite.testOrgID,
		"daily",
		now.AddDate(0, 0, -1),
		now.Add(time.Hour),
	)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), stats, 2, "Should return 2 cost metrics for test org")

	// Verify first cost metric
	assert.Equal(suite.T(), suite.testOrgID, stats[0].OrganizationID)
	assert.Equal(suite.T(), 0.003, stats[0].TotalCost)
	assert.Equal(suite.T(), int32(150), stats[0].TotalTokens)
	assert.Equal(suite.T(), gateway.EnvironmentProduction, stats[0].Environment)

	// Verify second cost metric
	assert.Equal(suite.T(), suite.testOrgID, stats[1].OrganizationID)
	assert.Equal(suite.T(), 0.0024, stats[1].TotalCost)
	assert.Equal(suite.T(), int32(120), stats[1].TotalTokens)
	assert.Equal(suite.T(), gateway.EnvironmentStaging, stats[1].Environment)
}

// TestBillingGeneration tests billing summary generation
func (suite *AnalyticsIntegrationTestSuite) TestBillingGeneration() {
	// Setup mock billing service with some usage records
	mockBilling := suite.billingService.(*MockBillingService)
	
	// Add usage records to billing service
	costMetrics := []*analytics.CostMetric{
		{
			OrganizationID: suite.testOrgID,
			TotalCost:      0.005,
			TotalTokens:    200,
			Currency:       "USD",
			Timestamp:      time.Now(),
		},
		{
			OrganizationID: suite.testOrgID,
			TotalCost:      0.003,
			TotalTokens:    150,
			Currency:       "USD",
			Timestamp:      time.Now(),
		},
		{
			OrganizationID: ulid.New(), // Different org
			TotalCost:      0.010,
			TotalTokens:    500,
			Currency:       "USD",
			Timestamp:      time.Now(),
		},
	}

	mockBilling.usageRecords = costMetrics

	// Generate billing summary
	bill, err := suite.gatewayWorker.GenerateBill(suite.ctx, suite.testOrgID, "monthly")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), bill)

	// Verify billing summary
	assert.Equal(suite.T(), suite.testOrgID, bill.OrganizationID)
	assert.Equal(suite.T(), "monthly", bill.Period)
	assert.Equal(suite.T(), int64(2), bill.TotalRequests) // Only 2 for our org
	assert.Equal(suite.T(), int64(350), bill.TotalTokens) // 200 + 150
	assert.Equal(suite.T(), 0.008, bill.TotalCost) // 0.005 + 0.003
	assert.Equal(suite.T(), "USD", bill.Currency)
	assert.Equal(suite.T(), 0.008, bill.NetCost)
	assert.Equal(suite.T(), "calculated", bill.Status)
	assert.NotZero(suite.T(), bill.GeneratedAt)
}

// TestWorkerHealthMonitoring tests worker health monitoring
func (suite *AnalyticsIntegrationTestSuite) TestWorkerHealthMonitoring() {
	// Worker should be unhealthy when stopped
	assert.False(suite.T(), suite.gatewayWorker.IsHealthy())

	health := suite.gatewayWorker.GetHealth()
	assert.False(suite.T(), health["running"].(bool))
	assert.False(suite.T(), health["healthy"].(bool))

	// Start worker
	suite.gatewayWorker.Start()
	defer suite.gatewayWorker.Stop()

	// Worker should be healthy when running
	assert.True(suite.T(), suite.gatewayWorker.IsHealthy())

	health = suite.gatewayWorker.GetHealth()
	assert.True(suite.T(), health["running"].(bool))
	assert.True(suite.T(), health["healthy"].(bool))
	assert.Equal(suite.T(), 10, health["batch_size"]) // Our test config
	assert.Equal(suite.T(), 0.1, health["flush_interval_seconds"]) // 100ms

	// Get detailed metrics
	metrics := suite.gatewayWorker.GetMetrics()
	assert.NotNil(suite.T(), metrics)
	assert.True(suite.T(), metrics["healthy"].(bool))
}

// TestClickHouseConnectivity tests direct ClickHouse operations
func (suite *AnalyticsIntegrationTestSuite) TestClickHouseConnectivity() {
	// Test basic connectivity
	err := suite.chDB.Ping()
	require.NoError(suite.T(), err, "ClickHouse should be accessible")

	// Test creating a temporary table
	tableName := "test_analytics_" + ulid.New().String()
	createSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id String,
			organization_id String,
			timestamp DateTime64(3),
			request_count UInt64,
			total_cost Float64
		) ENGINE = MergeTree()
		ORDER BY (organization_id, timestamp)
		TTL timestamp + INTERVAL 30 DAY DELETE
	`, tableName)

	_, err = suite.chDB.Exec(createSQL)
	require.NoError(suite.T(), err, "Should be able to create test table")

	// Test inserting data
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (id, organization_id, timestamp, request_count, total_cost) VALUES
		(?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?)
	`, tableName)

	now := time.Now()
	_, err = suite.chDB.Exec(insertSQL,
		"test-1", suite.testOrgID.String(), now, uint64(10), 0.05,
		"test-2", suite.testOrgID.String(), now.Add(-time.Hour), uint64(5), 0.02,
	)
	require.NoError(suite.T(), err, "Should be able to insert data")

	// Test querying data
	querySQL := fmt.Sprintf(`
		SELECT COUNT(*), SUM(request_count), SUM(total_cost)
		FROM %s
		WHERE organization_id = ?
	`, tableName)

	var count uint64
	var totalRequests uint64
	var totalCost float64

	err = suite.chDB.QueryRow(querySQL, suite.testOrgID.String()).Scan(&count, &totalRequests, &totalCost)
	require.NoError(suite.T(), err, "Should be able to query data")

	assert.Equal(suite.T(), uint64(2), count)
	assert.Equal(suite.T(), uint64(15), totalRequests)
	assert.Equal(suite.T(), 0.07, totalCost)

	// Clean up test table
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err = suite.chDB.Exec(dropSQL)
	require.NoError(suite.T(), err, "Should be able to drop test table")
}

// TestAnalyticsIntegrationSuite runs the complete analytics integration test suite
func TestAnalyticsIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AnalyticsIntegrationTestSuite))
}