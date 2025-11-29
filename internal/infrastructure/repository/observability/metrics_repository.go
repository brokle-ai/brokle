package observability

import (
	"context"
	"fmt"

	"brokle/internal/core/domain/observability"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// metricsRepository implements ClickHouse persistence for OTLP metrics
type metricsRepository struct {
	db clickhouse.Conn
}

// NewMetricsRepository creates a new metrics repository instance
func NewMetricsRepository(db clickhouse.Conn) observability.MetricsRepository {
	return &metricsRepository{db: db}
}

// ===== MetricSum Repository Methods =====

// CreateMetricSumBatch inserts multiple metric sums in a single batch
func (r *metricsRepository) CreateMetricSumBatch(ctx context.Context, metricsSums []*observability.MetricSum) error {
	if len(metricsSums) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO otel_metrics_sum (
			resource_attributes,
			scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			metric_name, metric_description, metric_unit,
			attributes,
			start_time_unix, time_unix,
			value,
			aggregation_temporality, is_monotonic,
			exemplars_timestamp, exemplars_value, exemplars_filtered_attributes,
			exemplars_trace_id, exemplars_span_id,
			project_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, metricSum := range metricsSums {
		err = batch.Append(
			metricSum.ResourceAttributes,
			metricSum.ScopeName,
			metricSum.ScopeVersion,
			metricSum.ScopeAttributes,
			metricSum.ResourceSchemaURL,
			metricSum.ScopeSchemaURL,
			metricSum.MetricName,
			metricSum.MetricDescription,
			metricSum.MetricUnit,
			metricSum.Attributes,
			metricSum.StartTimeUnix,
			metricSum.TimeUnix,
			metricSum.Value,
			metricSum.AggregationTemporality,
			metricSum.IsMonotonic,
			metricSum.ExemplarsTimestamp,
			metricSum.ExemplarsValue,
			metricSum.ExemplarsFilteredAttributes,
			metricSum.ExemplarsTraceID,
			metricSum.ExemplarsSpanID,
			metricSum.ProjectID,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// ===== MetricGauge Repository Methods =====

// CreateMetricGaugeBatch inserts multiple metric gauges in a single batch
func (r *metricsRepository) CreateMetricGaugeBatch(ctx context.Context, metricsGauges []*observability.MetricGauge) error {
	if len(metricsGauges) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO otel_metrics_gauge (
			resource_attributes,
			scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			metric_name, metric_description, metric_unit,
			attributes,
			start_time_unix, time_unix,
			value,
			exemplars_timestamp, exemplars_value, exemplars_filtered_attributes,
			exemplars_trace_id, exemplars_span_id,
			project_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, metricGauge := range metricsGauges {
		err = batch.Append(
			metricGauge.ResourceAttributes,
			metricGauge.ScopeName,
			metricGauge.ScopeVersion,
			metricGauge.ScopeAttributes,
			metricGauge.ResourceSchemaURL,
			metricGauge.ScopeSchemaURL,
			metricGauge.MetricName,
			metricGauge.MetricDescription,
			metricGauge.MetricUnit,
			metricGauge.Attributes,
			metricGauge.StartTimeUnix,
			metricGauge.TimeUnix,
			metricGauge.Value,
			metricGauge.ExemplarsTimestamp,
			metricGauge.ExemplarsValue,
			metricGauge.ExemplarsFilteredAttributes,
			metricGauge.ExemplarsTraceID,
			metricGauge.ExemplarsSpanID,
			metricGauge.ProjectID,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// ===== MetricHistogram Repository Methods =====

// CreateMetricHistogramBatch inserts multiple metric histograms in a single batch
func (r *metricsRepository) CreateMetricHistogramBatch(ctx context.Context, metricsHistograms []*observability.MetricHistogram) error {
	if len(metricsHistograms) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO otel_metrics_histogram (
			resource_attributes,
			scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			metric_name, metric_description, metric_unit,
			attributes,
			start_time_unix, time_unix,
			count, sum, min, max,
			bucket_counts, explicit_bounds,
			aggregation_temporality,
			exemplars_timestamp, exemplars_value, exemplars_filtered_attributes,
			exemplars_trace_id, exemplars_span_id,
			project_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, metricHistogram := range metricsHistograms {
		err = batch.Append(
			metricHistogram.ResourceAttributes,
			metricHistogram.ScopeName,
			metricHistogram.ScopeVersion,
			metricHistogram.ScopeAttributes,
			metricHistogram.ResourceSchemaURL,
			metricHistogram.ScopeSchemaURL,
			metricHistogram.MetricName,
			metricHistogram.MetricDescription,
			metricHistogram.MetricUnit,
			metricHistogram.Attributes,
			metricHistogram.StartTimeUnix,
			metricHistogram.TimeUnix,
			metricHistogram.Count,
			metricHistogram.Sum, // Nullable(Float64) - pointer type
			metricHistogram.Min, // Nullable(Float64) - pointer type
			metricHistogram.Max, // Nullable(Float64) - pointer type
			metricHistogram.BucketCounts,
			metricHistogram.ExplicitBounds,
			metricHistogram.AggregationTemporality,
			metricHistogram.ExemplarsTimestamp,
			metricHistogram.ExemplarsValue,
			metricHistogram.ExemplarsFilteredAttributes,
			metricHistogram.ExemplarsTraceID,
			metricHistogram.ExemplarsSpanID,
			metricHistogram.ProjectID,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// ===== MetricExponentialHistogram Repository Methods =====

// CreateMetricExponentialHistogramBatch inserts multiple exponential histogram metrics in a single batch
// OTLP 1.38+: Modern histogram using exponential bucketing (memory-efficient)
func (r *metricsRepository) CreateMetricExponentialHistogramBatch(ctx context.Context, metricsExpHistograms []*observability.MetricExponentialHistogram) error {
	if len(metricsExpHistograms) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO otel_metrics_exponential_histogram (
			resource_attributes,
			scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			metric_name, metric_description, metric_unit,
			attributes,
			start_time_unix, time_unix,
			count, sum,
			scale, zero_count,
			positive_offset, positive_bucket_counts,
			negative_offset, negative_bucket_counts,
			min, max,
			aggregation_temporality,
			exemplars_timestamp, exemplars_value, exemplars_filtered_attributes,
			exemplars_trace_id, exemplars_span_id,
			project_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, metricExpHistogram := range metricsExpHistograms {
		err = batch.Append(
			metricExpHistogram.ResourceAttributes,
			metricExpHistogram.ScopeName,
			metricExpHistogram.ScopeVersion,
			metricExpHistogram.ScopeAttributes,
			metricExpHistogram.ResourceSchemaURL,
			metricExpHistogram.ScopeSchemaURL,
			metricExpHistogram.MetricName,
			metricExpHistogram.MetricDescription,
			metricExpHistogram.MetricUnit,
			metricExpHistogram.Attributes,
			metricExpHistogram.StartTimeUnix,
			metricExpHistogram.TimeUnix,
			metricExpHistogram.Count,
			metricExpHistogram.Sum, // Nullable(Float64) - pointer type
			metricExpHistogram.Scale,
			metricExpHistogram.ZeroCount,
			metricExpHistogram.PositiveOffset,
			metricExpHistogram.PositiveBucketCounts,
			metricExpHistogram.NegativeOffset,
			metricExpHistogram.NegativeBucketCounts,
			metricExpHistogram.Min, // Nullable(Float64) - pointer type
			metricExpHistogram.Max, // Nullable(Float64) - pointer type
			metricExpHistogram.AggregationTemporality,
			metricExpHistogram.ExemplarsTimestamp,
			metricExpHistogram.ExemplarsValue,
			metricExpHistogram.ExemplarsFilteredAttributes,
			metricExpHistogram.ExemplarsTraceID,
			metricExpHistogram.ExemplarsSpanID,
			metricExpHistogram.ProjectID,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}
