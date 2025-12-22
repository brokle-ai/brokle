package observability

import (
	"fmt"
	"strings"
	"time"

	obsDomain "brokle/internal/core/domain/observability"
)

// SpanQueryBuilder converts filter AST to ClickHouse SQL with parameterized queries.
type SpanQueryBuilder struct {
	paramCount int
}

// NewSpanQueryBuilder creates a new query builder.
func NewSpanQueryBuilder() *SpanQueryBuilder {
	return &SpanQueryBuilder{}
}

// QueryResult contains the built query and its arguments.
type QueryResult struct {
	Query string
	Args  []interface{}
	Count int // number of conditions
}

// BuildQuery generates a parameterized ClickHouse query from a filter AST.
func (b *SpanQueryBuilder) BuildQuery(
	node obsDomain.FilterNode,
	projectID string,
	startTime, endTime *time.Time,
	limit, offset int,
) (*QueryResult, error) {
	b.paramCount = 0

	whereClause, args, err := b.buildNode(node)
	if err != nil {
		return nil, err
	}

	conditions := []string{"project_id = ?", "deleted_at IS NULL"}
	baseArgs := []interface{}{projectID}

	if startTime != nil {
		conditions = append(conditions, "start_time >= ?")
		baseArgs = append(baseArgs, *startTime)
	}
	if endTime != nil {
		conditions = append(conditions, "start_time <= ?")
		baseArgs = append(baseArgs, *endTime)
	}

	if whereClause != "" {
		conditions = append(conditions, "("+whereClause+")")
	}

	allArgs := append(baseArgs, args...)

	query := fmt.Sprintf(`
		SELECT %s
		FROM otel_traces
		WHERE %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, obsDomain.SpanSelectFields, strings.Join(conditions, " AND "))

	allArgs = append(allArgs, limit, offset)

	return &QueryResult{
		Query: query,
		Args:  allArgs,
		Count: b.paramCount,
	}, nil
}

// BuildCountQuery generates a COUNT query for pagination.
func (b *SpanQueryBuilder) BuildCountQuery(
	node obsDomain.FilterNode,
	projectID string,
	startTime, endTime *time.Time,
) (*QueryResult, error) {
	b.paramCount = 0

	whereClause, args, err := b.buildNode(node)
	if err != nil {
		return nil, err
	}

	conditions := []string{"project_id = ?", "deleted_at IS NULL"}
	baseArgs := []interface{}{projectID}

	if startTime != nil {
		conditions = append(conditions, "start_time >= ?")
		baseArgs = append(baseArgs, *startTime)
	}
	if endTime != nil {
		conditions = append(conditions, "start_time <= ?")
		baseArgs = append(baseArgs, *endTime)
	}

	if whereClause != "" {
		conditions = append(conditions, "("+whereClause+")")
	}

	allArgs := append(baseArgs, args...)

	query := fmt.Sprintf(`
		SELECT count(*) as total
		FROM otel_traces
		WHERE %s
	`, strings.Join(conditions, " AND "))

	return &QueryResult{
		Query: query,
		Args:  allArgs,
		Count: b.paramCount,
	}, nil
}

// buildNode recursively builds SQL from a filter AST node.
func (b *SpanQueryBuilder) buildNode(node obsDomain.FilterNode) (string, []interface{}, error) {
	if node == nil {
		return "", nil, nil
	}

	switch n := node.(type) {
	case *obsDomain.BinaryNode:
		return b.buildBinaryNode(n)
	case *obsDomain.ConditionNode:
		return b.buildConditionNode(n)
	default:
		return "", nil, fmt.Errorf("unknown node type: %T", node)
	}
}

// buildBinaryNode handles AND/OR binary expressions.
func (b *SpanQueryBuilder) buildBinaryNode(node *obsDomain.BinaryNode) (string, []interface{}, error) {
	leftSQL, leftArgs, err := b.buildNode(node.Left)
	if err != nil {
		return "", nil, err
	}

	rightSQL, rightArgs, err := b.buildNode(node.Right)
	if err != nil {
		return "", nil, err
	}

	sql := fmt.Sprintf("(%s %s %s)", leftSQL, node.Operator, rightSQL)
	args := append(leftArgs, rightArgs...)

	return sql, args, nil
}

// buildConditionNode converts a single condition to SQL.
func (b *SpanQueryBuilder) buildConditionNode(node *obsDomain.ConditionNode) (string, []interface{}, error) {
	column := b.getColumn(node.Field)

	switch node.Operator {
	case obsDomain.FilterOpEqual:
		return b.buildComparison(column, "=", node.Value)

	case obsDomain.FilterOpNotEqual:
		return b.buildComparison(column, "!=", node.Value)

	case obsDomain.FilterOpGreaterThan:
		return b.buildNumericComparison(column, node.Field, ">", node.Value)

	case obsDomain.FilterOpLessThan:
		return b.buildNumericComparison(column, node.Field, "<", node.Value)

	case obsDomain.FilterOpGreaterOrEqual:
		return b.buildNumericComparison(column, node.Field, ">=", node.Value)

	case obsDomain.FilterOpLessOrEqual:
		return b.buildNumericComparison(column, node.Field, "<=", node.Value)

	case obsDomain.FilterOpContains:
		return b.buildContains(column, node.Value, false)

	case obsDomain.FilterOpNotContains:
		return b.buildContains(column, node.Value, true)

	case obsDomain.FilterOpIn:
		return b.buildInClause(column, node.Value, false)

	case obsDomain.FilterOpNotIn:
		return b.buildInClause(column, node.Value, true)

	case obsDomain.FilterOpExists:
		return b.buildExists(node.Field, false)

	case obsDomain.FilterOpNotExists:
		return b.buildExists(node.Field, true)

	default:
		return "", nil, obsDomain.NewUnsupportedOperatorError(string(node.Operator))
	}
}

// getColumn returns the ClickHouse column for a field path.
func (b *SpanQueryBuilder) getColumn(field string) string {
	if col := obsDomain.GetMaterializedColumn(field); col != "" {
		return col
	}

	if strings.HasPrefix(field, "resource.") || strings.HasPrefix(field, "deployment.") {
		return fmt.Sprintf("resource_attributes['%s']", field)
	}

	return fmt.Sprintf("span_attributes['%s']", field)
}

// buildComparison builds a simple comparison (=, !=).
func (b *SpanQueryBuilder) buildComparison(column, op string, value interface{}) (string, []interface{}, error) {
	b.paramCount++
	return fmt.Sprintf("%s %s ?", column, op), []interface{}{value}, nil
}

// buildNumericComparison builds a numeric comparison with type coercion.
func (b *SpanQueryBuilder) buildNumericComparison(column, field, op string, value interface{}) (string, []interface{}, error) {
	b.paramCount++

	if obsDomain.IsMaterializedColumn(field) {
		return fmt.Sprintf("%s %s ?", column, op), []interface{}{value}, nil
	}

	// toFloat64OrNull handles non-numeric values gracefully for map columns
	return fmt.Sprintf("toFloat64OrNull(%s) %s ?", column, op), []interface{}{value}, nil
}

// buildContains builds a case-insensitive substring search.
// Uses positionCaseInsensitive for efficient ClickHouse substring search.
func (b *SpanQueryBuilder) buildContains(column string, value interface{}, negated bool) (string, []interface{}, error) {
	b.paramCount++

	if negated {
		return fmt.Sprintf("positionCaseInsensitive(%s, ?) = 0", column), []interface{}{value}, nil
	}
	return fmt.Sprintf("positionCaseInsensitive(%s, ?) > 0", column), []interface{}{value}, nil
}

// buildInClause builds an IN clause with array parameter.
func (b *SpanQueryBuilder) buildInClause(column string, value interface{}, negated bool) (string, []interface{}, error) {
	values, ok := value.([]string)
	if !ok {
		return "", nil, obsDomain.ErrInvalidValue
	}

	if len(values) == 0 {
		if negated {
			return "1=1", nil, nil // NOT IN empty set is always true
		}
		return "1=0", nil, nil // IN empty set is always false
	}

	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = "?"
		args[i] = v
		b.paramCount++
	}

	op := "IN"
	if negated {
		op = "NOT IN"
	}

	return fmt.Sprintf("%s %s (%s)", column, op, strings.Join(placeholders, ", ")), args, nil
}

// buildExists builds an EXISTS check using mapContains for efficient ClickHouse existence checks.
func (b *SpanQueryBuilder) buildExists(field string, negated bool) (string, []interface{}, error) {
	if obsDomain.IsMaterializedColumn(field) {
		col := obsDomain.GetMaterializedColumn(field)
		if negated {
			return fmt.Sprintf("(%s IS NULL OR %s = '')", col, col), nil, nil
		}
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", col, col), nil, nil
	}

	mapName := "span_attributes"
	attrKey := field
	if strings.HasPrefix(field, "resource.") {
		mapName = "resource_attributes"
		attrKey = field
	}

	if negated {
		return fmt.Sprintf("NOT mapContains(%s, '%s')", mapName, attrKey), nil, nil
	}
	return fmt.Sprintf("mapContains(%s, '%s')", mapName, attrKey), nil, nil
}
