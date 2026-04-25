// Package app composes the application's dependency graph.
//
// Files in this package each own one slice of the wiring:
//
//   - containers.go        — top-level + per-domain container struct types
//   - repositories.go      — Provide<Domain>Repositories ×13 + ProvideRepositories
//   - services_<domain>.go — Provide<Domain>Services for each of 12 domains
//   - services.go          — ProvideServerServices + ProvideWorkerServices orchestrators
//   - workers.go           — ProvideWorkers (background-worker pool wiring)
//   - server.go            — ProvideServer (HTTP + gRPC handler bootstrap)
//   - lifecycle.go         — HealthCheck, Shutdown, createEmailSender
//   - providers.go (this)  — ProvideDatabases, ProvideCore, ProvideEnterpriseServices
//
// The split mirrors the per-domain organization already used in
// internal/core/services/<domain>/, internal/infrastructure/repository/<domain>/
// and internal/transport/http/handlers/<domain>/. Adding a new
// service to one domain touches a single file (services_<domain>.go);
// adding a new domain adds one services_<domain>.go + one
// Provide<Domain>Repositories block in repositories.go.
package app

import (
	"context"
	"log/slog"

	"brokle/internal/config"
	eeAnalytics "brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/db"
)

// ProvideDatabases opens the three database backends Brokle owns
// (PostgreSQL pool, Redis, ClickHouse) and binds the TxManager to
// the pool. Failure of any backend short-circuits the others;
// already-opened resources close before return.
func ProvideDatabases(cfg *config.Config, logger *slog.Logger) (*DatabaseContainer, error) {
	pool, err := db.NewPool(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	redis, err := database.NewRedisDB(cfg, logger)
	if err != nil {
		pool.Close()
		return nil, err
	}

	clickhouse, err := database.NewClickHouseDB(cfg, logger)
	if err != nil {
		_ = redis.Close()
		pool.Close()
		return nil, err
	}

	return &DatabaseContainer{
		Pool:       pool,
		TxManager:  db.NewTxManager(pool),
		Redis:      redis,
		ClickHouse: clickhouse,
	}, nil
}

// ProvideCore builds the cross-mode foundation — databases, repos,
// transactor — that both server and worker entrypoints share.
// Services and Enterprise are deliberately left nil; the caller
// (cmd/server, cmd/worker) populates them via ProvideServerServices
// or ProvideWorkerServices and ProvideEnterpriseServices in mode-
// specific bootstrap.
func ProvideCore(cfg *config.Config, logger *slog.Logger) (*CoreContainer, error) {
	databases, err := ProvideDatabases(cfg, logger)
	if err != nil {
		return nil, err
	}

	repos := ProvideRepositories(databases, logger)

	// The TxManager satisfies common.Transactor; services that need
	// cross-aggregate writes (registration, annotation) take the
	// transactor interface, and the sqlc-backed TxManager binds it.
	transactor := databases.TxManager

	return &CoreContainer{
		Config:     cfg,
		Logger:     logger,
		Databases:  databases,
		Repos:      repos,
		Transactor: transactor,
		Services:   nil, // populated by mode-specific provider
		Enterprise: nil, // populated by mode-specific provider
	}, nil
}

// ProvideEnterpriseServices binds the enterprise-feature
// implementations. The constructors return either real or stub
// implementations based on Go build tags (see internal/ee/).
func ProvideEnterpriseServices(cfg *config.Config, logger *slog.Logger) *EnterpriseContainer {
	return &EnterpriseContainer{
		SSO:        sso.New(),
		RBAC:       rbac.New(),
		Compliance: compliance.New(),
		Analytics:  eeAnalytics.New(),
	}
}
