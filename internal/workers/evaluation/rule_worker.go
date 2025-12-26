package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/streams"
	"brokle/pkg/ulid"
)

const (
	evaluationJobsStream = "evaluation:jobs"
	ruleCacheTTL         = 30 * time.Second
)

// EvaluationJob represents a matched span-rule pair to be processed by EvaluationWorker
type EvaluationJob struct {
	JobID       ulid.ULID                `json:"job_id"`
	RuleID      ulid.ULID                `json:"rule_id"`
	ProjectID   ulid.ULID                `json:"project_id"`
	SpanData    map[string]interface{}   `json:"span_data"`
	TraceID     string                   `json:"trace_id"`
	SpanID      string                   `json:"span_id"`
	ScorerType  evaluation.ScorerType    `json:"scorer_type"`
	ScorerConfig map[string]any          `json:"scorer_config"`
	Variables   map[string]string        `json:"variables"` // Extracted variables from span
	CreatedAt   time.Time                `json:"created_at"`
}

// RuleWorkerConfig holds configuration for the rule worker
type RuleWorkerConfig struct {
	ConsumerGroup     string
	ConsumerID        string
	BatchSize         int
	BlockDuration     time.Duration
	MaxRetries        int
	RetryBackoff      time.Duration
	DiscoveryInterval time.Duration
	MaxStreamsPerRead int
	RuleCacheTTL      time.Duration
}

// RuleCache provides thread-safe caching of active rules per project
type RuleCache struct {
	cache     map[string]ruleCacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
}

type ruleCacheEntry struct {
	rules     []*evaluation.EvaluationRule
	expiresAt time.Time
}

func NewRuleCache(ttl time.Duration) *RuleCache {
	return &RuleCache{
		cache: make(map[string]ruleCacheEntry),
		ttl:   ttl,
	}
}

func (c *RuleCache) Get(projectID string) ([]*evaluation.EvaluationRule, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[projectID]
	if !exists {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.rules, true
}

func (c *RuleCache) Set(projectID string, rules []*evaluation.EvaluationRule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[projectID] = ruleCacheEntry{
		rules:     rules,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *RuleCache) Invalidate(projectID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, projectID)
}

// RuleWorker consumes spans from telemetry streams, matches against active rules, and emits evaluation jobs
type RuleWorker struct {
	redis             *database.RedisDB
	ruleService       evaluation.RuleService
	ruleCache         *RuleCache
	logger            *slog.Logger

	// Consumer configuration
	consumerGroup     string
	consumerID        string
	batchSize         int
	blockDuration     time.Duration
	maxRetries        int
	retryBackoff      time.Duration
	discoveryInterval time.Duration
	maxStreamsPerRead int

	// State management
	activeStreams       map[string]bool
	streamsMutex        sync.RWMutex
	streamRotation      int
	quit                chan struct{}
	wg                  sync.WaitGroup
	running             int64
	discoveryBackoff    time.Duration
	maxDiscoveryBackoff time.Duration

	// Metrics
	spansProcessed      int64
	rulesMatched        int64
	jobsEmitted         int64
	errorsCount         int64
}

// NewRuleWorker creates a new rule worker
func NewRuleWorker(
	redis *database.RedisDB,
	ruleService evaluation.RuleService,
	logger *slog.Logger,
	config *RuleWorkerConfig,
) *RuleWorker {
	if config == nil {
		config = &RuleWorkerConfig{
			ConsumerGroup:     "evaluation-rule-workers",
			ConsumerID:        "rule-worker-" + ulid.New().String(),
			BatchSize:         50,
			BlockDuration:     time.Second,
			MaxRetries:        3,
			RetryBackoff:      500 * time.Millisecond,
			DiscoveryInterval: 30 * time.Second,
			MaxStreamsPerRead: 10,
			RuleCacheTTL:      ruleCacheTTL,
		}
	}

	return &RuleWorker{
		redis:               redis,
		ruleService:         ruleService,
		ruleCache:           NewRuleCache(config.RuleCacheTTL),
		logger:              logger,
		consumerGroup:       config.ConsumerGroup,
		consumerID:          config.ConsumerID,
		batchSize:           config.BatchSize,
		blockDuration:       config.BlockDuration,
		maxRetries:          config.MaxRetries,
		retryBackoff:        config.RetryBackoff,
		discoveryInterval:   config.DiscoveryInterval,
		maxStreamsPerRead:   config.MaxStreamsPerRead,
		activeStreams:       make(map[string]bool),
		quit:                make(chan struct{}),
		discoveryBackoff:    time.Second,
		maxDiscoveryBackoff: 30 * time.Second,
	}
}

// Start begins the rule worker
func (w *RuleWorker) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt64(&w.running, 0, 1) {
		return errors.New("rule worker already running")
	}

	w.logger.Info("Starting rule worker",
		"consumer_group", w.consumerGroup,
		"consumer_id", w.consumerID,
		"batch_size", w.batchSize,
		"discovery_interval", w.discoveryInterval,
	)

	// Start consumption loop
	w.wg.Add(1)
	go w.consumeLoop(ctx)

	// Start stream discovery loop
	w.wg.Add(1)
	go w.discoveryLoop(ctx)

	return nil
}

// Stop gracefully stops the rule worker
func (w *RuleWorker) Stop() {
	if !atomic.CompareAndSwapInt64(&w.running, 1, 0) {
		return
	}

	w.logger.Info("Stopping rule worker")
	close(w.quit)
	w.wg.Wait()

	w.logger.Info("Rule worker stopped",
		"spans_processed", atomic.LoadInt64(&w.spansProcessed),
		"rules_matched", atomic.LoadInt64(&w.rulesMatched),
		"jobs_emitted", atomic.LoadInt64(&w.jobsEmitted),
		"errors_count", atomic.LoadInt64(&w.errorsCount),
	)
}

// discoveryLoop periodically discovers telemetry streams
func (w *RuleWorker) discoveryLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.discoveryInterval)
	defer ticker.Stop()

	// Initial discovery
	if err := w.performDiscovery(ctx); err != nil {
		w.logger.Error("Initial stream discovery failed", "error", err)
	}

	for {
		select {
		case <-w.quit:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.performDiscovery(ctx); err != nil {
				w.logger.Error("Stream discovery failed", "error", err, "backoff", w.discoveryBackoff)
				time.Sleep(w.discoveryBackoff)
				w.discoveryBackoff = minDuration(w.discoveryBackoff*2, w.maxDiscoveryBackoff)
			} else {
				w.discoveryBackoff = time.Second
			}
		}
	}
}

func (w *RuleWorker) performDiscovery(ctx context.Context) error {
	streams, err := w.discoverStreams(ctx)
	if err != nil {
		return err
	}

	if len(streams) == 0 {
		w.logger.Debug("No telemetry streams discovered")
		return nil
	}

	w.cleanupInactiveStreams(streams)
	return w.ensureConsumerGroups(ctx, streams)
}

func (w *RuleWorker) discoverStreams(ctx context.Context) ([]string, error) {
	var allStreams []string
	cursor := uint64(0)
	pattern := "telemetry:batches:*"

	for {
		keys, nextCursor, err := w.redis.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan streams: %w", err)
		}

		allStreams = append(allStreams, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	w.logger.Debug("Discovered telemetry streams", "stream_count", len(allStreams))
	return allStreams, nil
}

func (w *RuleWorker) ensureConsumerGroups(ctx context.Context, streamKeys []string) error {
	for _, streamKey := range streamKeys {
		w.streamsMutex.RLock()
		exists := w.activeStreams[streamKey]
		w.streamsMutex.RUnlock()

		if exists {
			continue
		}

		// Create consumer group (use "$" to only read new messages, not historical)
		err := w.redis.Client.XGroupCreateMkStream(ctx, streamKey, w.consumerGroup, "$").Err()
		if err != nil {
			if !strings.Contains(err.Error(), "BUSYGROUP") {
				w.logger.Warn("Failed to create consumer group", "error", err, "stream", streamKey)
				continue
			}
		}

		w.streamsMutex.Lock()
		w.activeStreams[streamKey] = true
		w.streamsMutex.Unlock()

		w.logger.Debug("Consumer group initialized", "stream", streamKey, "consumer_group", w.consumerGroup)
	}

	return nil
}

func (w *RuleWorker) cleanupInactiveStreams(discoveredStreams []string) {
	currentStreams := make(map[string]bool, len(discoveredStreams))
	for _, streamKey := range discoveredStreams {
		currentStreams[streamKey] = true
	}

	w.streamsMutex.Lock()
	defer w.streamsMutex.Unlock()

	var removedStreams []string
	for streamKey := range w.activeStreams {
		if !currentStreams[streamKey] {
			delete(w.activeStreams, streamKey)
			removedStreams = append(removedStreams, streamKey)
		}
	}

	if len(removedStreams) > 0 {
		w.logger.Info("Cleaned up inactive streams", "removed_count", len(removedStreams))
	}
}

// consumeLoop is the main consumption loop
func (w *RuleWorker) consumeLoop(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-w.quit:
			return
		case <-ctx.Done():
			return
		default:
			if err := w.consumeBatch(ctx); err != nil {
				if err != redis.Nil {
					w.logger.Error("Error consuming batch", "error", err)
					atomic.AddInt64(&w.errorsCount, 1)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (w *RuleWorker) consumeBatch(ctx context.Context) error {
	w.streamsMutex.Lock()
	var allStreamKeys []string
	for streamKey := range w.activeStreams {
		allStreamKeys = append(allStreamKeys, streamKey)
	}

	if len(allStreamKeys) > 0 && w.streamRotation >= len(allStreamKeys) {
		w.streamRotation = 0
	}

	if w.streamRotation > 0 && len(allStreamKeys) > w.streamRotation {
		allStreamKeys = append(allStreamKeys[w.streamRotation:], allStreamKeys[:w.streamRotation]...)
	}
	w.streamsMutex.Unlock()

	if len(allStreamKeys) == 0 {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	streamKeys := allStreamKeys
	if len(streamKeys) > w.maxStreamsPerRead {
		streamKeys = streamKeys[:w.maxStreamsPerRead]
	}

	// Build XReadGroup arguments
	streamArgs := make([]string, 0, len(streamKeys)*2)
	for _, streamKey := range streamKeys {
		streamArgs = append(streamArgs, streamKey)
	}
	for range streamKeys {
		streamArgs = append(streamArgs, ">")
	}

	results, err := w.redis.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.consumerGroup,
		Consumer: w.consumerID,
		Streams:  streamArgs,
		Count:    int64(w.batchSize),
		Block:    w.blockDuration,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	// Process messages from all streams
	for _, stream := range results {
		for _, msg := range stream.Messages {
			if err := w.processMessage(ctx, stream.Stream, msg); err != nil {
				w.logger.Error("Failed to process message", "error", err, "stream", stream.Stream, "message_id", msg.ID)
				atomic.AddInt64(&w.errorsCount, 1)
			}

			// Always acknowledge - rule matching is best-effort
			if ackErr := w.redis.Client.XAck(ctx, stream.Stream, w.consumerGroup, msg.ID).Err(); ackErr != nil {
				w.logger.Warn("Failed to acknowledge message", "error", ackErr, "stream", stream.Stream, "message_id", msg.ID)
			}
		}
	}

	w.streamsMutex.Lock()
	w.streamRotation += w.maxStreamsPerRead
	w.streamsMutex.Unlock()

	return nil
}

func (w *RuleWorker) processMessage(ctx context.Context, streamKey string, msg redis.XMessage) error {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		return errors.New("invalid message format: missing data field")
	}

	var batch streams.TelemetryStreamMessage
	if err := json.Unmarshal([]byte(dataStr), &batch); err != nil {
		return fmt.Errorf("failed to unmarshal batch data: %w", err)
	}

	// Get active rules for this project
	rules, err := w.getActiveRules(ctx, batch.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get active rules: %w", err)
	}

	if len(rules) == 0 {
		// No active rules for this project
		return nil
	}

	// Process each span event in the batch
	for _, event := range batch.Events {
		if event.EventType != string(observability.TelemetryEventTypeSpan) {
			continue
		}

		atomic.AddInt64(&w.spansProcessed, 1)

		// Match span against each rule
		for _, rule := range rules {
			if w.matchRule(rule, event) {
				atomic.AddInt64(&w.rulesMatched, 1)

				// Apply sampling rate
				if rule.SamplingRate < 1.0 && rand.Float64() > rule.SamplingRate {
					continue
				}

				// Extract variables from span
				variables := w.extractVariables(rule, event)

				// Create evaluation job
				job := &EvaluationJob{
					JobID:        ulid.New(),
					RuleID:       rule.ID,
					ProjectID:    batch.ProjectID,
					SpanData:     event.EventPayload,
					TraceID:      event.TraceID,
					SpanID:       event.SpanID,
					ScorerType:   rule.ScorerType,
					ScorerConfig: rule.ScorerConfig,
					Variables:    variables,
					CreatedAt:    time.Now(),
				}

				if err := w.emitJob(ctx, job); err != nil {
					w.logger.Error("Failed to emit evaluation job",
						"error", err,
						"job_id", job.JobID,
						"rule_id", rule.ID,
						"span_id", event.SpanID,
					)
					continue
				}

				atomic.AddInt64(&w.jobsEmitted, 1)
				w.logger.Debug("Emitted evaluation job",
					"job_id", job.JobID,
					"rule_id", rule.ID,
					"rule_name", rule.Name,
					"span_id", event.SpanID,
					"scorer_type", rule.ScorerType,
				)
			}
		}
	}

	return nil
}

func (w *RuleWorker) getActiveRules(ctx context.Context, projectID ulid.ULID) ([]*evaluation.EvaluationRule, error) {
	// Check cache first
	if rules, ok := w.ruleCache.Get(projectID.String()); ok {
		return rules, nil
	}

	// Fetch from service
	rules, err := w.ruleService.GetActiveByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Cache the results
	w.ruleCache.Set(projectID.String(), rules)
	return rules, nil
}

// matchRule checks if a span matches a rule's filters
func (w *RuleWorker) matchRule(rule *evaluation.EvaluationRule, event streams.TelemetryEventData) bool {
	// Check span name filter
	if len(rule.SpanNames) > 0 {
		spanName := safeExtractString(event.EventPayload, "span_name")
		matched := false
		for _, name := range rule.SpanNames {
			if name == spanName {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check filter clauses
	for _, clause := range rule.Filter {
		if !w.matchFilterClause(clause, event.EventPayload) {
			return false
		}
	}

	return true
}

func (w *RuleWorker) matchFilterClause(clause evaluation.FilterClause, payload map[string]interface{}) bool {
	value := extractFieldValue(payload, clause.Field)

	switch clause.Operator {
	case "equals", "eq":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", clause.Value)
	case "not_equals", "neq":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", clause.Value)
	case "contains":
		strValue := fmt.Sprintf("%v", value)
		strClause := fmt.Sprintf("%v", clause.Value)
		return strings.Contains(strValue, strClause)
	case "not_contains":
		strValue := fmt.Sprintf("%v", value)
		strClause := fmt.Sprintf("%v", clause.Value)
		return !strings.Contains(strValue, strClause)
	case "starts_with":
		strValue := fmt.Sprintf("%v", value)
		strClause := fmt.Sprintf("%v", clause.Value)
		return strings.HasPrefix(strValue, strClause)
	case "ends_with":
		strValue := fmt.Sprintf("%v", value)
		strClause := fmt.Sprintf("%v", clause.Value)
		return strings.HasSuffix(strValue, strClause)
	case "regex":
		strValue := fmt.Sprintf("%v", value)
		pattern := fmt.Sprintf("%v", clause.Value)
		matched, _ := regexp.MatchString(pattern, strValue)
		return matched
	case "is_empty":
		return value == nil || fmt.Sprintf("%v", value) == ""
	case "is_not_empty":
		return value != nil && fmt.Sprintf("%v", value) != ""
	case "gt":
		return compareNumeric(value, clause.Value) > 0
	case "gte":
		return compareNumeric(value, clause.Value) >= 0
	case "lt":
		return compareNumeric(value, clause.Value) < 0
	case "lte":
		return compareNumeric(value, clause.Value) <= 0
	default:
		w.logger.Warn("Unknown filter operator", "operator", clause.Operator)
		return true // Unknown operator - skip filter
	}
}

// extractFieldValue extracts a value from the payload using dot notation
// Supports: "input", "output", "span_name", "metadata.key", "span_attributes.key"
func extractFieldValue(payload map[string]interface{}, field string) interface{} {
	parts := strings.Split(field, ".")
	var current interface{} = payload

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[string]string:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

// extractVariables extracts variables from span based on rule's variable mapping
func (w *RuleWorker) extractVariables(rule *evaluation.EvaluationRule, event streams.TelemetryEventData) map[string]string {
	variables := make(map[string]string)

	for _, mapping := range rule.VariableMapping {
		var value interface{}

		switch mapping.Source {
		case "span_input":
			value = event.EventPayload["input"]
		case "span_output":
			value = event.EventPayload["output"]
		case "span_metadata":
			if metadata, ok := event.EventPayload["metadata"].(map[string]interface{}); ok {
				if mapping.JSONPath != "" {
					value = extractFieldValue(metadata, mapping.JSONPath)
				} else {
					value = metadata
				}
			}
		case "span_attributes":
			if attrs, ok := event.EventPayload["span_attributes"].(map[string]interface{}); ok {
				if mapping.JSONPath != "" {
					value = extractFieldValue(attrs, mapping.JSONPath)
				} else {
					value = attrs
				}
			}
		case "trace_input":
			// For trace-level input, use span input for now
			value = event.EventPayload["input"]
		default:
			// Direct field access
			if mapping.JSONPath != "" {
				value = extractFieldValue(event.EventPayload, mapping.JSONPath)
			} else {
				value = extractFieldValue(event.EventPayload, mapping.Source)
			}
		}

		if value != nil {
			switch v := value.(type) {
			case string:
				variables[mapping.VariableName] = v
			default:
				// Convert to JSON for complex types
				if jsonBytes, err := json.Marshal(v); err == nil {
					variables[mapping.VariableName] = string(jsonBytes)
				}
			}
		}
	}

	return variables
}

func (w *RuleWorker) emitJob(ctx context.Context, job *EvaluationJob) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = w.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: evaluationJobsStream,
		Values: map[string]interface{}{
			"job_id":     job.JobID.String(),
			"rule_id":    job.RuleID.String(),
			"project_id": job.ProjectID.String(),
			"span_id":    job.SpanID,
			"data":       string(jobData),
			"timestamp":  job.CreatedAt.Unix(),
		},
	}).Result()

	return err
}

// GetStats returns current worker statistics
func (w *RuleWorker) GetStats() map[string]int64 {
	w.streamsMutex.RLock()
	activeStreamCount := int64(len(w.activeStreams))
	w.streamsMutex.RUnlock()

	return map[string]int64{
		"spans_processed": atomic.LoadInt64(&w.spansProcessed),
		"rules_matched":   atomic.LoadInt64(&w.rulesMatched),
		"jobs_emitted":    atomic.LoadInt64(&w.jobsEmitted),
		"errors_count":    atomic.LoadInt64(&w.errorsCount),
		"active_streams":  activeStreamCount,
	}
}

// Utility functions

func safeExtractString(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	if val, ok := payload[key].(string); ok {
		return val
	}
	return ""
}

func compareNumeric(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)

	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
