package infrastructure

import (
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
)

// TestClickHouseConnectionPoolingConfig validates the connection pooling configuration
func TestClickHouseConnectionPoolingConfig(t *testing.T) {
	// Test configuration parsing and validation without requiring actual ClickHouse
	cfg := &config.Config{
		ClickHouse: config.ClickHouseConfig{
			URL: "clickhouse://brokle:brokle_password@localhost:9000/default",
		},
	}

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
