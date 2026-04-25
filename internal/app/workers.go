package app

import (
	"time"

	"brokle/internal/core/services/observability"
	"brokle/internal/workers"
	annotationWorker "brokle/internal/workers/annotation"
	evaluationWorker "brokle/internal/workers/evaluation"
	"brokle/pkg/uid"
)

// ProvideWorkers wires the background-worker pool that drains Redis
// streams (telemetry / evaluator / evaluation / manual-trigger) and
// runs the periodic billing aggregator, contract expiry, and
// annotation lock expiry jobs.
//
// Each worker takes a unique consumer ID derived from a UUIDv7 so a
// horizontally scaled worker fleet stays disjoint on the Redis-stream
// consumer groups.
func ProvideWorkers(core *CoreContainer) (*WorkerContainer, error) {
	deduplicationService := observability.NewTelemetryDeduplicationService(
		core.Repos.Observability.TelemetryDeduplication,
	)

	consumerConfig := &workers.TelemetryStreamConsumerConfig{
		ConsumerGroup:     "telemetry-workers",
		ConsumerID:        "worker-" + uid.New().String(),
		BatchSize:         50,
		BlockDuration:     time.Second,
		MaxRetries:        3,
		RetryBackoff:      500 * time.Millisecond,
		DiscoveryInterval: 30 * time.Second,
		MaxStreamsPerRead: 10,
	}

	telemetryConsumer := workers.NewTelemetryStreamConsumer(
		core.Databases.Redis,
		deduplicationService,
		core.Logger,
		consumerConfig,
		core.Services.Observability.TraceService,
		core.Services.Observability.ScoreService,
		core.Services.Observability.MetricsService,
		core.Services.Observability.LogsService,
		core.Services.Observability.GenAIEventsService,
		core.Services.Observability.ArchiveService, // S3 raw telemetry archival (nil if disabled)
		&core.Config.Archive,
	)

	// Evaluator worker — config-driven cadence + cache TTL with sane fallbacks.
	discoveryInterval, _ := time.ParseDuration(core.Config.Workers.EvaluatorWorker.DiscoveryInterval)
	if discoveryInterval == 0 {
		discoveryInterval = 30 * time.Second
	}
	evaluatorCacheTTL, _ := time.ParseDuration(core.Config.Workers.EvaluatorWorker.EvaluatorCacheTTL)
	if evaluatorCacheTTL == 0 {
		evaluatorCacheTTL = 30 * time.Second
	}

	evaluatorWorkerConfig := &evaluationWorker.EvaluatorWorkerConfig{
		ConsumerGroup:     "evaluator-workers",
		ConsumerID:        "evaluator-worker-" + uid.New().String(),
		BatchSize:         core.Config.Workers.EvaluatorWorker.BatchSize,
		BlockDuration:     time.Duration(core.Config.Workers.EvaluatorWorker.BlockDurationMs) * time.Millisecond,
		MaxRetries:        core.Config.Workers.EvaluatorWorker.MaxRetries,
		RetryBackoff:      time.Duration(core.Config.Workers.EvaluatorWorker.RetryBackoffMs) * time.Millisecond,
		DiscoveryInterval: discoveryInterval,
		MaxStreamsPerRead: core.Config.Workers.EvaluatorWorker.MaxStreamsPerRead,
		EvaluatorCacheTTL: evaluatorCacheTTL,
	}

	evaluatorWorkerInstance := evaluationWorker.NewEvaluatorWorker(
		core.Databases.Redis,
		core.Services.Evaluation.Evaluator,
		core.Services.Evaluation.EvaluatorExecution,
		core.Logger,
		evaluatorWorkerConfig,
	)

	// Built-in scorers (regex, builtin) plus the optional LLM scorer
	// gated on credentials + prompt services being available.
	builtinScorer := evaluationWorker.NewBuiltinScorer(core.Logger)
	regexScorer := evaluationWorker.NewRegexScorer(core.Logger)

	var llmScorer evaluationWorker.Scorer
	if core.Services.Credentials != nil && core.Services.Prompt != nil {
		llmScorer = evaluationWorker.NewLLMScorer(
			core.Services.Credentials.ProviderCredential,
			core.Services.Prompt.Execution,
			core.Logger,
		)
		core.Logger.Info("LLM scorer initialized for evaluation worker")
	} else {
		core.Logger.Warn("LLM scorer disabled: credentials or prompt services not available")
	}

	evalWorkerConfig := &evaluationWorker.EvaluationWorkerConfig{
		ConsumerGroup:  "evaluation-execution-workers",
		ConsumerID:     "eval-worker-" + uid.New().String(),
		BatchSize:      10,
		BlockDuration:  time.Second,
		MaxRetries:     3,
		RetryBackoff:   500 * time.Millisecond,
		MaxConcurrency: 5,
	}

	evalWorker := evaluationWorker.NewEvaluationWorker(
		core.Databases.Redis,
		core.Services.Observability.ScoreService,
		core.Services.Evaluation.EvaluatorExecution,
		llmScorer,
		builtinScorer,
		regexScorer,
		core.Logger,
		evalWorkerConfig,
	)

	manualTriggerWorkerConfig := &evaluationWorker.ManualTriggerWorkerConfig{
		ConsumerGroup:  "manual-trigger-workers",
		ConsumerID:     "manual-trigger-" + uid.New().String(),
		BlockDuration:  time.Second,
		MaxRetries:     3,
		RetryBackoff:   500 * time.Millisecond,
		MaxConcurrency: 3,
	}

	manualTriggerWorker := evaluationWorker.NewManualTriggerWorker(
		core.Databases.Redis,
		core.Services.Observability.TraceService,
		core.Services.Evaluation.EvaluatorExecution,
		core.Logger,
		manualTriggerWorkerConfig,
	)

	// Usage aggregation worker — periodic ClickHouse → PostgreSQL sync
	// with atomic billing + budget updates.
	usageAggWorker := workers.NewUsageAggregationWorker(
		core.Config,
		core.Logger,
		core.Transactor,
		core.Repos.Billing.BillableUsage,
		core.Repos.Billing.OrganizationBilling,
		core.Repos.Billing.UsageBudget,
		core.Repos.Billing.UsageAlert,
		core.Repos.Organization.Organization,
		core.Services.Billing.Pricing,
		nil, // NotificationWorker — wire when email notifications land.
	)

	// Daily contract-expiration job: expires contracts past end_date.
	contractExpWorker := workers.NewContractExpirationWorker(
		core.Config,
		core.Logger,
		core.Services.Billing.Contract,
		core.Repos.Billing.OrganizationBilling,
	)

	// Per-minute annotation lock-expiry job: releases stale per-item
	// reviewer locks so a paused/abandoned reviewer doesn't pin items.
	lockExpiryWorker := annotationWorker.NewLockExpiryWorker(
		core.Logger,
		core.Repos.Annotation.Queue,
		core.Repos.Annotation.Item,
	)

	return &WorkerContainer{
		TelemetryConsumer:        telemetryConsumer,
		EvaluatorWorker:          evaluatorWorkerInstance,
		EvaluationWorker:         evalWorker,
		ManualTriggerWorker:      manualTriggerWorker,
		UsageAggregationWorker:   usageAggWorker,
		ContractExpirationWorker: contractExpWorker,
		LockExpiryWorker:         lockExpiryWorker,
	}, nil
}
