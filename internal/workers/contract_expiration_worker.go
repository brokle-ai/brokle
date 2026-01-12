package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"brokle/internal/config"
	"brokle/internal/core/domain/billing"
)

// ContractExpirationWorker expires contracts that have passed their end date
type ContractExpirationWorker struct {
	config          *config.Config
	logger          *slog.Logger
	contractService billing.ContractService
	billingRepo     billing.OrganizationBillingRepository
	quit            chan struct{}
	wg              sync.WaitGroup
	ticker          *time.Ticker
}

// NewContractExpirationWorker creates a new contract expiration worker
func NewContractExpirationWorker(
	config *config.Config,
	logger *slog.Logger,
	contractService billing.ContractService,
	billingRepo billing.OrganizationBillingRepository,
) *ContractExpirationWorker {
	return &ContractExpirationWorker{
		config:          config,
		logger:          logger,
		contractService: contractService,
		billingRepo:     billingRepo,
		quit:            make(chan struct{}),
	}
}

// Start starts the contract expiration worker
func (w *ContractExpirationWorker) Start() {
	w.logger.Info("Starting contract expiration worker")

	// Start the main loop goroutine
	w.wg.Add(1)
	go w.mainLoop()
}

// Stop stops the contract expiration worker and waits for graceful shutdown
func (w *ContractExpirationWorker) Stop() {
	w.logger.Info("Stopping contract expiration worker")
	close(w.quit)
	w.wg.Wait()
}

// mainLoop handles the worker lifecycle: immediate run, wait until midnight, then daily runs
func (w *ContractExpirationWorker) mainLoop() {
	defer w.wg.Done()

	// Run immediately on start
	w.run()

	// Calculate time until next midnight UTC
	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	timeUntilMidnight := nextMidnight.Sub(now)

	// Wait until first midnight with quit check
	select {
	case <-time.After(timeUntilMidnight):
		w.run()
	case <-w.quit:
		w.logger.Info("Contract expiration worker stopped during initial wait")
		return
	}

	// Then run every 24 hours
	w.ticker = time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-w.ticker.C:
			w.run()
		case <-w.quit:
			w.ticker.Stop()
			w.logger.Info("Contract expiration worker stopped")
			return
		}
	}
}

// run executes a single expiration check cycle
func (w *ContractExpirationWorker) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	w.logger.Info("Starting contract expiration check")
	startTime := time.Now()

	// Get contracts expiring now or earlier (timestamp-based check)
	// Example: Worker runs Jan 9 00:00, contract expires Jan 8 10:15
	// Result: Found (10:15 yesterday <= now) ✓
	// Max delay: 24 hours (acceptable for enterprise contracts)
	contracts, err := w.contractService.GetExpiringContracts(ctx, 0)
	if err != nil {
		w.logger.Error("failed to get expiring contracts", "error", err)
		return
	}

	var expiredCount, failedCount int

	for _, contract := range contracts {
		w.logger.Info("Expiring contract",
			"contract_id", contract.ID,
			"organization_id", contract.OrganizationID,
			"contract_name", contract.ContractName,
			"end_date", contract.EndDate,
		)

		// Expire the contract
		if err := w.contractService.ExpireContract(ctx, contract.ID); err != nil {
			w.logger.Error("failed to expire contract",
				"error", err,
				"contract_id", contract.ID,
				"organization_id", contract.OrganizationID,
			)
			failedCount++
			continue
		}

		expiredCount++

		w.logger.Info("contract expired successfully",
			"contract_id", contract.ID,
			"organization_id", contract.OrganizationID,
		)
	}

	duration := time.Since(startTime)
	w.logger.Info("Contract expiration check completed",
		"contracts_expired", expiredCount,
		"contracts_failed", failedCount,
		"duration_ms", duration.Milliseconds(),
	)
}
