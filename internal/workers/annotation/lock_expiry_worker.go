package annotation

import (
	"context"
	"log/slog"
	"sync"
	"time"

	annotationDomain "brokle/internal/core/domain/annotation"
)

// LockExpiryWorker releases expired locks on annotation queue items.
// Items are locked when claimed but locks expire after the queue's configured timeout
// (default 5 minutes) to prevent items being stuck if annotators abandon work.
type LockExpiryWorker struct {
	logger    *slog.Logger
	queueRepo annotationDomain.QueueRepository
	itemRepo  annotationDomain.ItemRepository
	quit      chan struct{}
	wg        sync.WaitGroup
	ticker    *time.Ticker
	interval  time.Duration
}

// NewLockExpiryWorker creates a new lock expiry worker.
func NewLockExpiryWorker(
	logger *slog.Logger,
	queueRepo annotationDomain.QueueRepository,
	itemRepo annotationDomain.ItemRepository,
) *LockExpiryWorker {
	return &LockExpiryWorker{
		logger:    logger,
		queueRepo: queueRepo,
		itemRepo:  itemRepo,
		quit:      make(chan struct{}),
		interval:  1 * time.Minute, // Run every minute
	}
}

// Start starts the lock expiry worker.
func (w *LockExpiryWorker) Start() {
	w.logger.Info("Starting annotation lock expiry worker", "interval", w.interval)

	w.wg.Add(1)
	go w.mainLoop()
}

// Stop stops the lock expiry worker and waits for graceful shutdown.
func (w *LockExpiryWorker) Stop() {
	w.logger.Info("Stopping annotation lock expiry worker")
	close(w.quit)
	w.wg.Wait()
}

// mainLoop runs the periodic lock expiry check.
func (w *LockExpiryWorker) mainLoop() {
	defer w.wg.Done()

	// Run immediately on start
	w.run()

	// Then run at regular intervals
	w.ticker = time.NewTicker(w.interval)
	for {
		select {
		case <-w.ticker.C:
			w.run()
		case <-w.quit:
			w.ticker.Stop()
			w.logger.Info("Annotation lock expiry worker stopped")
			return
		}
	}
}

// run executes a single lock expiry cycle.
func (w *LockExpiryWorker) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()

	// Get all active queues across all projects
	queues, err := w.getAllActiveQueues(ctx)
	if err != nil {
		w.logger.Error("failed to get active queues for lock expiry", "error", err)
		return
	}

	if len(queues) == 0 {
		return // No active queues, nothing to do
	}

	var totalReleased int64
	var failedQueues int

	for _, queue := range queues {
		lockTimeout := queue.Settings.LockTimeoutSeconds
		if lockTimeout <= 0 {
			lockTimeout = 300 // Default 5 minutes
		}

		released, err := w.itemRepo.ReleaseExpiredLocks(ctx, queue.ID, lockTimeout)
		if err != nil {
			w.logger.Error("failed to release expired locks for queue",
				"error", err,
				"queue_id", queue.ID,
				"project_id", queue.ProjectID,
			)
			failedQueues++
			continue
		}

		if released > 0 {
			w.logger.Info("released expired locks",
				"queue_id", queue.ID,
				"project_id", queue.ProjectID,
				"released", released,
			)
		}
		totalReleased += released
	}

	duration := time.Since(startTime)

	if totalReleased > 0 || failedQueues > 0 {
		w.logger.Info("lock expiry cycle completed",
			"queues_processed", len(queues),
			"locks_released", totalReleased,
			"failed_queues", failedQueues,
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// getAllActiveQueues retrieves all active queues across all projects.
// This is a worker-specific operation that bypasses project scoping.
func (w *LockExpiryWorker) getAllActiveQueues(ctx context.Context) ([]*annotationDomain.AnnotationQueue, error) {
	// The queue repository's List method requires a projectID.
	// For the worker, we need to add a method to list all active queues.
	// Since we're adding this to the worker, we'll cast to the concrete repository
	// that implements an additional method.

	// Check if the repository implements ListAllActive
	if allLister, ok := w.queueRepo.(AllActiveQueuesLister); ok {
		return allLister.ListAllActive(ctx)
	}

	// Fallback: log warning and return empty
	w.logger.Warn("queue repository does not implement ListAllActive, lock expiry will not run")
	return nil, nil
}

// AllActiveQueuesLister is an optional interface for repositories that can list all active queues.
type AllActiveQueuesLister interface {
	ListAllActive(ctx context.Context) ([]*annotationDomain.AnnotationQueue, error)
}
