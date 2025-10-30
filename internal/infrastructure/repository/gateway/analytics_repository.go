package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/internal/workers/analytics"
	"brokle/pkg/ulid"
)

// Repository implements the analytics repository using ClickHouse
type Repository struct {
	conn   clickhouse.Conn
	logger *logrus.Logger
}

// NewRepository creates a new analytics repository instance
func NewRepository(conn clickhouse.Conn, logger *logrus.Logger) *Repository {
	return &Repository{
		conn:   conn,
		logger: logger,
	}
}

// BatchInsertRequestMetrics inserts request metrics in batch for performance
func (r *Repository) BatchInsertRequestMetrics(ctx context.Context, metrics []*analytics.RequestMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO gateway_request_metrics (
			id, request_id, organization_id, user_id,
			provider_id, provider_name, model_id, model_name,
			request_type, method, endpoint, status, status_code,
			duration, input_tokens, output_tokens, total_tokens,
			estimated_cost, actual_cost, currency, routing_reason,
			cache_hit, error, metadata, timestamp
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?
		)`

	batch, err := r.conn.PrepareBatch(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to prepare batch for request metrics")
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, metric := range metrics {
		var userID interface{}
		if metric.UserID != nil {
			userID = metric.UserID.String()
		} else {
			userID = nil
		}

		// Convert metadata to JSON string
		var metadataJSON string
		if metric.Metadata != nil && len(metric.Metadata) > 0 {
			// Simple JSON serialization - in production you might want proper JSON marshaling
			metadataJSON = fmt.Sprintf("%v", metric.Metadata)
		}

		err := batch.Append(
			metric.ID.String(),
			metric.RequestID.String(),
			metric.OrganizationID.String(),
			userID,
			metric.ProviderID.String(),
			metric.ProviderName,
			metric.ModelID.String(),
			metric.ModelName,
			string(metric.RequestType),
			metric.Method,
			metric.Endpoint,
			metric.Status,
			metric.StatusCode,
			metric.Duration.Nanoseconds(),
			metric.InputTokens,
			metric.OutputTokens,
			metric.TotalTokens,
			metric.EstimatedCost,
			metric.ActualCost,
			metric.Currency,
			metric.RoutingReason,
			metric.CacheHit,
			metric.Error,
			metadataJSON,
			metric.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).WithField("metric_id", metric.ID).Error("Failed to append request metric to batch")
			return fmt.Errorf("failed to append metric to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.WithError(err).WithField("batch_size", len(metrics)).Error("Failed to send request metrics batch")
		return fmt.Errorf("failed to send batch: %w", err)
	}

	r.logger.WithField("batch_size", len(metrics)).Debug("Successfully inserted request metrics batch")
	return nil
}

// BatchInsertUsageMetrics inserts usage metrics in batch
func (r *Repository) BatchInsertUsageMetrics(ctx context.Context, metrics []*analytics.UsageMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO gateway_usage_metrics (
			id, organization_id, provider_id, model_id,
			request_type, period, period_start, period_end,
			request_count, success_count, error_count,
			total_input_tokens, total_output_tokens, total_tokens,
			total_cost, currency, avg_duration, min_duration, max_duration,
			cache_hit_rate, timestamp
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?
		)`

	batch, err := r.conn.PrepareBatch(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to prepare batch for usage metrics")
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, metric := range metrics {
		err := batch.Append(
			metric.ID.String(),
			metric.OrganizationID.String(),
			metric.ProviderID.String(),
			metric.ModelID.String(),
			string(metric.RequestType),
			metric.Period,
			metric.PeriodStart,
			metric.PeriodEnd,
			metric.RequestCount,
			metric.SuccessCount,
			metric.ErrorCount,
			metric.TotalInputTokens,
			metric.TotalOutputTokens,
			metric.TotalTokens,
			metric.TotalCost,
			metric.Currency,
			metric.AvgDuration,
			metric.MinDuration.Nanoseconds(),
			metric.MaxDuration.Nanoseconds(),
			metric.CacheHitRate,
			metric.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).WithField("metric_id", metric.ID).Error("Failed to append usage metric to batch")
			return fmt.Errorf("failed to append metric to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.WithError(err).WithField("batch_size", len(metrics)).Error("Failed to send usage metrics batch")
		return fmt.Errorf("failed to send batch: %w", err)
	}

	r.logger.WithField("batch_size", len(metrics)).Debug("Successfully inserted usage metrics batch")
	return nil
}

// BatchInsertCostMetrics inserts cost metrics in batch
func (r *Repository) BatchInsertCostMetrics(ctx context.Context, metrics []*analytics.CostMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO gateway_cost_metrics (
			id, request_id, organization_id,
			provider_id, model_id, request_type,
			input_tokens, output_tokens, total_tokens,
			input_cost, output_cost, total_cost,
			estimated_cost, cost_difference, currency,
			billing_tier, discount_applied, timestamp
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?
		)`

	batch, err := r.conn.PrepareBatch(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to prepare batch for cost metrics")
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, metric := range metrics {
		err := batch.Append(
			metric.ID.String(),
			metric.RequestID.String(),
			metric.OrganizationID.String(),
			metric.ProviderID.String(),
			metric.ModelID.String(),
			string(metric.RequestType),
			metric.InputTokens,
			metric.OutputTokens,
			metric.TotalTokens,
			metric.InputCost,
			metric.OutputCost,
			metric.TotalCost,
			metric.EstimatedCost,
			metric.CostDifference,
			metric.Currency,
			metric.BillingTier,
			metric.DiscountApplied,
			metric.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).WithField("metric_id", metric.ID).Error("Failed to append cost metric to batch")
			return fmt.Errorf("failed to append metric to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.WithError(err).WithField("batch_size", len(metrics)).Error("Failed to send cost metrics batch")
		return fmt.Errorf("failed to send batch: %w", err)
	}

	r.logger.WithField("batch_size", len(metrics)).Debug("Successfully inserted cost metrics batch")
	return nil
}

// GetUsageStats retrieves usage statistics for reporting
func (r *Repository) GetUsageStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*analytics.UsageMetric, error) {
	query := `
		SELECT 
			id, organization_id, provider_id, model_id,
			request_type, period, period_start, period_end,
			request_count, success_count, error_count,
			total_input_tokens, total_output_tokens, total_tokens,
			total_cost, currency, avg_duration, min_duration, max_duration,
			cache_hit_rate, timestamp
		FROM gateway_usage_metrics
		WHERE organization_id = ?
			AND period = ?
			AND period_start >= ?
			AND period_start < ?
		ORDER BY period_start DESC`

	rows, err := r.conn.Query(ctx, query, orgID.String(), period, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage stats: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.UsageMetric
	for rows.Next() {
		metric := &analytics.UsageMetric{}
		var (
			id, orgIDStr, providerIDStr, modelIDStr string
			reqTypeStr                              string
			minDurationNs, maxDurationNs                    int64
		)

		err := rows.Scan(
			&id,
			&orgIDStr,
			&providerIDStr,
			&modelIDStr,
			&reqTypeStr,
			&metric.Period,
			&metric.PeriodStart,
			&metric.PeriodEnd,
			&metric.RequestCount,
			&metric.SuccessCount,
			&metric.ErrorCount,
			&metric.TotalInputTokens,
			&metric.TotalOutputTokens,
			&metric.TotalTokens,
			&metric.TotalCost,
			&metric.Currency,
			&metric.AvgDuration,
			&minDurationNs,
			&maxDurationNs,
			&metric.CacheHitRate,
			&metric.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan usage metric: %w", err)
		}

		// Parse ULIDs
		if metric.ID, err = ulid.Parse(id); err != nil {
			return nil, fmt.Errorf("failed to parse metric ID: %w", err)
		}
		if metric.OrganizationID, err = ulid.Parse(orgIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse organization ID: %w", err)
		}
		if metric.ProviderID, err = ulid.Parse(providerIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse provider ID: %w", err)
		}
		if metric.ModelID, err = ulid.Parse(modelIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse model ID: %w", err)
		}

		// Parse enums and durations
		metric.RequestType = gateway.RequestType(reqTypeStr)
		metric.MinDuration = time.Duration(minDurationNs)
		metric.MaxDuration = time.Duration(maxDurationNs)

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage stats rows: %w", err)
	}

	return metrics, nil
}

// GetCostStats retrieves cost statistics for reporting
func (r *Repository) GetCostStats(ctx context.Context, orgID ulid.ULID, period string, start, end time.Time) ([]*analytics.CostMetric, error) {
	query := `
		SELECT 
			id, request_id, organization_id,
			provider_id, model_id, request_type,
			input_tokens, output_tokens, total_tokens,
			input_cost, output_cost, total_cost,
			estimated_cost, cost_difference, currency,
			billing_tier, discount_applied, timestamp
		FROM gateway_cost_metrics
		WHERE organization_id = ?
			AND timestamp >= ?
			AND timestamp < ?
		ORDER BY timestamp DESC
		LIMIT 10000`

	rows, err := r.conn.Query(ctx, query, orgID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost stats: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.CostMetric
	for rows.Next() {
		metric := &analytics.CostMetric{}
		var (
			id, reqIDStr, orgIDStr string
			providerIDStr, modelIDStr      string
			reqTypeStr                     string
		)

		err := rows.Scan(
			&id,
			&reqIDStr,
			&orgIDStr,
			&providerIDStr,
			&modelIDStr,
			&reqTypeStr,
			&metric.InputTokens,
			&metric.OutputTokens,
			&metric.TotalTokens,
			&metric.InputCost,
			&metric.OutputCost,
			&metric.TotalCost,
			&metric.EstimatedCost,
			&metric.CostDifference,
			&metric.Currency,
			&metric.BillingTier,
			&metric.DiscountApplied,
			&metric.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan cost metric: %w", err)
		}

		// Parse ULIDs
		if metric.ID, err = ulid.Parse(id); err != nil {
			return nil, fmt.Errorf("failed to parse metric ID: %w", err)
		}
		if metric.RequestID, err = ulid.Parse(reqIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse request ID: %w", err)
		}
		if metric.OrganizationID, err = ulid.Parse(orgIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse organization ID: %w", err)
		}
		if metric.ProviderID, err = ulid.Parse(providerIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse provider ID: %w", err)
		}
		if metric.ModelID, err = ulid.Parse(modelIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse model ID: %w", err)
		}

		// Parse enums
		metric.RequestType = gateway.RequestType(reqTypeStr)

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cost stats rows: %w", err)
	}

	return metrics, nil
}

// GetRequestMetrics retrieves request metrics for analysis
func (r *Repository) GetRequestMetrics(ctx context.Context, orgID ulid.ULID, start, end time.Time, limit int) ([]*analytics.RequestMetric, error) {
	if limit <= 0 {
		limit = 1000
	}

	query := `
		SELECT 
			id, request_id, organization_id, user_id,
			provider_id, provider_name, model_id, model_name,
			request_type, method, endpoint, status, status_code,
			duration, input_tokens, output_tokens, total_tokens,
			estimated_cost, actual_cost, currency, routing_reason,
			cache_hit, error, metadata, timestamp
		FROM gateway_request_metrics
		WHERE organization_id = ?
			AND timestamp >= ?
			AND timestamp < ?
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := r.conn.Query(ctx, query, orgID.String(), start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query request metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.RequestMetric
	for rows.Next() {
		metric := &analytics.RequestMetric{}
		var (
			id, reqIDStr, orgIDStr string
			userIDStr              sql.NullString
			durationNs             int64
			metadataStr            string
		)

		err := rows.Scan(
			&id,
			&reqIDStr,
			&orgIDStr,
			&userIDStr,
			&metric.ProviderID,
			&metric.ProviderName,
			&metric.ModelID,
			&metric.ModelName,
			&metric.RequestType,
			&metric.Method,
			&metric.Endpoint,
			&metric.Status,
			&metric.StatusCode,
			&durationNs,
			&metric.InputTokens,
			&metric.OutputTokens,
			&metric.TotalTokens,
			&metric.EstimatedCost,
			&metric.ActualCost,
			&metric.Currency,
			&metric.RoutingReason,
			&metric.CacheHit,
			&metric.Error,
			&metadataStr,
			&metric.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan request metric: %w", err)
		}

		// Parse ULIDs
		if metric.ID, err = ulid.Parse(id); err != nil {
			return nil, fmt.Errorf("failed to parse metric ID: %w", err)
		}
		if metric.RequestID, err = ulid.Parse(reqIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse request ID: %w", err)
		}
		if metric.OrganizationID, err = ulid.Parse(orgIDStr); err != nil {
			return nil, fmt.Errorf("failed to parse organization ID: %w", err)
		}

		// Handle nullable user ID
		if userIDStr.Valid {
			if userID, err := ulid.Parse(userIDStr.String); err == nil {
				metric.UserID = &userID
			}
		}

		// Parse duration
		metric.Duration = time.Duration(durationNs)

		// Parse metadata (simplified - in production you might want proper JSON unmarshaling)
		if metadataStr != "" {
			metric.Metadata = map[string]interface{}{
				"raw": metadataStr,
			}
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating request metrics rows: %w", err)
	}

	return metrics, nil
}

// GetOrganizationStats retrieves aggregated statistics for an organization
func (r *Repository) GetOrganizationStats(ctx context.Context, orgID ulid.ULID, start, end time.Time) (*OrganizationStats, error) {
	query := `
		SELECT 
			count() as total_requests,
			countIf(status_code < 400) as successful_requests,
			countIf(status_code >= 400) as failed_requests,
			sum(input_tokens) as total_input_tokens,
			sum(output_tokens) as total_output_tokens,
			sum(total_tokens) as total_tokens,
			sum(actual_cost) as total_cost,
			avg(duration) as avg_duration,
			countIf(cache_hit = 1) as cache_hits,
			uniqExact(provider_id) as unique_providers,
			uniqExact(model_id) as unique_models
		FROM gateway_request_metrics
		WHERE organization_id = ?
			AND timestamp >= ?
			AND timestamp < ?`

	stats := &OrganizationStats{}
	var avgDurationNs int64

	err := r.conn.QueryRow(ctx, query, orgID.String(), start, end).Scan(
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.FailedRequests,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalTokens,
		&stats.TotalCost,
		&avgDurationNs,
		&stats.CacheHits,
		&stats.UniqueProviders,
		&stats.UniqueModels,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query organization stats: %w", err)
	}

	stats.OrganizationID = orgID
	stats.PeriodStart = start
	stats.PeriodEnd = end
	stats.AvgDuration = time.Duration(avgDurationNs)

	// Calculate derived stats
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests)
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalRequests)
	}

	return stats, nil
}

// OrganizationStats represents aggregated statistics for an organization
type OrganizationStats struct {
	OrganizationID     ulid.ULID     `json:"organization_id"`
	PeriodStart        time.Time     `json:"period_start"`
	PeriodEnd          time.Time     `json:"period_end"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	SuccessRate        float64       `json:"success_rate"`
	TotalInputTokens   int64         `json:"total_input_tokens"`
	TotalOutputTokens  int64         `json:"total_output_tokens"`
	TotalTokens        int64         `json:"total_tokens"`
	TotalCost          float64       `json:"total_cost"`
	AvgDuration        time.Duration `json:"avg_duration"`
	CacheHits          int64         `json:"cache_hits"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	UniqueProviders    int64         `json:"unique_providers"`
	UniqueModels       int64         `json:"unique_models"`
}

// Health check
func (r *Repository) GetHealth(ctx context.Context) error {
	query := "SELECT 1"
	var result int

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.conn.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return fmt.Errorf("analytics repository health check failed: %w", err)
	}

	return nil
}