//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"brokle/internal/config"
	"brokle/internal/core/domain/gateway"
	gatewayRepo "brokle/internal/infrastructure/repository/gateway"
	"brokle/pkg/database"
	"brokle/pkg/ulid"
)

// DatabaseIntegrationTestSuite provides tests for database operations
type DatabaseIntegrationTestSuite struct {
	suite.Suite
	cfg                      *config.Config
	pgDB                     *sql.DB
	chDB                     *sql.DB
	providerRepo             gateway.ProviderRepository
	modelRepo                gateway.ModelRepository
	providerConfigRepo       gateway.ProviderConfigRepository
	ctx                      context.Context
	testProviderID           ulid.ULID
	testModelID              ulid.ULID
	testProviderConfigID     ulid.ULID
}

// SetupSuite sets up the test suite with database connections
func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set server mode and JWT secret for config validation
	os.Setenv("APP_MODE", "server")
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-integration-tests-32-characters-long")

	// Load test configuration
	cfg, err := config.Load()
	require.NoError(suite.T(), err)

	// Override configuration for testing
	cfg.Database.PostgreSQL.Database = cfg.Database.PostgreSQL.Database + "_test"
	cfg.Database.ClickHouse.Database = cfg.Database.ClickHouse.Database + "_test"
	suite.cfg = cfg

	// Initialize PostgreSQL connection
	suite.pgDB, err = database.NewPostgreSQL(cfg.Database.PostgreSQL)
	require.NoError(suite.T(), err)

	// Initialize ClickHouse connection
	suite.chDB, err = database.NewClickHouse(cfg.Database.ClickHouse)
	require.NoError(suite.T(), err)

	// Initialize repositories
	suite.providerRepo = gatewayRepo.NewProviderRepository(suite.pgDB)
	suite.modelRepo = gatewayRepo.NewModelRepository(suite.pgDB)
	suite.providerConfigRepo = gatewayRepo.NewProviderConfigRepository(suite.pgDB)

	// Verify database connectivity
	suite.verifyDatabaseConnectivity()

	// Setup test data
	suite.setupTestData()
}

// TearDownSuite cleans up after the test suite
func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	// Clean up environment variables
	os.Unsetenv("APP_MODE")
	os.Unsetenv("JWT_SECRET")

	// Clean up test data
	suite.cleanupTestData()

	// Close database connections
	if suite.pgDB != nil {
		_ = suite.pgDB.Close()
	}
	if suite.chDB != nil {
		_ = suite.chDB.Close()
	}
}

// verifyDatabaseConnectivity checks that all databases are reachable
func (suite *DatabaseIntegrationTestSuite) verifyDatabaseConnectivity() {
	// Test PostgreSQL
	err := suite.pgDB.Ping()
	require.NoError(suite.T(), err, "Failed to connect to PostgreSQL")

	// Test ClickHouse
	err = suite.chDB.Ping()
	require.NoError(suite.T(), err, "Failed to connect to ClickHouse")

	suite.T().Log("Database connectivity verified")
}

// setupTestData creates test data for integration tests
func (suite *DatabaseIntegrationTestSuite) setupTestData() {
	suite.testProviderID = ulid.New()
	suite.testModelID = ulid.New()
	suite.testProviderConfigID = ulid.New()
}

// cleanupTestData removes test data after tests
func (suite *DatabaseIntegrationTestSuite) cleanupTestData() {
	// Clean up provider configurations
	_, _ = suite.pgDB.Exec("DELETE FROM provider_configs WHERE id = $1", suite.testProviderConfigID)

	// Clean up models
	_, _ = suite.pgDB.Exec("DELETE FROM models WHERE id = $1", suite.testModelID)

	// Clean up providers
	_, _ = suite.pgDB.Exec("DELETE FROM providers WHERE id = $1", suite.testProviderID)
}

// TestProviderRepository tests provider repository operations
func (suite *DatabaseIntegrationTestSuite) TestProviderRepository() {
	// Create test provider
	provider := &providers.Provider{
		ID:          suite.testProviderID,
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Type:        providers.ProviderTypeOpenAI,
		BaseURL:     "https://api.test-provider.com",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test Create
	err := suite.providerRepo.Create(suite.ctx, provider)
	require.NoError(suite.T(), err)

	// Test GetByID
	retrieved, err := suite.providerRepo.GetByID(suite.ctx, suite.testProviderID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), provider.ID, retrieved.ID)
	assert.Equal(suite.T(), provider.Name, retrieved.Name)
	assert.Equal(suite.T(), provider.DisplayName, retrieved.DisplayName)
	assert.Equal(suite.T(), provider.Type, retrieved.Type)

	// Test GetByName
	retrievedByName, err := suite.providerRepo.GetByName(suite.ctx, "test-provider")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), provider.ID, retrievedByName.ID)

	// Test List
	allProviders, err := suite.providerRepo.List(suite.ctx)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(allProviders), 0)

	// Find our test provider in the list
	found := false
	for _, p := range allProviders {
		if p.ID == suite.testProviderID {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Test provider not found in list")

	// Test ListActive
	activeProviders, err := suite.providerRepo.ListActive(suite.ctx)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(activeProviders), 0)

	// Test Update
	provider.DisplayName = "Updated Test Provider"
	provider.UpdatedAt = time.Now()
	err = suite.providerRepo.Update(suite.ctx, provider)
	require.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.providerRepo.GetByID(suite.ctx, suite.testProviderID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Test Provider", updated.DisplayName)

	// Test Delete
	err = suite.providerRepo.Delete(suite.ctx, suite.testProviderID)
	require.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.providerRepo.GetByID(suite.ctx, suite.testProviderID)
	assert.Error(suite.T(), err, "Provider should not exist after deletion")
}

// TestModelRepository tests model repository operations
func (suite *DatabaseIntegrationTestSuite) TestModelRepository() {
	// First create a provider for the model
	provider := &providers.Provider{
		ID:          suite.testProviderID,
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Type:        providers.ProviderTypeOpenAI,
		BaseURL:     "https://api.test-provider.com",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := suite.providerRepo.Create(suite.ctx, provider)
	require.NoError(suite.T(), err)

	// Create test model
	model := &providers.Model{
		ID:             suite.testModelID,
		ProviderID:     suite.testProviderID,
		Name:           "test-model",
		DisplayName:    "Test Model",
		Type:           providers.ModelTypeLLM,
		MaxTokens:      4096,
		ContextLength:  8192,
		InputCostPer1K: 0.001,
		OutputCostPer1K: 0.002,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test Create
	err = suite.modelRepo.Create(suite.ctx, model)
	require.NoError(suite.T(), err)

	// Test GetByID
	retrieved, err := suite.modelRepo.GetByID(suite.ctx, suite.testModelID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.ID, retrieved.ID)
	assert.Equal(suite.T(), model.Name, retrieved.Name)
	assert.Equal(suite.T(), model.DisplayName, retrieved.DisplayName)
	assert.Equal(suite.T(), model.ProviderID, retrieved.ProviderID)

	// Test GetByName
	retrievedByName, err := suite.modelRepo.GetByName(suite.ctx, "test-model")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.ID, retrievedByName.ID)

	// Test ListByProvider
	providerModels, err := suite.modelRepo.ListByProvider(suite.ctx, suite.testProviderID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), providerModels, 1)
	assert.Equal(suite.T(), suite.testModelID, providerModels[0].ID)

	// Test ListActive
	activeModels, err := suite.modelRepo.ListActive(suite.ctx)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(activeModels), 0)

	// Find our test model in the active list
	found := false
	for _, m := range activeModels {
		if m.ID == suite.testModelID {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Test model not found in active list")

	// Test Update
	model.DisplayName = "Updated Test Model"
	model.MaxTokens = 8192
	model.UpdatedAt = time.Now()
	err = suite.modelRepo.Update(suite.ctx, model)
	require.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.modelRepo.GetByID(suite.ctx, suite.testModelID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Test Model", updated.DisplayName)
	assert.Equal(suite.T(), int32(8192), updated.MaxTokens)

	// Test Delete
	err = suite.modelRepo.Delete(suite.ctx, suite.testModelID)
	require.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.modelRepo.GetByID(suite.ctx, suite.testModelID)
	assert.Error(suite.T(), err, "Model should not exist after deletion")
}

// TestProviderConfigRepository tests provider configuration repository operations
func (suite *DatabaseIntegrationTestSuite) TestProviderConfigRepository() {
	// First create a provider for the configuration
	provider := &providers.Provider{
		ID:          suite.testProviderID,
		Name:        "test-provider",
		DisplayName: "Test Provider",
		Type:        providers.ProviderTypeOpenAI,
		BaseURL:     "https://api.test-provider.com",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := suite.providerRepo.Create(suite.ctx, provider)
	require.NoError(suite.T(), err)

	// Create test organization ID
	testOrgID := ulid.New()

	// Create test provider configuration
	config := &providers.ProviderConfig{
		ID:                     suite.testProviderConfigID,
		ProviderID:            suite.testProviderID,
		OrganizationID:        testOrgID,
		Name:                  "test-config",
		APIKey:                "test-api-key-encrypted",
		BaseURL:               "https://custom.api.com",
		MaxRequestsPerMinute:  60,
		MaxTokensPerMinute:    100000,
		TimeoutSeconds:        30,
		IsActive:              true,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Test Create
	err = suite.providerConfigRepo.Create(suite.ctx, config)
	require.NoError(suite.T(), err)

	// Test GetByID
	retrieved, err := suite.providerConfigRepo.GetByID(suite.ctx, suite.testProviderConfigID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), config.ID, retrieved.ID)
	assert.Equal(suite.T(), config.Name, retrieved.Name)
	assert.Equal(suite.T(), config.ProviderID, retrieved.ProviderID)
	assert.Equal(suite.T(), config.OrganizationID, retrieved.OrganizationID)

	// Test GetByOrganizationAndProvider
	orgConfigs, err := suite.providerConfigRepo.GetByOrganizationAndProvider(
		suite.ctx, testOrgID, suite.testProviderID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), orgConfigs, 1)
	assert.Equal(suite.T(), suite.testProviderConfigID, orgConfigs[0].ID)

	// Test ListByOrganization
	allOrgConfigs, err := suite.providerConfigRepo.ListByOrganization(suite.ctx, testOrgID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), allOrgConfigs, 1)

	// Test ListByProvider
	providerConfigs, err := suite.providerConfigRepo.ListByProvider(suite.ctx, suite.testProviderID)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(providerConfigs), 0)

	// Test Update
	config.Name = "updated-test-config"
	config.MaxRequestsPerMinute = 120
	config.UpdatedAt = time.Now()
	err = suite.providerConfigRepo.Update(suite.ctx, config)
	require.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.providerConfigRepo.GetByID(suite.ctx, suite.testProviderConfigID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updated-test-config", updated.Name)
	assert.Equal(suite.T(), int32(120), updated.MaxRequestsPerMinute)

	// Test Delete
	err = suite.providerConfigRepo.Delete(suite.ctx, suite.testProviderConfigID)
	require.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.providerConfigRepo.GetByID(suite.ctx, suite.testProviderConfigID)
	assert.Error(suite.T(), err, "Provider config should not exist after deletion")
}

// TestDatabaseTransactions tests transaction support
func (suite *DatabaseIntegrationTestSuite) TestDatabaseTransactions() {
	// Begin a transaction
	tx, err := suite.pgDB.BeginTx(suite.ctx, nil)
	require.NoError(suite.T(), err)

	// Create a test provider within the transaction
	testID := ulid.New()
	_, err = tx.Exec(`
		INSERT INTO providers (id, name, display_name, type, base_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, testID, "tx-test-provider", "TX Test Provider", "openai", "https://api.test.com", true, time.Now(), time.Now())
	require.NoError(suite.T(), err)

	// Verify the provider exists within the transaction
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM providers WHERE id = $1", testID).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)

	// Rollback the transaction
	err = tx.Rollback()
	require.NoError(suite.T(), err)

	// Verify the provider does not exist outside the transaction
	err = suite.pgDB.QueryRow("SELECT COUNT(*) FROM providers WHERE id = $1", testID).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count, "Provider should not exist after transaction rollback")

	// Test successful transaction
	tx, err = suite.pgDB.BeginTx(suite.ctx, nil)
	require.NoError(suite.T(), err)

	_, err = tx.Exec(`
		INSERT INTO providers (id, name, display_name, type, base_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, testID, "tx-test-provider", "TX Test Provider", "openai", "https://api.test.com", true, time.Now(), time.Now())
	require.NoError(suite.T(), err)

	// Commit the transaction
	err = tx.Commit()
	require.NoError(suite.T(), err)

	// Verify the provider exists after commit
	err = suite.pgDB.QueryRow("SELECT COUNT(*) FROM providers WHERE id = $1", testID).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count, "Provider should exist after transaction commit")

	// Clean up
	_, _ = suite.pgDB.Exec("DELETE FROM providers WHERE id = $1", testID)
}

// TestClickHouseOperations tests ClickHouse specific operations
func (suite *DatabaseIntegrationTestSuite) TestClickHouseOperations() {
	// Test basic connectivity and table creation
	testTableName := "test_metrics_" + ulid.New().String()

	// Create a test table
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id String,
			timestamp DateTime64(3),
			metric_name String,
			value Float64
		) ENGINE = MergeTree()
		ORDER BY (timestamp, id)
		TTL timestamp + INTERVAL 7 DAY DELETE
	`, testTableName)

	_, err := suite.chDB.Exec(createTableSQL)
	require.NoError(suite.T(), err)

	// Insert test data
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (id, timestamp, metric_name, value) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?)
	`, testTableName)

	now := time.Now()
	_, err = suite.chDB.Exec(insertSQL,
		"test-1", now, "request_count", 100.0,
		"test-2", now.Add(-1*time.Hour), "request_count", 50.0,
		"test-3", now.Add(-2*time.Hour), "error_count", 5.0,
	)
	require.NoError(suite.T(), err)

	// Query data
	querySQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", testTableName)
	var count int
	err = suite.chDB.QueryRow(querySQL).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)

	// Test aggregation
	aggregationSQL := fmt.Sprintf(`
		SELECT metric_name, SUM(value) as total
		FROM %s
		GROUP BY metric_name
		ORDER BY metric_name
	`, testTableName)

	rows, err := suite.chDB.Query(aggregationSQL)
	require.NoError(suite.T(), err)
	defer rows.Close()

	results := make(map[string]float64)
	for rows.Next() {
		var metricName string
		var total float64
		err = rows.Scan(&metricName, &total)
		require.NoError(suite.T(), err)
		results[metricName] = total
	}

	assert.Equal(suite.T(), 5.0, results["error_count"])
	assert.Equal(suite.T(), 150.0, results["request_count"])

	// Clean up test table
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", testTableName)
	_, err = suite.chDB.Exec(dropSQL)
	require.NoError(suite.T(), err)
}

// TestConnectionPooling tests database connection pooling
func (suite *DatabaseIntegrationTestSuite) TestConnectionPooling() {
	// Test multiple concurrent connections
	const numConcurrent = 10
	ch := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			// Each goroutine performs a database operation
			var count int
			err := suite.pgDB.QueryRow("SELECT COUNT(*) FROM providers").Scan(&count)
			ch <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numConcurrent; i++ {
		err := <-ch
		assert.NoError(suite.T(), err, "Concurrent database operation failed")
	}
}

// TestDatabaseHealth tests database health checks
func (suite *DatabaseIntegrationTestSuite) TestDatabaseHealth() {
	// Test PostgreSQL health
	err := suite.pgDB.Ping()
	assert.NoError(suite.T(), err, "PostgreSQL health check failed")

	// Test ClickHouse health
	err = suite.chDB.Ping()
	assert.NoError(suite.T(), err, "ClickHouse health check failed")

	// Test PostgreSQL with timeout context
	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()

	err = suite.pgDB.PingContext(ctx)
	assert.NoError(suite.T(), err, "PostgreSQL health check with timeout failed")

	// Test ClickHouse with timeout context
	ctx2, cancel2 := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel2()

	err = suite.chDB.PingContext(ctx2)
	assert.NoError(suite.T(), err, "ClickHouse health check with timeout failed")
}

// TestMigrationState tests the current migration state
func (suite *DatabaseIntegrationTestSuite) TestMigrationState() {
	// Check if migration tables exist
	var exists bool

	// PostgreSQL migration table
	err := suite.pgDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), exists, "PostgreSQL migration table should exist")

	// Check some expected tables exist
	expectedTables := []string{"providers", "models", "provider_configs"}
	for _, tableName := range expectedTables {
		err = suite.pgDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, tableName).Scan(&exists)
		require.NoError(suite.T(), err)
		assert.True(suite.T(), exists, "Table %s should exist", tableName)
	}

	// TODO: Add ClickHouse migration table checks when available
}

// TestDatabaseIntegrationSuite runs the complete database integration test suite
func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}