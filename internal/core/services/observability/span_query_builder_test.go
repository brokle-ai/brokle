package observability

import (
	"strings"
	"testing"
	"time"

	obsDomain "brokle/internal/core/domain/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanQueryBuilder_BuildQuery(t *testing.T) {
	tests := []struct {
		name           string
		node           obsDomain.FilterNode
		projectID      string
		startTime      *time.Time
		endTime        *time.Time
		limit          int
		offset         int
		wantContains   []string // SQL fragments that should be present
		wantArgCount   int      // Expected number of arguments
		wantFirstArgs  []interface{}
		wantErr        bool
	}{
		{
			name: "simple equality - materialized column",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpEqual,
				Value:    "chatbot",
			},
			projectID:     "proj123",
			limit:         100,
			offset:        0,
			wantContains:  []string{"service_name = ?", "project_id = ?", "deleted_at IS NULL"},
			wantArgCount:  4, // projectID, filter value, limit, offset
			wantFirstArgs: []interface{}{"proj123"},
			wantErr:       false,
		},
		{
			name: "simple equality - span attribute",
			node: &obsDomain.ConditionNode{
				Field:    "custom.field",
				Operator: obsDomain.FilterOpEqual,
				Value:    "value",
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"span_attributes['custom.field'] = ?", "project_id = ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "resource attribute",
			node: &obsDomain.ConditionNode{
				Field:    "resource.deployment.env",
				Operator: obsDomain.FilterOpEqual,
				Value:    "production",
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"resource_attributes['resource.deployment.env'] = ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "not equal operator",
			node: &obsDomain.ConditionNode{
				Field:    "gen_ai.system",
				Operator: obsDomain.FilterOpNotEqual,
				Value:    "anthropic",
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"provider_name != ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "greater than - materialized column",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpGreaterThan,
				Value:    float64(1000),
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"service_name > ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "greater than - span attribute with type coercion",
			node: &obsDomain.ConditionNode{
				Field:    "gen_ai.usage.total_tokens",
				Operator: obsDomain.FilterOpGreaterThan,
				Value:    float64(1000),
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"toFloat64OrNull(span_attributes['gen_ai.usage.total_tokens']) > ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "less than or equal",
			node: &obsDomain.ConditionNode{
				Field:    "custom.latency",
				Operator: obsDomain.FilterOpLessOrEqual,
				Value:    float64(500),
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"toFloat64OrNull(span_attributes['custom.latency']) <= ?"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "CONTAINS operator",
			node: &obsDomain.ConditionNode{
				Field:    "span.name",
				Operator: obsDomain.FilterOpContains,
				Value:    "llm",
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"positionCaseInsensitive(span_name, ?) > 0"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "NOT CONTAINS operator",
			node: &obsDomain.ConditionNode{
				Field:    "span.name",
				Operator: obsDomain.FilterOpNotContains,
				Value:    "test",
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"positionCaseInsensitive(span_name, ?) = 0"},
			wantArgCount: 4,
			wantErr:      false,
		},
		{
			name: "IN clause",
			node: &obsDomain.ConditionNode{
				Field:    "gen_ai.request.model",
				Operator: obsDomain.FilterOpIn,
				Value:    []string{"gpt-4o", "gpt-4", "gpt-3.5-turbo"},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"model_name IN (?, ?, ?)"},
			wantArgCount: 6, // projectID, 3 IN values, limit, offset
			wantErr:      false,
		},
		{
			name: "NOT IN clause",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpNotIn,
				Value:    []string{"test", "dev"},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"service_name NOT IN (?, ?)"},
			wantArgCount: 5, // projectID, 2 NOT IN values, limit, offset
			wantErr:      false,
		},
		{
			name: "empty IN clause",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpIn,
				Value:    []string{},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"1=0"}, // Always false for empty IN
			wantArgCount: 3,               // projectID, limit, offset (no filter values)
			wantErr:      false,
		},
		{
			name: "EXISTS operator - span attribute",
			node: &obsDomain.ConditionNode{
				Field:    "custom.field",
				Operator: obsDomain.FilterOpExists,
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"mapContains(span_attributes, 'custom.field')"},
			wantArgCount: 3, // No filter args for EXISTS
			wantErr:      false,
		},
		{
			name: "NOT EXISTS operator - resource attribute",
			node: &obsDomain.ConditionNode{
				Field:    "resource.custom",
				Operator: obsDomain.FilterOpNotExists,
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"NOT mapContains(resource_attributes, 'resource.custom')"},
			wantArgCount: 3,
			wantErr:      false,
		},
		{
			name: "EXISTS operator - materialized column",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpExists,
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"(service_name IS NOT NULL AND service_name != '')"},
			wantArgCount: 3,
			wantErr:      false,
		},
		{
			name: "AND expression",
			node: &obsDomain.BinaryNode{
				Left: &obsDomain.ConditionNode{
					Field:    "service.name",
					Operator: obsDomain.FilterOpEqual,
					Value:    "chatbot",
				},
				Operator: obsDomain.LogicAnd,
				Right: &obsDomain.ConditionNode{
					Field:    "gen_ai.system",
					Operator: obsDomain.FilterOpEqual,
					Value:    "openai",
				},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"(service_name = ? AND provider_name = ?)"},
			wantArgCount: 5, // projectID, 2 filter values, limit, offset
			wantErr:      false,
		},
		{
			name: "OR expression",
			node: &obsDomain.BinaryNode{
				Left: &obsDomain.ConditionNode{
					Field:    "gen_ai.system",
					Operator: obsDomain.FilterOpEqual,
					Value:    "openai",
				},
				Operator: obsDomain.LogicOr,
				Right: &obsDomain.ConditionNode{
					Field:    "gen_ai.system",
					Operator: obsDomain.FilterOpEqual,
					Value:    "anthropic",
				},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"(provider_name = ? OR provider_name = ?)"},
			wantArgCount: 5,
			wantErr:      false,
		},
		{
			name: "nested expression with parentheses",
			node: &obsDomain.BinaryNode{
				Left: &obsDomain.BinaryNode{
					Left: &obsDomain.ConditionNode{
						Field:    "service.name",
						Operator: obsDomain.FilterOpEqual,
						Value:    "api",
					},
					Operator: obsDomain.LogicAnd,
					Right: &obsDomain.ConditionNode{
						Field:    "gen_ai.system",
						Operator: obsDomain.FilterOpEqual,
						Value:    "openai",
					},
				},
				Operator: obsDomain.LogicOr,
				Right: &obsDomain.ConditionNode{
					Field:    "service.name",
					Operator: obsDomain.FilterOpEqual,
					Value:    "worker",
				},
			},
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"((service_name = ? AND provider_name = ?) OR service_name = ?)"},
			wantArgCount: 6,
			wantErr:      false,
		},
		{
			name: "with time range",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpEqual,
				Value:    "chatbot",
			},
			projectID:    "proj123",
			startTime:    timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endTime:      timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
			limit:        100,
			offset:       0,
			wantContains: []string{"start_time >= ?", "start_time <= ?", "service_name = ?"},
			wantArgCount: 6, // projectID, startTime, endTime, filter value, limit, offset
			wantErr:      false,
		},
		{
			name: "nil node - no filter",
			node:         nil,
			projectID:    "proj123",
			limit:        100,
			offset:       0,
			wantContains: []string{"project_id = ?", "deleted_at IS NULL"},
			wantArgCount: 3, // projectID, limit, offset
			wantErr:      false,
		},
		{
			name: "IN clause with invalid value type",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpIn,
				Value:    "not-a-slice", // Should be []string
			},
			projectID: "proj123",
			limit:     100,
			offset:    0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			result, err := builder.BuildQuery(tt.node, tt.projectID, tt.startTime, tt.endTime, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check SQL contains expected fragments
			for _, fragment := range tt.wantContains {
				assert.Contains(t, result.Query, fragment, "SQL should contain: %s", fragment)
			}

			// Check argument count
			assert.Len(t, result.Args, tt.wantArgCount, "Unexpected argument count")

			// Check first args if specified
			if len(tt.wantFirstArgs) > 0 {
				for i, expected := range tt.wantFirstArgs {
					assert.Equal(t, expected, result.Args[i], "Argument %d mismatch", i)
				}
			}

			// Verify query structure
			assert.Contains(t, result.Query, "SELECT")
			assert.Contains(t, result.Query, "FROM otel_traces")
			assert.Contains(t, result.Query, "WHERE")
			assert.Contains(t, result.Query, "ORDER BY start_time DESC")
			assert.Contains(t, result.Query, "LIMIT ? OFFSET ?")
		})
	}
}

func TestSpanQueryBuilder_BuildCountQuery(t *testing.T) {
	tests := []struct {
		name          string
		node          obsDomain.FilterNode
		projectID     string
		startTime     *time.Time
		endTime       *time.Time
		wantContains  []string
		wantArgCount  int
		wantErr       bool
	}{
		{
			name: "simple count query",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpEqual,
				Value:    "chatbot",
			},
			projectID:    "proj123",
			wantContains: []string{"SELECT count(*) as total", "service_name = ?"},
			wantArgCount: 2, // projectID, filter value
			wantErr:      false,
		},
		{
			name: "count with time range",
			node: &obsDomain.ConditionNode{
				Field:    "gen_ai.system",
				Operator: obsDomain.FilterOpEqual,
				Value:    "openai",
			},
			projectID:    "proj123",
			startTime:    timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endTime:      timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
			wantContains: []string{"count(*)", "start_time >= ?", "start_time <= ?", "provider_name = ?"},
			wantArgCount: 4, // projectID, startTime, endTime, filter value
			wantErr:      false,
		},
		{
			name:          "count with nil filter",
			node:          nil,
			projectID:     "proj123",
			wantContains:  []string{"SELECT count(*) as total", "project_id = ?"},
			wantArgCount:  1, // Just projectID
			wantErr:       false,
		},
		{
			name: "complex expression count",
			node: &obsDomain.BinaryNode{
				Left: &obsDomain.ConditionNode{
					Field:    "service.name",
					Operator: obsDomain.FilterOpEqual,
					Value:    "api",
				},
				Operator: obsDomain.LogicAnd,
				Right: &obsDomain.ConditionNode{
					Field:    "gen_ai.usage.total_tokens",
					Operator: obsDomain.FilterOpGreaterThan,
					Value:    float64(1000),
				},
			},
			projectID:    "proj123",
			wantContains: []string{"(service_name = ? AND toFloat64OrNull(span_attributes['gen_ai.usage.total_tokens']) > ?)"},
			wantArgCount: 3, // projectID, 2 filter values
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			result, err := builder.BuildCountQuery(tt.node, tt.projectID, tt.startTime, tt.endTime)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check SQL contains expected fragments
			for _, fragment := range tt.wantContains {
				assert.Contains(t, result.Query, fragment, "SQL should contain: %s", fragment)
			}

			// Check argument count
			assert.Len(t, result.Args, tt.wantArgCount, "Unexpected argument count")

			// Verify query structure - count query should NOT have LIMIT/OFFSET/ORDER BY
			assert.Contains(t, result.Query, "SELECT count(*)")
			assert.Contains(t, result.Query, "FROM otel_traces")
			assert.NotContains(t, result.Query, "LIMIT")
			assert.NotContains(t, result.Query, "OFFSET")
			assert.NotContains(t, result.Query, "ORDER BY")
		})
	}
}

func TestSpanQueryBuilder_GetColumn(t *testing.T) {
	builder := NewSpanQueryBuilder()

	tests := []struct {
		field    string
		expected string
	}{
		// Materialized columns
		{"service.name", "service_name"},
		{"gen_ai.request.model", "model_name"},
		{"gen_ai.system", "provider_name"},
		{"gen_ai.provider.name", "provider_name"},
		{"brokle.span.type", "span_type"},
		{"user.id", "user_id"},
		{"session.id", "session_id"},
		{"span.name", "span_name"},
		{"trace.id", "trace_id"},
		{"span.id", "span_id"},
		{"status.code", "status_code"},

		// Resource attributes
		{"resource.deployment.env", "resource_attributes['resource.deployment.env']"},
		{"deployment.environment", "resource_attributes['deployment.environment']"},

		// Span attributes (default)
		{"custom.field", "span_attributes['custom.field']"},
		{"gen_ai.usage.total_tokens", "span_attributes['gen_ai.usage.total_tokens']"},
		{"my.app.metric", "span_attributes['my.app.metric']"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result, err := builder.getColumn(tt.field)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSpanQueryBuilder_SQLInjectionPrevention(t *testing.T) {
	// All user input should be parameterized, not interpolated
	tests := []struct {
		name         string
		node         obsDomain.FilterNode
		checkQuery   func(t *testing.T, query string)
		checkArgs    func(t *testing.T, args []interface{})
	}{
		{
			name: "malicious value in equality",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpEqual,
				Value:    "'; DROP TABLE otel_traces; --",
			},
			checkQuery: func(t *testing.T, query string) {
				// Value should NOT appear in query
				assert.NotContains(t, query, "DROP TABLE")
				assert.NotContains(t, query, "'; DROP")
				// Should use parameterized placeholder
				assert.Contains(t, query, "service_name = ?")
			},
			checkArgs: func(t *testing.T, args []interface{}) {
				// Malicious string should be safely in args
				found := false
				for _, arg := range args {
					if arg == "'; DROP TABLE otel_traces; --" {
						found = true
						break
					}
				}
				assert.True(t, found, "Malicious value should be safely parameterized")
			},
		},
		{
			name: "malicious value in CONTAINS",
			node: &obsDomain.ConditionNode{
				Field:    "span.name",
				Operator: obsDomain.FilterOpContains,
				Value:    "%'; DELETE FROM otel_traces WHERE '1'='1",
			},
			checkQuery: func(t *testing.T, query string) {
				assert.NotContains(t, query, "DELETE FROM")
				assert.Contains(t, query, "positionCaseInsensitive(span_name, ?) > 0")
			},
			checkArgs: func(t *testing.T, args []interface{}) {
				found := false
				for _, arg := range args {
					if s, ok := arg.(string); ok && strings.Contains(s, "DELETE FROM") {
						found = true
						break
					}
				}
				assert.True(t, found, "Malicious value should be safely parameterized")
			},
		},
		{
			name: "malicious values in IN clause",
			node: &obsDomain.ConditionNode{
				Field:    "service.name",
				Operator: obsDomain.FilterOpIn,
				Value:    []string{"safe", "'; DROP TABLE--", "also_safe"},
			},
			checkQuery: func(t *testing.T, query string) {
				assert.NotContains(t, query, "DROP TABLE")
				assert.Contains(t, query, "IN (?, ?, ?)")
			},
			checkArgs: func(t *testing.T, args []interface{}) {
				// All IN values should be in args
				argStrs := make([]string, 0)
				for _, arg := range args {
					if s, ok := arg.(string); ok {
						argStrs = append(argStrs, s)
					}
				}
				assert.Contains(t, argStrs, "'; DROP TABLE--")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			result, err := builder.BuildQuery(tt.node, "proj123", nil, nil, 100, 0)
			require.NoError(t, err)

			tt.checkQuery(t, result.Query)
			tt.checkArgs(t, result.Args)
		})
	}
}

func TestSpanQueryBuilder_ArgumentOrdering(t *testing.T) {
	// Verify arguments are in correct order for the query
	builder := NewSpanQueryBuilder()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	node := &obsDomain.BinaryNode{
		Left: &obsDomain.ConditionNode{
			Field:    "service.name",
			Operator: obsDomain.FilterOpEqual,
			Value:    "api",
		},
		Operator: obsDomain.LogicAnd,
		Right: &obsDomain.ConditionNode{
			Field:    "gen_ai.system",
			Operator: obsDomain.FilterOpEqual,
			Value:    "openai",
		},
	}

	result, err := builder.BuildQuery(node, "proj123", &startTime, &endTime, 50, 10)
	require.NoError(t, err)

	// Expected order: projectID, startTime, endTime, filter values..., limit, offset
	assert.Equal(t, "proj123", result.Args[0], "First arg should be projectID")
	assert.Equal(t, startTime, result.Args[1], "Second arg should be startTime")
	assert.Equal(t, endTime, result.Args[2], "Third arg should be endTime")
	// Filter values in order of AST traversal (left then right)
	assert.Equal(t, "api", result.Args[3], "Fourth arg should be first filter value")
	assert.Equal(t, "openai", result.Args[4], "Fifth arg should be second filter value")
	assert.Equal(t, 50, result.Args[5], "Sixth arg should be limit")
	assert.Equal(t, 10, result.Args[6], "Seventh arg should be offset")
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestSpanQueryBuilder_FieldNameInjectionPrevention(t *testing.T) {
	tests := []struct {
		name       string
		field      string
		wantErr    bool
		errContain string
	}{
		{
			name:    "valid simple field",
			field:   "service.name",
			wantErr: false,
		},
		{
			name:    "valid nested field",
			field:   "custom.nested.field",
			wantErr: false,
		},
		{
			name:    "valid underscore field",
			field:   "my_custom_field",
			wantErr: false,
		},
		{
			name:       "injection attempt with quote",
			field:      "foo'];DROP TABLE otel_traces--",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection attempt with semicolon",
			field:      "foo;DROP",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection attempt with space",
			field:      "foo bar",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection attempt with parentheses",
			field:      "foo()",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection attempt with comment",
			field:      "foo--comment",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection attempt with equals",
			field:      "foo=1",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "empty field name",
			field:      "",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "field too long",
			field:      strings.Repeat("a", 201),
			wantErr:    true,
			errContain: "too long",
		},
		{
			name:    "field at max length is ok",
			field:   strings.Repeat("a", 200),
			wantErr: false,
		},
		{
			name:       "field starting with number",
			field:      "123field",
			wantErr:    true,
			errContain: "invalid field name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			node := &obsDomain.ConditionNode{
				Field:    tt.field,
				Operator: obsDomain.FilterOpEqual,
				Value:    "testvalue",
			}

			result, err := builder.BuildQuery(node, "proj123", nil, nil, 100, 0)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				// Verify the field doesn't appear unescaped in query
				assert.NotContains(t, result.Query, "DROP TABLE")
			}
		})
	}
}

func TestSpanQueryBuilder_OperatorInjectionPrevention(t *testing.T) {
	tests := []struct {
		name       string
		operator   obsDomain.LogicOperator
		wantErr    bool
		errContain string
	}{
		{
			name:     "valid AND operator",
			operator: obsDomain.LogicAnd,
			wantErr:  false,
		},
		{
			name:     "valid OR operator",
			operator: obsDomain.LogicOr,
			wantErr:  false,
		},
		{
			name:       "injection via operator",
			operator:   obsDomain.LogicOperator("OR 1=1; DROP TABLE otel_traces; --"),
			wantErr:    true,
			errContain: "invalid logic operator",
		},
		{
			name:       "empty operator",
			operator:   obsDomain.LogicOperator(""),
			wantErr:    true,
			errContain: "invalid logic operator",
		},
		{
			name:       "unknown operator",
			operator:   obsDomain.LogicOperator("XOR"),
			wantErr:    true,
			errContain: "invalid logic operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			node := &obsDomain.BinaryNode{
				Left: &obsDomain.ConditionNode{
					Field:    "service.name",
					Operator: obsDomain.FilterOpEqual,
					Value:    "api",
				},
				Operator: tt.operator,
				Right: &obsDomain.ConditionNode{
					Field:    "gen_ai.system",
					Operator: obsDomain.FilterOpEqual,
					Value:    "openai",
				},
			}

			result, err := builder.BuildQuery(node, "proj123", nil, nil, 100, 0)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestSpanQueryBuilder_ExistsFieldInjectionPrevention(t *testing.T) {
	tests := []struct {
		name       string
		field      string
		wantErr    bool
		errContain string
	}{
		{
			name:    "valid resource field",
			field:   "resource.service.name",
			wantErr: false,
		},
		{
			name:       "injection in exists check",
			field:      "foo');DROP TABLE otel_traces--",
			wantErr:    true,
			errContain: "invalid field name",
		},
		{
			name:       "injection with single quote",
			field:      "foo'",
			wantErr:    true,
			errContain: "invalid field name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSpanQueryBuilder()
			node := &obsDomain.ConditionNode{
				Field:    tt.field,
				Operator: obsDomain.FilterOpExists,
				Value:    nil,
			}

			result, err := builder.BuildQuery(node, "proj123", nil, nil, 100, 0)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				// Verify field appears properly escaped
				assert.NotContains(t, result.Query, "DROP TABLE")
			}
		})
	}
}

func TestValidateFieldName(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		wantErr bool
	}{
		{"valid simple", "field", false},
		{"valid with underscore", "my_field", false},
		{"valid with dot", "service.name", false},
		{"valid with number", "field1", false},
		{"valid starting with underscore", "_private", false},
		{"invalid empty", "", true},
		{"invalid with space", "field name", true},
		{"invalid with quote", "field'", true},
		{"invalid with semicolon", "field;", true},
		{"invalid with paren", "field()", true},
		{"invalid starting with number", "1field", true},
		{"invalid with dash", "field-name", true},
		{"too long", strings.Repeat("x", 201), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldName(tt.field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEscapeAttributeKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with'quote", "with''quote"},
		{"many'''quotes", "many''''''quotes"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeAttributeKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
