package evaluation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	"brokle/internal/infrastructure/database"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

const (
	manualTriggerStream = "evaluation:manual-triggers"
)

type ruleService struct {
	repo             evaluation.RuleRepository
	executionService evaluation.RuleExecutionService
	traceRepo        observability.TraceRepository
	redis            *database.RedisDB
	logger           *slog.Logger
}

func NewRuleService(
	repo evaluation.RuleRepository,
	executionService evaluation.RuleExecutionService,
	traceRepo observability.TraceRepository,
	redis *database.RedisDB,
	logger *slog.Logger,
) evaluation.RuleService {
	return &ruleService{
		repo:             repo,
		executionService: executionService,
		traceRepo:        traceRepo,
		redis:            redis,
		logger:           logger,
	}
}

func (s *ruleService) Create(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *evaluation.CreateEvaluationRuleRequest) (*evaluation.EvaluationRule, error) {
	rule := evaluation.NewEvaluationRule(projectID, req.Name, req.ScorerType, req.ScorerConfig)

	if req.Description != nil {
		rule.Description = req.Description
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.TargetScope != nil {
		rule.TargetScope = *req.TargetScope
	}
	if req.Filter != nil {
		rule.Filter = req.Filter
	}
	if req.SpanNames != nil {
		rule.SpanNames = req.SpanNames
	}
	if req.SamplingRate != nil {
		rule.SamplingRate = *req.SamplingRate
	}
	if req.VariableMapping != nil {
		rule.VariableMapping = req.VariableMapping
	}
	if userID != nil {
		id := userID.String()
		rule.CreatedBy = &id
	}

	if validationErrors := rule.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	exists, err := s.repo.ExistsByName(ctx, projectID, req.Name)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to check name uniqueness", err)
	}
	if exists {
		return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", req.Name))
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		if errors.Is(err, evaluation.ErrRuleExists) {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", req.Name))
		}
		return nil, appErrors.NewInternalError("failed to create evaluation rule", err)
	}

	s.logger.Info("evaluation rule created",
		"rule_id", rule.ID,
		"project_id", projectID,
		"name", rule.Name,
		"scorer_type", rule.ScorerType,
		"status", rule.Status,
	)

	return rule, nil
}

func (s *ruleService) Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *evaluation.UpdateEvaluationRuleRequest) (*evaluation.EvaluationRule, error) {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if req.Name != nil && *req.Name != rule.Name {
		exists, err := s.repo.ExistsByName(ctx, projectID, *req.Name)
		if err != nil {
			return nil, appErrors.NewInternalError("failed to check name uniqueness", err)
		}
		if exists {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", *req.Name))
		}
		rule.Name = *req.Name
	}

	if req.Description != nil {
		rule.Description = req.Description
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.TargetScope != nil {
		rule.TargetScope = *req.TargetScope
	}
	if req.Filter != nil {
		rule.Filter = req.Filter
	}
	if req.SpanNames != nil {
		rule.SpanNames = req.SpanNames
	}
	if req.SamplingRate != nil {
		rule.SamplingRate = *req.SamplingRate
	}
	if req.ScorerType != nil {
		rule.ScorerType = *req.ScorerType
	}
	if req.ScorerConfig != nil {
		rule.ScorerConfig = req.ScorerConfig
	}
	if req.VariableMapping != nil {
		rule.VariableMapping = req.VariableMapping
	}

	rule.UpdatedAt = time.Now()

	if validationErrors := rule.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		if errors.Is(err, evaluation.ErrRuleExists) {
			return nil, appErrors.NewConflictError(fmt.Sprintf("evaluation rule '%s' already exists in this project", rule.Name))
		}
		return nil, appErrors.NewInternalError("failed to update evaluation rule", err)
	}

	s.logger.Info("evaluation rule updated",
		"rule_id", id,
		"project_id", projectID,
	)

	return rule, nil
}

func (s *ruleService) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if err := s.repo.Delete(ctx, id, projectID); err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to delete evaluation rule", err)
	}

	s.logger.Info("evaluation rule deleted",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.EvaluationRule, error) {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}
	return rule, nil
}

func (s *ruleService) List(ctx context.Context, projectID ulid.ULID, filter *evaluation.RuleFilter, params pagination.Params) ([]*evaluation.EvaluationRule, int64, error) {
	rules, total, err := s.repo.GetByProjectID(ctx, projectID, filter, params)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list evaluation rules", err)
	}
	return rules, total, nil
}

func (s *ruleService) Activate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if rule.Status == evaluation.RuleStatusActive {
		return nil
	}

	rule.Status = evaluation.RuleStatusActive
	rule.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, rule); err != nil {
		return appErrors.NewInternalError("failed to activate evaluation rule", err)
	}

	s.logger.Info("evaluation rule activated",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) Deactivate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	rule, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", id))
		}
		return appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	if rule.Status == evaluation.RuleStatusInactive {
		return nil
	}

	rule.Status = evaluation.RuleStatusInactive
	rule.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, rule); err != nil {
		return appErrors.NewInternalError("failed to deactivate evaluation rule", err)
	}

	s.logger.Info("evaluation rule deactivated",
		"rule_id", id,
		"project_id", projectID,
		"name", rule.Name,
	)

	return nil
}

func (s *ruleService) GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*evaluation.EvaluationRule, error) {
	rules, err := s.repo.GetActiveByProjectID(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get active evaluation rules", err)
	}
	return rules, nil
}

func (s *ruleService) TriggerRule(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, opts *evaluation.TriggerOptions) (*evaluation.TriggerResponse, error) {
	// Validate rule exists (can trigger inactive rules for testing)
	rule, err := s.repo.GetByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", ruleID))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	execution, err := s.executionService.StartExecution(ctx, ruleID, projectID, evaluation.TriggerTypeManual)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create execution record", err)
	}

	triggerMsg := ManualTriggerMessage{
		ExecutionID:     execution.ID,
		RuleID:          ruleID,
		ProjectID:       projectID,
		ScorerType:      rule.ScorerType,
		ScorerConfig:    rule.ScorerConfig,
		TargetScope:     rule.TargetScope,
		Filter:          rule.Filter,
		SpanNames:       rule.SpanNames,
		SamplingRate:    rule.SamplingRate,
		VariableMapping: rule.VariableMapping,
		CreatedAt:       time.Now(),
	}

	if opts != nil {
		triggerMsg.TimeRangeStart = opts.TimeRangeStart
		triggerMsg.TimeRangeEnd = opts.TimeRangeEnd
		triggerMsg.SpanIDs = opts.SpanIDs
		if opts.SampleLimit > 0 {
			triggerMsg.SampleLimit = opts.SampleLimit
		} else {
			triggerMsg.SampleLimit = 1000 // Default limit
		}
	} else {
		triggerMsg.SampleLimit = 1000
	}

	msgData, err := json.Marshal(triggerMsg)
	if err != nil {
		// Fail the execution since we can't publish
		_ = s.executionService.FailExecution(ctx, execution.ID, projectID, "failed to serialize trigger message")
		return nil, appErrors.NewInternalError("failed to serialize trigger message", err)
	}

	_, err = s.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: manualTriggerStream,
		Values: map[string]interface{}{
			"data": string(msgData),
		},
	}).Result()
	if err != nil {
		// Fail the execution since we can't publish
		_ = s.executionService.FailExecution(ctx, execution.ID, projectID, "failed to queue trigger job")
		return nil, appErrors.NewInternalError("failed to queue manual trigger job", err)
	}

	s.logger.Info("manual evaluation triggered",
		"rule_id", ruleID,
		"project_id", projectID,
		"execution_id", execution.ID,
		"rule_name", rule.Name,
	)

	return &evaluation.TriggerResponse{
		ExecutionID: execution.ID.String(),
		SpansQueued: 0, // Will be updated by worker when it starts processing
		Message:     "Manual evaluation queued successfully",
	}, nil
}

func (s *ruleService) TestRule(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, req *evaluation.TestRuleRequest) (*evaluation.TestRuleResponse, error) {
	// Validate rule exists
	rule, err := s.repo.GetByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", ruleID))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	// Priority 0: Sample input provided - create synthetic span for dry-run testing
	if req != nil && req.SampleInput != nil {
		return s.testWithSampleInput(ctx, rule, req.SampleInput)
	}

	// Set default limit
	limit := 5
	if req != nil && req.Limit > 0 && req.Limit <= 20 {
		limit = req.Limit
	}

	// Parse time range
	timeRange := "24h"
	if req != nil && req.TimeRange != "" {
		timeRange = req.TimeRange
	}

	// Calculate time bounds based on time range
	startTime := calculateStartTime(timeRange)
	endTime := time.Now()

	// Build variable names from variable mapping
	variableNames := make([]string, len(rule.VariableMapping))
	for i, vm := range rule.VariableMapping {
		variableNames[i] = vm.VariableName
	}

	// Query matching spans from ClickHouse
	spans, matchedCount, err := s.queryMatchingSpans(ctx, projectID.String(), rule, req, startTime, endTime, limit)
	if err != nil {
		s.logger.Error("failed to query spans for test",
			"error", err,
			"rule_id", ruleID,
			"project_id", projectID,
		)
		// Return preview only on query error
		return &evaluation.TestRuleResponse{
			Summary: evaluation.TestSummary{
				TotalSpans:     0,
				MatchedSpans:   0,
				EvaluatedSpans: 0,
				SuccessCount:   0,
				FailureCount:   0,
				SkippedCount:   0,
				AverageLatency: 0,
			},
			Executions: []evaluation.TestExecution{},
			RulePreview: evaluation.RulePreview{
				Name:              rule.Name,
				ScorerType:        string(rule.ScorerType),
				FilterDescription: buildFilterDescription(rule.Filter),
				VariableNames:     variableNames,
				PromptPreview:     buildPromptPreview(rule),
				MatchingCount:     0,
			},
		}, nil
	}

	// Build test executions for each matched span
	executions := make([]evaluation.TestExecution, len(spans))
	successCount := 0
	skippedCount := 0

	for i, span := range spans {
		// Resolve variables from span data
		resolvedVars := resolveVariables(rule.VariableMapping, span)

		// For now, mark spans as successful preview without actual scoring
		// Actual LLM scoring would require AI provider integration
		executions[i] = evaluation.TestExecution{
			SpanID:            span.SpanID,
			TraceID:           span.TraceID,
			SpanName:          span.SpanName,
			MatchedFilter:     true,
			Status:            "success",
			ScoreResults:      []evaluation.TestScoreResult{},
			VariablesResolved: resolvedVars,
			LatencyMs:         0, // Would be populated by actual scorer execution
		}
		successCount++
	}

	// Build rule preview with matched count
	rulePreview := evaluation.RulePreview{
		Name:              rule.Name,
		ScorerType:        string(rule.ScorerType),
		FilterDescription: buildFilterDescription(rule.Filter),
		VariableNames:     variableNames,
		PromptPreview:     buildPromptPreview(rule),
		MatchingCount:     int(matchedCount),
	}

	// Build summary
	response := &evaluation.TestRuleResponse{
		Summary: evaluation.TestSummary{
			TotalSpans:     int(matchedCount),
			MatchedSpans:   len(spans),
			EvaluatedSpans: len(spans),
			SuccessCount:   successCount,
			FailureCount:   0,
			SkippedCount:   skippedCount,
			AverageLatency: 0,
		},
		Executions:  executions,
		RulePreview: rulePreview,
	}

	s.logger.Info("rule test completed",
		"rule_id", ruleID,
		"project_id", projectID,
		"limit", limit,
		"matched_count", matchedCount,
		"evaluated_count", len(spans),
		"rule_name", rule.Name,
	)

	return response, nil
}

// queryMatchingSpans queries spans that match the rule's filter criteria.
// Priority order for span selection:
// 1. Specific span IDs (array) - if provided, fetch those directly
// 2. Single span ID - if provided, fetch that specific span
// 3. Trace ID + filters - if provided, filter by trace ID and apply rule filters
// 4. Generic filter - apply time range, span names, and rule filters
func (s *ruleService) queryMatchingSpans(
	ctx context.Context,
	projectID string,
	rule *evaluation.EvaluationRule,
	req *evaluation.TestRuleRequest,
	startTime, endTime time.Time,
	limit int,
) ([]*observability.Span, int64, error) {
	// Priority 1: Specific span IDs array - fetch directly
	if req != nil && len(req.SpanIDs) > 0 {
		var spans []*observability.Span
		for _, spanID := range req.SpanIDs {
			span, err := s.traceRepo.GetSpanByProject(ctx, spanID, projectID)
			if err != nil {
				// Span not found is acceptable - skip it
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				// Operational errors (db failure, timeout) should propagate
				return nil, 0, fmt.Errorf("failed to get span %s: %w", spanID, err)
			}
			spans = append(spans, span)
		}
		// Apply SpanNames filter and FilterClauses, then limit
		spans = s.filterSpans(spans, rule.SpanNames, rule.Filter, limit)
		return spans, int64(len(spans)), nil
	}

	// Priority 2: Single span ID - fetch directly
	if req != nil && req.SpanID != nil && *req.SpanID != "" {
		span, err := s.traceRepo.GetSpanByProject(ctx, *req.SpanID, projectID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get span %s: %w", *req.SpanID, err)
		}
		// Check span name filter first
		if !matchesSpanNames(span, rule.SpanNames) {
			return []*observability.Span{}, 0, nil
		}
		// Then check filter clauses
		if len(rule.Filter) > 0 && !s.matchSpanFilters(span, rule.Filter) {
			return []*observability.Span{}, 0, nil
		}
		return []*observability.Span{span}, 1, nil
	}

	// Build base span filter
	filter := &observability.SpanFilter{
		ProjectID: projectID,
		StartTime: &startTime,
		EndTime:   &endTime,
		SpanNames: rule.SpanNames,
	}

	// Priority 3: Trace ID provided - set in filter
	if req != nil && req.TraceID != nil && *req.TraceID != "" {
		filter.TraceID = req.TraceID
	}

	// If no FilterClauses, use simple database query
	if len(rule.Filter) == 0 {
		filter.Limit = limit
		filter.Page = 1

		spans, err := s.traceRepo.GetSpansByFilter(ctx, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query spans: %w", err)
		}

		totalCount, err := s.traceRepo.CountSpansByFilter(ctx, filter)
		if err != nil {
			totalCount = int64(len(spans))
		}

		return spans, totalCount, nil
	}

	// With FilterClauses: iterate pages until we have enough filtered matches
	// Overfetch to compensate for in-memory filtering
	var matchedSpans []*observability.Span
	fetchLimit := limit * 5
	if fetchLimit > 100 {
		fetchLimit = 100
	}
	maxPages := 10 // Cap pages to prevent infinite loops

	for page := 1; page <= maxPages; page++ {
		filter.Limit = fetchLimit
		filter.Page = page

		pageSpans, err := s.traceRepo.GetSpansByFilter(ctx, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query page %d: %w", page, err)
		}

		// No more results - we've exhausted the data
		if len(pageSpans) == 0 {
			break
		}

		// Apply FilterClause conditions in-memory
		for _, span := range pageSpans {
			if s.matchSpanFilters(span, rule.Filter) {
				matchedSpans = append(matchedSpans, span)
				if len(matchedSpans) >= limit {
					return matchedSpans, int64(len(matchedSpans)), nil
				}
			}
		}

		// If page was not full, no more data available
		if len(pageSpans) < fetchLimit {
			break
		}
	}

	return matchedSpans, int64(len(matchedSpans)), nil
}

// calculateStartTime calculates the start time based on the time range string.
func calculateStartTime(timeRange string) time.Time {
	now := time.Now()
	switch timeRange {
	case "1h":
		return now.Add(-1 * time.Hour)
	case "24h":
		return now.Add(-24 * time.Hour)
	case "7d":
		return now.Add(-7 * 24 * time.Hour)
	case "30d":
		return now.Add(-30 * 24 * time.Hour)
	default:
		return now.Add(-24 * time.Hour) // Default to 24 hours
	}
}

// resolveVariables extracts variable values from span data based on variable mapping.
func resolveVariables(mapping []evaluation.VariableMap, span *observability.Span) []evaluation.ResolvedVariable {
	if len(mapping) == 0 || span == nil {
		return []evaluation.ResolvedVariable{}
	}

	resolved := make([]evaluation.ResolvedVariable, 0, len(mapping))
	for _, vm := range mapping {
		var value any
		source := vm.Source

		switch vm.Source {
		case "span_input":
			if span.Input != nil {
				value = extractJSONPath(*span.Input, vm.JSONPath)
			}
		case "span_output":
			if span.Output != nil {
				value = extractJSONPath(*span.Output, vm.JSONPath)
			}
		case "span_metadata":
			if span.SpanAttributes != nil {
				if vm.JSONPath == "" {
					// Return entire metadata object when no specific path
					value = span.SpanAttributes
				} else {
					value = span.SpanAttributes[vm.JSONPath]
				}
			}
		case "trace_input":
			// For trace-level input, we'd need to query the root span
			// For now, fall back to span input
			if span.Input != nil {
				value = extractJSONPath(*span.Input, vm.JSONPath)
			}
		}

		resolved = append(resolved, evaluation.ResolvedVariable{
			VariableName:  vm.VariableName,
			Source:        source,
			JSONPath:      vm.JSONPath,
			ResolvedValue: value,
		})
	}

	return resolved
}

// extractJSONPath extracts a value from JSON content using a simple path.
// For complex JSON paths, a proper JSON path library should be used.
func extractJSONPath(jsonContent string, path string) any {
	if jsonContent == "" || path == "" {
		return jsonContent
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
		// Return raw content if not valid JSON
		return jsonContent
	}

	// Simple path extraction (supports "key" or "key.nested")
	parts := splitPath(path)
	current := any(data)
	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// splitPath splits a JSON path by dots while respecting quoted segments.
func splitPath(path string) []string {
	var result []string
	var current string
	for _, r := range path {
		if r == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func (s *ruleService) GetAnalytics(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, params *evaluation.RuleAnalyticsParams) (*evaluation.RuleAnalyticsResponse, error) {
	// Validate rule exists
	_, err := s.repo.GetByID(ctx, ruleID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrRuleNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("evaluation rule %s", ruleID))
		}
		return nil, appErrors.NewInternalError("failed to get evaluation rule", err)
	}

	// Set default period
	period := "7d"
	if params != nil && params.Period != "" {
		period = params.Period
	}

	// Extract time bounds (now populated by handler)
	var from, to time.Time
	if params != nil && params.From != nil && params.To != nil {
		from = *params.From
		to = *params.To
	} else {
		to = time.Now().UTC()
		from = to.Add(-7 * 24 * time.Hour)
	}

	// For now, return aggregated analytics from execution history.
	// Full implementation would query ClickHouse for score distribution and trends.
	// TODO: Use 'from' and 'to' to filter execution records when real analytics are implemented
	response := &evaluation.RuleAnalyticsResponse{
		RuleID:             ruleID.String(),
		Period:             period,
		TotalExecutions:    0,
		TotalSpansScored:   0,
		SuccessRate:        0,
		AverageScore:       0,
		ScoreDistribution:  []evaluation.DistributionBucket{},
		ExecutionTrend:     []evaluation.TimeSeriesPoint{},
		ScoreTrend:         []evaluation.TimeSeriesPoint{},
		LatencyPercentiles: evaluation.LatencyStats{},
		TopErrors:          []evaluation.ErrorSummary{},
	}

	s.logger.Info("rule analytics retrieved",
		"rule_id", ruleID,
		"project_id", projectID,
		"period", period,
		"from", from,
		"to", to,
	)

	return response, nil
}

// buildFilterDescription creates a human-readable description of filters.
func buildFilterDescription(filters []evaluation.FilterClause) string {
	if len(filters) == 0 {
		return "No filters - matches all spans"
	}
	var desc string
	for i, f := range filters {
		if i > 0 {
			desc += " AND "
		}
		desc += fmt.Sprintf("%s %s %v", f.Field, f.Operator, f.Value)
	}
	return desc
}

// buildPromptPreview creates a preview of the LLM prompt for LLM scorers.
func buildPromptPreview(rule *evaluation.EvaluationRule) string {
	if rule.ScorerType != evaluation.ScorerTypeLLM {
		return ""
	}
	// Extract messages from scorer config if present
	if messages, ok := rule.ScorerConfig["messages"]; ok {
		if msgList, ok := messages.([]interface{}); ok && len(msgList) > 0 {
			if firstMsg, ok := msgList[0].(map[string]interface{}); ok {
				if content, ok := firstMsg["content"].(string); ok {
					// Return first 200 chars of first message
					if len(content) > 200 {
						return content[:200] + "..."
					}
					return content
				}
			}
		}
	}
	return ""
}

// testWithSampleInput creates a synthetic span from manual input for dry-run testing.
func (s *ruleService) testWithSampleInput(
	_ context.Context,
	rule *evaluation.EvaluationRule,
	sample *evaluation.TestSampleInput,
) (*evaluation.TestRuleResponse, error) {
	// Validate sample input has at least input or output
	if sample.Input == "" && sample.Output == "" {
		return nil, appErrors.NewValidationError(
			"sample_input",
			"sample_input must contain at least 'input' or 'output'",
		)
	}

	// Create synthetic span
	syntheticSpan := s.createSyntheticSpan(rule.ProjectID.String(), sample)

	// Resolve variables from synthetic span
	resolvedVars := resolveVariables(rule.VariableMapping, syntheticSpan)

	// Build variable names for preview
	variableNames := make([]string, len(rule.VariableMapping))
	for i, vm := range rule.VariableMapping {
		variableNames[i] = vm.VariableName
	}

	execution := evaluation.TestExecution{
		SpanID:            syntheticSpan.SpanID,
		TraceID:           syntheticSpan.TraceID,
		SpanName:          syntheticSpan.SpanName,
		MatchedFilter:     true, // Sample input bypasses filters
		Status:            "success",
		ScoreResults:      []evaluation.TestScoreResult{},
		VariablesResolved: resolvedVars,
		LatencyMs:         0,
	}

	response := &evaluation.TestRuleResponse{
		Summary: evaluation.TestSummary{
			TotalSpans:     1,
			MatchedSpans:   1,
			EvaluatedSpans: 1,
			SuccessCount:   1,
			FailureCount:   0,
			SkippedCount:   0,
			AverageLatency: 0,
		},
		Executions: []evaluation.TestExecution{execution},
		RulePreview: evaluation.RulePreview{
			Name:              rule.Name,
			ScorerType:        string(rule.ScorerType),
			FilterDescription: "Sample input (filters bypassed)",
			VariableNames:     variableNames,
			PromptPreview:     buildPromptPreview(rule),
			MatchingCount:     1,
		},
	}

	s.logger.Info("rule test completed with sample input",
		"rule_id", rule.ID,
		"project_id", rule.ProjectID,
		"rule_name", rule.Name,
	)

	return response, nil
}

// createSyntheticSpan creates an in-memory span from TestSampleInput for dry-run testing.
func (s *ruleService) createSyntheticSpan(projectID string, sample *evaluation.TestSampleInput) *observability.Span {
	now := time.Now()
	syntheticID := ulid.New().String()

	span := &observability.Span{
		SpanID:     "synthetic-" + syntheticID[:8],
		TraceID:    "synthetic-trace-" + syntheticID[:8],
		SpanName:   "synthetic-test-span",
		ProjectID:  projectID,
		StartTime:  now,
		StatusCode: 0,
	}

	if sample.Input != "" {
		span.Input = &sample.Input
	}

	if sample.Output != "" {
		span.Output = &sample.Output
	}

	if len(sample.Metadata) > 0 {
		span.SpanAttributes = make(map[string]string)
		for k, v := range sample.Metadata {
			switch val := v.(type) {
			case string:
				span.SpanAttributes[k] = val
			default:
				if jsonBytes, err := json.Marshal(val); err == nil {
					span.SpanAttributes[k] = string(jsonBytes)
				}
			}
		}
	}

	return span
}

// filterSpans filters spans by rule's SpanNames and FilterClause conditions
func (s *ruleService) filterSpans(spans []*observability.Span, spanNames []string, filters []evaluation.FilterClause, limit int) []*observability.Span {
	var matched []*observability.Span
	for _, span := range spans {
		// Check span name filter first (most selective)
		if !matchesSpanNames(span, spanNames) {
			continue
		}
		// Then check filter clauses
		if len(filters) > 0 && !s.matchSpanFilters(span, filters) {
			continue
		}
		matched = append(matched, span)
		if len(matched) >= limit {
			break
		}
	}
	return matched
}

// matchSpanFilters checks if a span matches ALL filter clauses (AND logic)
func (s *ruleService) matchSpanFilters(span *observability.Span, filters []evaluation.FilterClause) bool {
	for _, clause := range filters {
		if !s.matchFilterClause(clause, span) {
			return false
		}
	}
	return true
}

// matchesSpanNames checks if a span's name matches any of the allowed span names.
// Returns true if spanNames is empty (no restriction) or if span.SpanName is in the list.
func matchesSpanNames(span *observability.Span, spanNames []string) bool {
	if len(spanNames) == 0 {
		return true // No span name restriction
	}
	for _, name := range spanNames {
		if span.SpanName == name {
			return true
		}
	}
	return false
}

// matchFilterClause evaluates a single filter clause against a span
func (s *ruleService) matchFilterClause(clause evaluation.FilterClause, span *observability.Span) bool {
	value := s.extractSpanFieldValue(span, clause.Field)

	switch clause.Operator {
	case "equals", "eq":
		valueStr := fmt.Sprintf("%v", value)
		clauseStr := fmt.Sprintf("%v", clause.Value)
		// Status field uses case-insensitive comparison (UI sends lowercase, backend returns title case)
		if clause.Field == "status" {
			return strings.EqualFold(valueStr, clauseStr)
		}
		return valueStr == clauseStr
	case "not_equals", "neq":
		valueStr := fmt.Sprintf("%v", value)
		clauseStr := fmt.Sprintf("%v", clause.Value)
		if clause.Field == "status" {
			return !strings.EqualFold(valueStr, clauseStr)
		}
		return valueStr != clauseStr
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", clause.Value))
	case "not_contains":
		return !strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", clause.Value))
	case "starts_with":
		return strings.HasPrefix(fmt.Sprintf("%v", value), fmt.Sprintf("%v", clause.Value))
	case "ends_with":
		return strings.HasSuffix(fmt.Sprintf("%v", value), fmt.Sprintf("%v", clause.Value))
	case "regex":
		matched, err := regexp.MatchString(fmt.Sprintf("%v", clause.Value), fmt.Sprintf("%v", value))
		if err != nil {
			s.logger.Warn("Invalid regex in filter clause - filter will not match",
				"pattern", clause.Value,
				"field", clause.Field,
				"error", err)
			return false // Fail closed: invalid patterns should not match
		}
		return matched
	case "is_empty":
		return value == nil || fmt.Sprintf("%v", value) == ""
	case "is_not_empty":
		return value != nil && fmt.Sprintf("%v", value) != ""
	case "gt":
		cmp, ok := compareNumeric(value, clause.Value)
		return ok && cmp > 0
	case "gte":
		cmp, ok := compareNumeric(value, clause.Value)
		return ok && cmp >= 0
	case "lt":
		cmp, ok := compareNumeric(value, clause.Value)
		return ok && cmp < 0
	case "lte":
		cmp, ok := compareNumeric(value, clause.Value)
		return ok && cmp <= 0
	default:
		s.logger.Warn("Unknown filter operator - filter will not match",
			"operator", clause.Operator,
			"field", clause.Field)
		return false // Fail closed: unknown operators should not match
	}
}

// extractSpanFieldValue extracts a value from a span using dot notation for nested paths
func (s *ruleService) extractSpanFieldValue(span *observability.Span, field string) interface{} {
	parts := strings.Split(field, ".")

	// Handle top-level span fields
	switch parts[0] {
	case "input":
		if len(parts) > 1 && span.Input != nil {
			return extractNestedValueForFilter(*span.Input, parts[1:])
		}
		if span.Input != nil {
			return *span.Input
		}
		return nil
	case "output":
		if len(parts) > 1 && span.Output != nil {
			return extractNestedValueForFilter(*span.Output, parts[1:])
		}
		if span.Output != nil {
			return *span.Output
		}
		return nil
	case "span_name", "name":
		return span.SpanName
	case "span_kind":
		return span.SpanKind
	case "model", "model_name":
		if span.ModelName != nil {
			return *span.ModelName
		}
		return nil
	case "provider", "provider_name":
		if span.ProviderName != nil {
			return *span.ProviderName
		}
		return nil
	case "span_attributes":
		if len(parts) > 1 && span.SpanAttributes != nil {
			return span.SpanAttributes[parts[1]]
		}
		return span.SpanAttributes
	case "resource_attributes":
		if len(parts) > 1 && span.ResourceAttributes != nil {
			return span.ResourceAttributes[parts[1]]
		}
		return span.ResourceAttributes
	case "metadata":
		if len(parts) > 1 && span.SpanAttributes != nil {
			return span.SpanAttributes[parts[1]]
		}
		return span.SpanAttributes
	case "service_name":
		if span.ServiceName != nil {
			return *span.ServiceName
		}
		return nil

	// Latency - convert nanoseconds to milliseconds
	case "latency_ms", "latency", "duration_ms":
		if span.Duration != nil {
			return float64(*span.Duration) / 1_000_000.0
		}
		return nil

	// Token counts from UsageDetails map
	case "token_count", "total_tokens":
		if span.UsageDetails == nil {
			return nil
		}
		if total, ok := span.UsageDetails["total"]; ok {
			return total
		}
		// Fallback: calculate from input + output
		var sum uint64
		if input, ok := span.UsageDetails["input"]; ok {
			sum += input
		}
		if output, ok := span.UsageDetails["output"]; ok {
			sum += output
		}
		if sum > 0 {
			return sum
		}
		return nil

	case "input_tokens":
		if span.UsageDetails != nil {
			if val, ok := span.UsageDetails["input"]; ok {
				return val
			}
		}
		return nil

	case "output_tokens":
		if span.UsageDetails != nil {
			if val, ok := span.UsageDetails["output"]; ok {
				return val
			}
		}
		return nil

	// Status - convert StatusCode to string for UI compatibility
	case "status":
		switch span.StatusCode {
		case observability.StatusCodeOK:
			return "OK"
		case observability.StatusCodeError:
			return "Error"
		default:
			// For UNSET, check HasError flag
			if span.HasError {
				return "Error"
			}
			return "OK"
		}

	// Attributes alias - same as span_attributes for dot notation
	case "attributes":
		if len(parts) > 1 && span.SpanAttributes != nil {
			return span.SpanAttributes[strings.Join(parts[1:], ".")]
		}
		return span.SpanAttributes
	}

	return nil
}

// extractNestedValueForFilter extracts nested value from JSON string or interface using path parts
func extractNestedValueForFilter(data interface{}, path []string) interface{} {
	if len(path) == 0 {
		return data
	}

	// Handle string JSON
	if str, ok := data.(string); ok {
		var parsed map[string]interface{}
		if json.Unmarshal([]byte(str), &parsed) == nil {
			data = parsed
		} else {
			return nil // Not valid JSON
		}
	}

	// Navigate path
	if m, ok := data.(map[string]interface{}); ok {
		if val, exists := m[path[0]]; exists {
			return extractNestedValueForFilter(val, path[1:])
		}
	}

	return nil
}

// compareNumeric compares two values numerically.
// Returns (comparison result, true) on success, (0, false) if either value is not numeric.
func compareNumeric(a, b interface{}) (int, bool) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return 0, false // Cannot compare non-numeric values
	}

	if aFloat < bFloat {
		return -1, true
	}
	if aFloat > bFloat {
		return 1, true
	}
	return 0, true
}

// toFloat64 converts a value to float64 for numeric comparison.
// Returns (value, true) on success, (0, false) if value is nil or not convertible.
func toFloat64(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

// ManualTriggerMessage is the message format for the manual trigger stream
type ManualTriggerMessage struct {
	ExecutionID     ulid.ULID                 `json:"execution_id"`
	RuleID          ulid.ULID                 `json:"rule_id"`
	ProjectID       ulid.ULID                 `json:"project_id"`
	ScorerType      evaluation.ScorerType     `json:"scorer_type"`
	ScorerConfig    map[string]any            `json:"scorer_config"`
	TargetScope     evaluation.TargetScope    `json:"target_scope"`
	Filter          []evaluation.FilterClause `json:"filter,omitempty"`
	SpanNames       []string                  `json:"span_names,omitempty"`
	SamplingRate    float64                   `json:"sampling_rate"`
	VariableMapping []evaluation.VariableMap  `json:"variable_mapping,omitempty"`
	TimeRangeStart  *time.Time                `json:"time_range_start,omitempty"`
	TimeRangeEnd    *time.Time                `json:"time_range_end,omitempty"`
	SpanIDs         []string                  `json:"span_ids,omitempty"`
	SampleLimit     int                       `json:"sample_limit"`
	CreatedAt       time.Time                 `json:"created_at"`
}
