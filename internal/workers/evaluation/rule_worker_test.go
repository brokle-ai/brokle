package evaluation

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/streams"
	"brokle/pkg/ulid"
)

// Test helper for creating a test logger
func newTestRuleWorkerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// =============================================================================
// RuleCache Tests
// =============================================================================

func TestRuleCache_NewRuleCache(t *testing.T) {
	ttl := 30 * time.Second
	cache := NewRuleCache(ttl)

	require.NotNil(t, cache)
	assert.NotNil(t, cache.cache)
	assert.Equal(t, ttl, cache.ttl)
}

func TestRuleCache_GetFromEmptyCache(t *testing.T) {
	cache := NewRuleCache(30 * time.Second)

	rules, found := cache.Get("project123")
	assert.False(t, found)
	assert.Nil(t, rules)
}

func TestRuleCache_SetAndGet(t *testing.T) {
	cache := NewRuleCache(30 * time.Second)

	projectID := "project123"
	rules := []*evaluation.EvaluationRule{
		{
			ID:   ulid.New(),
			Name: "test-rule",
		},
	}

	cache.Set(projectID, rules)
	retrieved, found := cache.Get(projectID)

	assert.True(t, found)
	require.Len(t, retrieved, 1)
	assert.Equal(t, rules[0].ID, retrieved[0].ID)
	assert.Equal(t, rules[0].Name, retrieved[0].Name)
}

func TestRuleCache_Expiration(t *testing.T) {
	// Use a very short TTL for testing
	cache := NewRuleCache(10 * time.Millisecond)

	projectID := "project123"
	rules := []*evaluation.EvaluationRule{
		{ID: ulid.New(), Name: "test-rule"},
	}

	cache.Set(projectID, rules)

	// Should be found immediately
	_, found := cache.Get(projectID)
	assert.True(t, found)

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.Get(projectID)
	assert.False(t, found)
}

func TestRuleCache_Invalidate(t *testing.T) {
	cache := NewRuleCache(30 * time.Second)

	projectID := "project123"
	rules := []*evaluation.EvaluationRule{
		{ID: ulid.New(), Name: "test-rule"},
	}

	cache.Set(projectID, rules)

	// Should be found before invalidation
	_, found := cache.Get(projectID)
	assert.True(t, found)

	cache.Invalidate(projectID)

	// Should not be found after invalidation
	_, found = cache.Get(projectID)
	assert.False(t, found)
}

func TestRuleCache_MultipleProjects(t *testing.T) {
	cache := NewRuleCache(30 * time.Second)

	rules1 := []*evaluation.EvaluationRule{{ID: ulid.New(), Name: "rule1"}}
	rules2 := []*evaluation.EvaluationRule{{ID: ulid.New(), Name: "rule2"}}

	cache.Set("project1", rules1)
	cache.Set("project2", rules2)

	retrieved1, found1 := cache.Get("project1")
	retrieved2, found2 := cache.Get("project2")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.Equal(t, "rule1", retrieved1[0].Name)
	assert.Equal(t, "rule2", retrieved2[0].Name)

	// Invalidate one project
	cache.Invalidate("project1")

	_, found1 = cache.Get("project1")
	_, found2 = cache.Get("project2")
	assert.False(t, found1)
	assert.True(t, found2)
}

// =============================================================================
// extractFieldValue Tests
// =============================================================================

func TestExtractFieldValue_SimpleField(t *testing.T) {
	payload := map[string]interface{}{
		"input":  "hello world",
		"output": "response text",
	}

	assert.Equal(t, "hello world", extractFieldValue(payload, "input"))
	assert.Equal(t, "response text", extractFieldValue(payload, "output"))
}

func TestExtractFieldValue_NestedDotNotation(t *testing.T) {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "user123",
				"name": "John",
			},
			"version": "1.0",
		},
	}

	assert.Equal(t, "user123", extractFieldValue(payload, "metadata.user.id"))
	assert.Equal(t, "John", extractFieldValue(payload, "metadata.user.name"))
	assert.Equal(t, "1.0", extractFieldValue(payload, "metadata.version"))
}

func TestExtractFieldValue_MissingField(t *testing.T) {
	payload := map[string]interface{}{
		"input": "hello",
	}

	assert.Nil(t, extractFieldValue(payload, "nonexistent"))
	assert.Nil(t, extractFieldValue(payload, "metadata.key"))
	assert.Nil(t, extractFieldValue(payload, "a.b.c.d"))
}

func TestExtractFieldValue_MapStringString(t *testing.T) {
	payload := map[string]interface{}{
		"headers": map[string]string{
			"content-type": "application/json",
			"authorization": "Bearer token",
		},
	}

	// First level returns the nested map
	headers := extractFieldValue(payload, "headers")
	require.NotNil(t, headers)

	// When the nested value is map[string]string, it should handle dot notation
	assert.Equal(t, "application/json", extractFieldValue(payload, "headers.content-type"))
	assert.Equal(t, "Bearer token", extractFieldValue(payload, "headers.authorization"))
}

func TestExtractFieldValue_EmptyPayload(t *testing.T) {
	assert.Nil(t, extractFieldValue(nil, "field"))
	assert.Nil(t, extractFieldValue(map[string]interface{}{}, "field"))
}

func TestExtractFieldValue_NumericValues(t *testing.T) {
	payload := map[string]interface{}{
		"count":    42,
		"price":    19.99,
		"metadata": map[string]interface{}{"score": 0.95},
	}

	assert.Equal(t, 42, extractFieldValue(payload, "count"))
	assert.Equal(t, 19.99, extractFieldValue(payload, "price"))
	assert.Equal(t, 0.95, extractFieldValue(payload, "metadata.score"))
}

// =============================================================================
// matchFilterClause Tests
// =============================================================================

func TestMatchFilterClause_Equals(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"status": "active", "count": 10}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "equals string match",
			clause:   evaluation.FilterClause{Field: "status", Operator: "equals", Value: "active"},
			expected: true,
		},
		{
			name:     "equals string no match",
			clause:   evaluation.FilterClause{Field: "status", Operator: "equals", Value: "inactive"},
			expected: false,
		},
		{
			name:     "eq alias",
			clause:   evaluation.FilterClause{Field: "status", Operator: "eq", Value: "active"},
			expected: true,
		},
		{
			name:     "equals number match",
			clause:   evaluation.FilterClause{Field: "count", Operator: "equals", Value: 10},
			expected: true,
		},
		{
			name:     "equals number no match",
			clause:   evaluation.FilterClause{Field: "count", Operator: "equals", Value: 20},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_NotEquals(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"status": "active"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "not_equals when different",
			clause:   evaluation.FilterClause{Field: "status", Operator: "not_equals", Value: "inactive"},
			expected: true,
		},
		{
			name:     "not_equals when same",
			clause:   evaluation.FilterClause{Field: "status", Operator: "not_equals", Value: "active"},
			expected: false,
		},
		{
			name:     "neq alias",
			clause:   evaluation.FilterClause{Field: "status", Operator: "neq", Value: "inactive"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_Contains(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"message": "hello world today"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "contains match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "contains", Value: "world"},
			expected: true,
		},
		{
			name:     "contains no match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "contains", Value: "goodbye"},
			expected: false,
		},
		{
			name:     "contains partial",
			clause:   evaluation.FilterClause{Field: "message", Operator: "contains", Value: "ello wor"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_NotContains(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"message": "hello world"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "not_contains when absent",
			clause:   evaluation.FilterClause{Field: "message", Operator: "not_contains", Value: "goodbye"},
			expected: true,
		},
		{
			name:     "not_contains when present",
			clause:   evaluation.FilterClause{Field: "message", Operator: "not_contains", Value: "world"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_StartsWith(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"message": "hello world"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "starts_with match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "starts_with", Value: "hello"},
			expected: true,
		},
		{
			name:     "starts_with no match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "starts_with", Value: "world"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_EndsWith(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"message": "hello world"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "ends_with match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "ends_with", Value: "world"},
			expected: true,
		},
		{
			name:     "ends_with no match",
			clause:   evaluation.FilterClause{Field: "message", Operator: "ends_with", Value: "hello"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_Regex(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"email": "test@example.com"}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "regex email match",
			clause:   evaluation.FilterClause{Field: "email", Operator: "regex", Value: `[a-z]+@[a-z]+\.[a-z]+`},
			expected: true,
		},
		{
			name:     "regex no match",
			clause:   evaluation.FilterClause{Field: "email", Operator: "regex", Value: `^[0-9]+$`},
			expected: false,
		},
		{
			name:     "regex digit pattern",
			clause:   evaluation.FilterClause{Field: "email", Operator: "regex", Value: `\d`},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_IsEmpty(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	tests := []struct {
		name     string
		payload  map[string]interface{}
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "is_empty when nil",
			payload:  map[string]interface{}{},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_empty"},
			expected: true,
		},
		{
			name:     "is_empty when empty string",
			payload:  map[string]interface{}{"field": ""},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_empty"},
			expected: true,
		},
		{
			name:     "is_empty when has value",
			payload:  map[string]interface{}{"field": "value"},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_empty"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, tt.payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_IsNotEmpty(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	tests := []struct {
		name     string
		payload  map[string]interface{}
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "is_not_empty when has value",
			payload:  map[string]interface{}{"field": "value"},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_not_empty"},
			expected: true,
		},
		{
			name:     "is_not_empty when nil",
			payload:  map[string]interface{}{},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_not_empty"},
			expected: false,
		},
		{
			name:     "is_not_empty when empty string",
			payload:  map[string]interface{}{"field": ""},
			clause:   evaluation.FilterClause{Field: "field", Operator: "is_not_empty"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, tt.payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_NumericComparisons(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"score": 75.5}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "gt true",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gt", Value: 70.0},
			expected: true,
		},
		{
			name:     "gt false",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gt", Value: 80.0},
			expected: false,
		},
		{
			name:     "gt equal value",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gt", Value: 75.5},
			expected: false,
		},
		{
			name:     "gte true when greater",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gte", Value: 70.0},
			expected: true,
		},
		{
			name:     "gte true when equal",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gte", Value: 75.5},
			expected: true,
		},
		{
			name:     "gte false",
			clause:   evaluation.FilterClause{Field: "score", Operator: "gte", Value: 80.0},
			expected: false,
		},
		{
			name:     "lt true",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lt", Value: 80.0},
			expected: true,
		},
		{
			name:     "lt false",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lt", Value: 70.0},
			expected: false,
		},
		{
			name:     "lt equal value",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lt", Value: 75.5},
			expected: false,
		},
		{
			name:     "lte true when less",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lte", Value: 80.0},
			expected: true,
		},
		{
			name:     "lte true when equal",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lte", Value: 75.5},
			expected: true,
		},
		{
			name:     "lte false",
			clause:   evaluation.FilterClause{Field: "score", Operator: "lte", Value: 70.0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchFilterClause_UnknownOperator(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{"field": "value"}

	clause := evaluation.FilterClause{
		Field:    "field",
		Operator: "unknown_operator",
		Value:    "something",
	}

	// Unknown operators should return true (skip filter)
	result := worker.matchFilterClause(clause, payload)
	assert.True(t, result)
}

func TestMatchFilterClause_NestedField(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"user": map[string]interface{}{
				"role": "admin",
			},
		},
	}

	tests := []struct {
		name     string
		clause   evaluation.FilterClause
		expected bool
	}{
		{
			name:     "nested field equals",
			clause:   evaluation.FilterClause{Field: "metadata.user.role", Operator: "equals", Value: "admin"},
			expected: true,
		},
		{
			name:     "nested field not equals",
			clause:   evaluation.FilterClause{Field: "metadata.user.role", Operator: "equals", Value: "user"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchFilterClause(tt.clause, payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// matchRule Tests
// =============================================================================

func TestMatchRule_BySpanName(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		SpanNames: []string{"chat_completion", "embeddings"},
	}

	tests := []struct {
		name     string
		event    streams.TelemetryEventData
		expected bool
	}{
		{
			name: "matches first span name",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "chat_completion"},
			},
			expected: true,
		},
		{
			name: "matches second span name",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "embeddings"},
			},
			expected: true,
		},
		{
			name: "no match",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "other_span"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchRule(rule, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchRule_ByFilters(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		SpanNames: []string{},
		Filter: []evaluation.FilterClause{
			{Field: "status", Operator: "equals", Value: "success"},
		},
	}

	tests := []struct {
		name     string
		event    streams.TelemetryEventData
		expected bool
	}{
		{
			name: "filter matches",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "success"},
			},
			expected: true,
		},
		{
			name: "filter does not match",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "error"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchRule(rule, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchRule_MultipleFilters(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		SpanNames: []string{},
		Filter: []evaluation.FilterClause{
			{Field: "status", Operator: "equals", Value: "success"},
			{Field: "score", Operator: "gte", Value: 80.0},
		},
	}

	tests := []struct {
		name     string
		event    streams.TelemetryEventData
		expected bool
	}{
		{
			name: "all filters match (AND logic)",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "success", "score": 90.0},
			},
			expected: true,
		},
		{
			name: "first filter fails",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "error", "score": 90.0},
			},
			expected: false,
		},
		{
			name: "second filter fails",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "success", "score": 70.0},
			},
			expected: false,
		},
		{
			name: "both filters fail",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"status": "error", "score": 70.0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchRule(rule, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchRule_EmptyFilters(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		SpanNames: []string{},
		Filter:    []evaluation.FilterClause{},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{"anything": "any value"},
	}

	// Empty filters should always match
	result := worker.matchRule(rule, event)
	assert.True(t, result)
}

func TestMatchRule_SpanNameAndFilters(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		SpanNames: []string{"chat_completion"},
		Filter: []evaluation.FilterClause{
			{Field: "status", Operator: "equals", Value: "success"},
		},
	}

	tests := []struct {
		name     string
		event    streams.TelemetryEventData
		expected bool
	}{
		{
			name: "span name and filter both match",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "chat_completion", "status": "success"},
			},
			expected: true,
		},
		{
			name: "span name matches, filter fails",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "chat_completion", "status": "error"},
			},
			expected: false,
		},
		{
			name: "span name fails, filter matches",
			event: streams.TelemetryEventData{
				EventPayload: map[string]interface{}{"span_name": "other", "status": "success"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.matchRule(rule, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// extractVariables Tests
// =============================================================================

func TestExtractVariables_SpanInput(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "input", Source: "span_input"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"input": "user question here",
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "user question here", variables["input"])
}

func TestExtractVariables_SpanOutput(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "output", Source: "span_output"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"output": "AI response here",
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "AI response here", variables["output"])
}

func TestExtractVariables_SpanMetadataWithJSONPath(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "user_id", Source: "span_metadata", JSONPath: "user.id"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"metadata": map[string]interface{}{
				"user": map[string]interface{}{
					"id": "user123",
				},
			},
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "user123", variables["user_id"])
}

func TestExtractVariables_SpanAttributes(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "model", Source: "span_attributes", JSONPath: "llm.model"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"span_attributes": map[string]interface{}{
				"llm": map[string]interface{}{
					"model": "gpt-4",
				},
			},
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "gpt-4", variables["model"])
}

func TestExtractVariables_ComplexTypeToJSON(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "config", Source: "span_metadata"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"metadata": map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
	}

	variables := worker.extractVariables(rule, event)

	// Complex types should be serialized to JSON
	assert.Contains(t, variables["config"], "key1")
	assert.Contains(t, variables["config"], "value1")
}

func TestExtractVariables_MissingSource(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "missing", Source: "span_input"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{},
	}

	variables := worker.extractVariables(rule, event)

	// Missing sources should result in no variable
	_, exists := variables["missing"]
	assert.False(t, exists)
}

func TestExtractVariables_MultipleVariables(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "input", Source: "span_input"},
			{VariableName: "output", Source: "span_output"},
			{VariableName: "model", Source: "span_attributes", JSONPath: "model"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"input":  "question",
			"output": "answer",
			"span_attributes": map[string]interface{}{
				"model": "gpt-4",
			},
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "question", variables["input"])
	assert.Equal(t, "answer", variables["output"])
	assert.Equal(t, "gpt-4", variables["model"])
}

func TestExtractVariables_DirectFieldAccess(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "custom", Source: "custom_field"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"custom_field": "custom value",
		},
	}

	variables := worker.extractVariables(rule, event)

	assert.Equal(t, "custom value", variables["custom"])
}

func TestExtractVariables_TraceInput(t *testing.T) {
	worker := &RuleWorker{logger: newTestRuleWorkerLogger()}

	rule := &evaluation.EvaluationRule{
		VariableMapping: []evaluation.VariableMap{
			{VariableName: "trace_in", Source: "trace_input"},
		},
	}

	event := streams.TelemetryEventData{
		EventPayload: map[string]interface{}{
			"input": "trace input value",
		},
	}

	variables := worker.extractVariables(rule, event)

	// trace_input falls back to span input
	assert.Equal(t, "trace input value", variables["trace_in"])
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestSafeExtractString(t *testing.T) {
	tests := []struct {
		name     string
		payload  map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "existing string key",
			payload:  map[string]interface{}{"name": "test"},
			key:      "name",
			expected: "test",
		},
		{
			name:     "missing key",
			payload:  map[string]interface{}{"name": "test"},
			key:      "other",
			expected: "",
		},
		{
			name:     "nil payload",
			payload:  nil,
			key:      "name",
			expected: "",
		},
		{
			name:     "non-string value",
			payload:  map[string]interface{}{"count": 42},
			key:      "count",
			expected: "",
		},
		{
			name:     "empty payload",
			payload:  map[string]interface{}{},
			key:      "name",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeExtractString(tt.payload, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{
			name:     "equal floats",
			a:        10.0,
			b:        10.0,
			expected: 0,
		},
		{
			name:     "a greater than b",
			a:        15.0,
			b:        10.0,
			expected: 1,
		},
		{
			name:     "a less than b",
			a:        5.0,
			b:        10.0,
			expected: -1,
		},
		{
			name:     "int vs float",
			a:        10,
			b:        10.0,
			expected: 0,
		},
		{
			name:     "string number",
			a:        "10.5",
			b:        10.0,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNumeric(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{
			name:     "float64",
			input:    42.5,
			expected: 42.5,
		},
		{
			name:     "float32",
			input:    float32(42.5),
			expected: 42.5,
		},
		{
			name:     "int",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "int64",
			input:    int64(42),
			expected: 42.0,
		},
		{
			name:     "int32",
			input:    int32(42),
			expected: 42.0,
		},
		{
			name:     "string number",
			input:    "42.5",
			expected: 42.5,
		},
		{
			name:     "invalid string",
			input:    "not a number",
			expected: 0,
		},
		{
			name:     "nil",
			input:    nil,
			expected: 0,
		},
		{
			name:     "unknown type",
			input:    struct{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFloat64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinDuration(t *testing.T) {
	tests := []struct {
		name     string
		a        time.Duration
		b        time.Duration
		expected time.Duration
	}{
		{
			name:     "a is smaller",
			a:        5 * time.Second,
			b:        10 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "b is smaller",
			a:        10 * time.Second,
			b:        5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "equal",
			a:        5 * time.Second,
			b:        5 * time.Second,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minDuration(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
