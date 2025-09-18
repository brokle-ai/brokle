package clickhouse

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/infrastructure/database"
)

// AnalyticsRepository handles analytics data operations in ClickHouse
type AnalyticsRepository struct {
	db *database.ClickHouseDB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *database.ClickHouseDB) *AnalyticsRepository {
	return &AnalyticsRepository{
		db: db,
	}
}

// RequestLog represents an AI request log entry
type RequestLog struct {
	ID             string    `ch:"id"`
	UserID         string    `ch:"user_id"`
	OrganizationID string    `ch:"organization_id"`
	ProjectID      string    `ch:"project_id"`
	Environment    string    `ch:"environment"`
	Provider       string    `ch:"provider"`
	Model          string    `ch:"model"`
	Method         string    `ch:"method"`
	Endpoint       string    `ch:"endpoint"`
	StatusCode     int       `ch:"status_code"`
	InputTokens    int64     `ch:"input_tokens"`
	OutputTokens   int64     `ch:"output_tokens"`
	TotalTokens    int64     `ch:"total_tokens"`
	Cost           float64   `ch:"cost"`
	Latency        int64     `ch:"latency"`
	QualityScore   float64   `ch:"quality_score"`
	Cached         bool      `ch:"cached"`
	Error          string    `ch:"error"`
	Timestamp      time.Time `ch:"timestamp"`
}

// MetricPoint represents a time-series metric point
type MetricPoint struct {
	Timestamp      time.Time `ch:"timestamp"`
	OrganizationID string    `ch:"organization_id"`
	ProjectID      string    `ch:"project_id"`
	Environment    string    `ch:"environment"`
	MetricName     string    `ch:"metric_name"`
	MetricValue    float64   `ch:"metric_value"`
	Tags           string    `ch:"tags"`
}

// ProviderHealth represents provider health metrics
type ProviderHealth struct {
	Provider       string    `ch:"provider"`
	Model          string    `ch:"model"`
	SuccessRate    float64   `ch:"success_rate"`
	AverageLatency float64   `ch:"average_latency"`
	ErrorRate      float64   `ch:"error_rate"`
	Timestamp      time.Time `ch:"timestamp"`
}

// InsertRequestLog inserts a new request log
func (r *AnalyticsRepository) InsertRequestLog(ctx context.Context, log *RequestLog) error {
	query := `
		INSERT INTO request_logs (
			id, user_id, organization_id, project_id, environment,
			provider, model, method, endpoint, status_code,
			input_tokens, output_tokens, total_tokens, cost, latency,
			quality_score, cached, error, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	return r.db.Execute(ctx, query,
		log.ID, log.UserID, log.OrganizationID, log.ProjectID, log.Environment,
		log.Provider, log.Model, log.Method, log.Endpoint, log.StatusCode,
		log.InputTokens, log.OutputTokens, log.TotalTokens, log.Cost, log.Latency,
		log.QualityScore, log.Cached, log.Error, log.Timestamp,
	)
}

// InsertBatchRequestLogs inserts multiple request logs efficiently
func (r *AnalyticsRepository) InsertBatchRequestLogs(ctx context.Context, logs []*RequestLog) error {
	if len(logs) == 0 {
		return nil
	}

	batch := make([][]interface{}, len(logs))
	for i, log := range logs {
		batch[i] = []interface{}{
			log.ID, log.UserID, log.OrganizationID, log.ProjectID, log.Environment,
			log.Provider, log.Model, log.Method, log.Endpoint, log.StatusCode,
			log.InputTokens, log.OutputTokens, log.TotalTokens, log.Cost, log.Latency,
			log.QualityScore, log.Cached, log.Error, log.Timestamp,
		}
	}

	query := `
		INSERT INTO request_logs (
			id, user_id, organization_id, project_id, environment,
			provider, model, method, endpoint, status_code,
			input_tokens, output_tokens, total_tokens, cost, latency,
			quality_score, cached, error, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range batch {
		if err := r.db.Execute(ctx, query, row...); err != nil {
			return fmt.Errorf("failed to insert request log batch: %w", err)
		}
	}

	return nil
}

// GetRequestStats retrieves request statistics for a time period
func (r *AnalyticsRepository) GetRequestStats(ctx context.Context, filter AnalyticsFilter) (*RequestStats, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			COUNT() as total_requests,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as successful_requests,
			AVG(latency) as avg_latency,
			SUM(cost) as total_cost,
			SUM(total_tokens) as total_tokens,
			AVG(quality_score) as avg_quality_score,
			SUM(CASE WHEN cached = true THEN 1 ELSE 0 END) as cache_hits
		FROM request_logs 
		%s`, whereClause)

	row := r.db.QueryRow(ctx, query, args...)
	
	var stats RequestStats
	if err := row.Scan(
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.AverageLatency,
		&stats.TotalCost,
		&stats.TotalTokens,
		&stats.AverageQualityScore,
		&stats.CacheHits,
	); err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}

	// Calculate derived metrics
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalRequests) * 100
	}

	return &stats, nil
}

// GetProviderStats retrieves provider-wise statistics
func (r *AnalyticsRepository) GetProviderStats(ctx context.Context, filter AnalyticsFilter) ([]*ProviderStats, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			provider,
			COUNT() as total_requests,
			AVG(latency) as avg_latency,
			SUM(cost) as total_cost,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as successful_requests,
			AVG(quality_score) as avg_quality_score
		FROM request_logs 
		%s
		GROUP BY provider
		ORDER BY total_requests DESC`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()

	var stats []*ProviderStats
	for rows.Next() {
		var stat ProviderStats
		if err := rows.Scan(
			&stat.Provider,
			&stat.TotalRequests,
			&stat.AverageLatency,
			&stat.TotalCost,
			&stat.SuccessfulRequests,
			&stat.AverageQualityScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan provider stats: %w", err)
		}

		// Calculate success rate
		if stat.TotalRequests > 0 {
			stat.SuccessRate = float64(stat.SuccessfulRequests) / float64(stat.TotalRequests) * 100
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetTimeSeriesData retrieves time-series data for charting
func (r *AnalyticsRepository) GetTimeSeriesData(ctx context.Context, filter AnalyticsFilter, interval string, metric string) ([]*TimeSeriesPoint, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	// Convert interval to ClickHouse interval function
	var intervalFunc string
	switch interval {
	case "hour":
		intervalFunc = "toStartOfHour(timestamp)"
	case "day":
		intervalFunc = "toDate(timestamp)"
	case "week":
		intervalFunc = "toMonday(timestamp)"
	case "month":
		intervalFunc = "toStartOfMonth(timestamp)"
	default:
		intervalFunc = "toStartOfHour(timestamp)"
	}

	var aggregateFunc string
	switch metric {
	case "requests":
		aggregateFunc = "COUNT()"
	case "cost":
		aggregateFunc = "SUM(cost)"
	case "latency":
		aggregateFunc = "AVG(latency)"
	case "tokens":
		aggregateFunc = "SUM(total_tokens)"
	case "quality_score":
		aggregateFunc = "AVG(quality_score)"
	default:
		aggregateFunc = "COUNT()"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as time_bucket,
			%s as value
		FROM request_logs 
		%s
		GROUP BY time_bucket
		ORDER BY time_bucket`, intervalFunc, aggregateFunc, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series data: %w", err)
	}
	defer rows.Close()

	var points []*TimeSeriesPoint
	for rows.Next() {
		var point TimeSeriesPoint
		if err := rows.Scan(&point.Timestamp, &point.Value); err != nil {
			return nil, fmt.Errorf("failed to scan time series point: %w", err)
		}
		points = append(points, &point)
	}

	return points, nil
}

// GetCostBreakdown retrieves cost breakdown by different dimensions
func (r *AnalyticsRepository) GetCostBreakdown(ctx context.Context, filter AnalyticsFilter, groupBy string) ([]*CostBreakdown, error) {
	whereClause, args := r.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT 
			%s as dimension,
			SUM(cost) as total_cost,
			COUNT() as request_count,
			AVG(cost) as avg_cost_per_request
		FROM request_logs 
		%s
		GROUP BY %s
		ORDER BY total_cost DESC
		LIMIT 50`, groupBy, whereClause, groupBy)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []*CostBreakdown
	for rows.Next() {
		var item CostBreakdown
		if err := rows.Scan(
			&item.Dimension,
			&item.TotalCost,
			&item.RequestCount,
			&item.AverageCostPerRequest,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cost breakdown: %w", err)
		}
		breakdown = append(breakdown, &item)
	}

	return breakdown, nil
}

// InsertMetric inserts a metric point
func (r *AnalyticsRepository) InsertMetric(ctx context.Context, metric *MetricPoint) error {
	query := `
		INSERT INTO metrics (
			timestamp, organization_id, project_id, environment,
			metric_name, metric_value, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.db.Execute(ctx, query,
		metric.Timestamp, metric.OrganizationID, metric.ProjectID, metric.Environment,
		metric.MetricName, metric.MetricValue, metric.Tags,
	)
}

// buildWhereClause builds WHERE clause from filter
func (r *AnalyticsRepository) buildWhereClause(filter AnalyticsFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "timestamp >= ? AND timestamp <= ?")
	args = append(args, filter.StartTime, filter.EndTime)

	if filter.OrganizationID != "" {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, filter.OrganizationID)
	}

	if filter.ProjectID != "" {
		conditions = append(conditions, "project_id = ?")
		args = append(args, filter.ProjectID)
	}

	if filter.Environment != "" {
		conditions = append(conditions, "environment = ?")
		args = append(args, filter.Environment)
	}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, filter.Provider)
	}

	if filter.Model != "" {
		conditions = append(conditions, "model = ?")
		args = append(args, filter.Model)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0])) + 
		func() string {
			if len(conditions) > 1 {
				return " AND " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[1:]))
			}
			return ""
		}(), args
}

// Data structures for analytics results

type AnalyticsFilter struct {
	StartTime      time.Time
	EndTime        time.Time
	OrganizationID string
	ProjectID      string
	Environment    string
	UserID         string
	Provider       string
	Model          string
}

type RequestStats struct {
	TotalRequests        int64   `json:"total_requests"`
	SuccessfulRequests   int64   `json:"successful_requests"`
	SuccessRate          float64 `json:"success_rate"`
	AverageLatency       float64 `json:"average_latency"`
	TotalCost            float64 `json:"total_cost"`
	TotalTokens          int64   `json:"total_tokens"`
	AverageQualityScore  float64 `json:"average_quality_score"`
	CacheHits            int64   `json:"cache_hits"`
	CacheHitRate         float64 `json:"cache_hit_rate"`
}

type ProviderStats struct {
	Provider             string  `json:"provider"`
	TotalRequests        int64   `json:"total_requests"`
	SuccessfulRequests   int64   `json:"successful_requests"`
	SuccessRate          float64 `json:"success_rate"`
	AverageLatency       float64 `json:"average_latency"`
	TotalCost            float64 `json:"total_cost"`
	AverageQualityScore  float64 `json:"average_quality_score"`
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type CostBreakdown struct {
	Dimension             string  `json:"dimension"`
	TotalCost             float64 `json:"total_cost"`
	RequestCount          int64   `json:"request_count"`
	AverageCostPerRequest float64 `json:"average_cost_per_request"`
}