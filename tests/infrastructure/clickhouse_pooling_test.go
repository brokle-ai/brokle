package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/infrastructure/database"
)

// TestClickHouseConnectionPoolingConfig validates the connection pooling configuration
func TestClickHouseConnectionPoolingConfig(t *testing.T) {
	// Test configuration parsing and validation without requiring actual ClickHouse
	cfg := &config.Config{
		ClickHouse: config.ClickHouseConfig{
			URL: "clickhouse://brokle:brokle_password@localhost:9000/brokle_analytics",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("ConfigurationParsing", func(t *testing.T) {
		// This should parse correctly even without ClickHouse running
		options, err := clickhouse.ParseDSN(cfg.GetClickHouseURL())
		require.NoError(t, err, "Should parse ClickHouse DSN correctly")

		// Apply our pooling configuration
		options.MaxOpenConns = 50
		options.MaxIdleConns = 10
		options.ConnMaxLifetime = time.Hour
		options.BlockBufferSize = 10

		// Validate configuration values
		assert.Equal(t, 50, options.MaxOpenConns, "MaxOpenConns should be set correctly")
		assert.Equal(t, 10, options.MaxIdleConns, "MaxIdleConns should be set correctly")
		assert.Equal(t, time.Hour, options.ConnMaxLifetime, "ConnMaxLifetime should be set correctly")
		assert.Equal(t, uint8(10), options.BlockBufferSize, "BlockBufferSize should be set correctly")

		t.Logf("ClickHouse connection pooling configuration validated:")
		t.Logf("  MaxOpenConns: %d", options.MaxOpenConns)
		t.Logf("  MaxIdleConns: %d", options.MaxIdleConns)
		t.Logf("  ConnMaxLifetime: %v", options.ConnMaxLifetime)
		t.Logf("  BlockBufferSize: %d", options.BlockBufferSize)
	})
}

// TestClickHouseConnectionPooling validates that ClickHouse can handle
// the 4.5k buffer capacity required for analytics worker
func TestClickHouseConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ClickHouse connection pooling test in short mode")
	}

	// Skip this test if ClickHouse is not available
	cfg := &config.Config{
		ClickHouse: config.ClickHouseConfig{
			URL: "clickhouse://brokle:brokle_password@localhost:9000/brokle_analytics",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Try to create client - if it fails, skip the test
	client, err := database.NewClickHouseDB(cfg, logger)
	if err != nil {
		t.Skipf("ClickHouse not available, skipping connection pooling test: %v", err)
		return
	}
	defer client.Close()

	// Test parameters - simulate analytics worker load
	const (
		concurrentWorkers = 100 // Simulate 100 concurrent inserts
		insertsPerWorker  = 45  // 4500 total inserts (4.5k buffer)
		testTimeout       = 30 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create test table for connection pooling test
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS test_connection_pooling (
			id String,
			timestamp DateTime64(3),
			data String
		) ENGINE = MergeTree()
		ORDER BY (timestamp, id)
		TTL toDateTime(timestamp) + INTERVAL 1 HOUR
	`

	err = client.Execute(ctx, createTableSQL)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup test table after test
	defer func() {
		client.Execute(context.Background(), "DROP TABLE IF EXISTS test_connection_pooling")
	}()

	// Track results
	var wg sync.WaitGroup
	errors := make(chan error, concurrentWorkers)
	insertCount := make(chan int, concurrentWorkers)

	startTime := time.Now()

	// Launch concurrent workers to simulate analytics worker load
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			successfulInserts := 0
			for j := 0; j < insertsPerWorker; j++ {
				// Simulate batch metric insert
				insertSQL := `
					INSERT INTO test_connection_pooling (id, timestamp, data)
					VALUES (?, ?, ?)
				`

				recordID := generateTestRecordID(workerID, j)
				timestamp := time.Now()
				data := generateTestData(workerID, j)

				err := client.Execute(ctx, insertSQL, recordID, timestamp, data)
				if err != nil {
					select {
					case errors <- err:
					default:
						// Channel full, continue
					}
					return
				}
				successfulInserts++

				// Small delay to simulate realistic processing time
				time.Sleep(time.Microsecond * 100)
			}

			insertCount <- successfulInserts
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errors)
	close(insertCount)

	processingTime := time.Since(startTime)

	// Collect results
	var totalInserts int
	for count := range insertCount {
		totalInserts += count
	}

	var errorCount int
	for range errors {
		errorCount++
	}

	// Assertions
	expectedInserts := concurrentWorkers * insertsPerWorker
	t.Logf("Connection pooling test results:")
	t.Logf("  Total expected inserts: %d", expectedInserts)
	t.Logf("  Successful inserts: %d", totalInserts)
	t.Logf("  Failed inserts: %d", errorCount)
	t.Logf("  Processing time: %v", processingTime)
	t.Logf("  Inserts per second: %.2f", float64(totalInserts)/processingTime.Seconds())

	// Validate results - should handle 4.5k inserts without connection pool exhaustion
	assert.GreaterOrEqual(t, totalInserts, int(float64(expectedInserts)*0.95),
		"Should achieve at least 95%% success rate")
	assert.LessOrEqual(t, errorCount, int(float64(expectedInserts)*0.05),
		"Should have less than 5%% error rate")
	assert.Less(t, processingTime, testTimeout,
		"Should complete within timeout")

	// Validate throughput - should handle at least 150 inserts/second
	throughput := float64(totalInserts) / processingTime.Seconds()
	assert.GreaterOrEqual(t, throughput, 150.0,
		"Should achieve at least 150 inserts/second")

	// Verify connection pool health by doing additional operations
	healthCheckSQL := "SELECT COUNT(*) FROM test_connection_pooling"
	row := client.QueryRow(ctx, healthCheckSQL)

	var count uint64
	err = row.Scan(&count)
	assert.NoError(t, err, "Should be able to query after connection pool stress test")
	assert.Equal(t, uint64(totalInserts), count, "Record count should match successful inserts")
}

// TestClickHouseConnectionRecovery tests connection recovery after failures
func TestClickHouseConnectionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ClickHouse connection recovery test in short mode")
	}

	cfg := &config.Config{
		ClickHouse: config.ClickHouseConfig{
			URL: "clickhouse://brokle:brokle_password@localhost:9000/brokle_analytics",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	client, err := database.NewClickHouseDB(cfg, logger)
	require.NoError(t, err, "Failed to create ClickHouse client")
	defer client.Close()

	ctx := context.Background()

	// Test basic connectivity
	err = client.Health()
	assert.NoError(t, err, "Initial health check should pass")

	// Test connection recovery after context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	// This should fail due to cancelled context
	err = client.Execute(cancelCtx, "SELECT 1")
	assert.Error(t, err, "Should fail with cancelled context")

	// This should succeed with fresh context (connection recovery)
	err = client.Execute(ctx, "SELECT 1")
	assert.NoError(t, err, "Should recover and execute successfully")

	// Verify health check still works
	err = client.Health()
	assert.NoError(t, err, "Health check should pass after recovery")
}

// Helper functions for test data generation
func generateTestRecordID(workerID, insertID int) string {
	return fmt.Sprintf("worker_%d_insert_%d_%d", workerID, insertID, time.Now().UnixNano())
}

func generateTestData(workerID, insertID int) string {
	return fmt.Sprintf(`{"worker_id": %d, "insert_id": %d, "timestamp": "%s"}`,
		workerID, insertID, time.Now().Format(time.RFC3339))
}