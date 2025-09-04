package analytics

import (
	"time"
)

// AI Platform specific data structures

// AIRequest represents an AI request for analytics
type AIRequest struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	TokensUsed  int       `json:"tokens_used"`
	Cost        float64   `json:"cost"`
	Latency     float64   `json:"latency"`
	Quality     float64   `json:"quality"`
	Success     bool      `json:"success"`
	CacheHit    bool      `json:"cache_hit"`
	UserID      string    `json:"user_id,omitempty"`
	OrgID       string    `json:"org_id"`
	ProjectID   string    `json:"project_id"`
	Environment string    `json:"environment"`
}

// CacheEvent represents a cache-related event
type CacheEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Key         string    `json:"key"`
	Hit         bool      `json:"hit"`
	SavedCost   float64   `json:"saved_cost"`
	SavedTime   float64   `json:"saved_time"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	Similarity  float64   `json:"similarity,omitempty"`
}

// RoutingDecision represents an AI routing decision
type RoutingDecision struct {
	ID                   string    `json:"id"`
	Timestamp            time.Time `json:"timestamp"`
	SelectedProvider     string    `json:"selected_provider"`
	Reason               string    `json:"reason"`
	DecisionLatency      float64   `json:"decision_latency"`
	RoutedRequestCost    float64   `json:"routed_request_cost"`
	AccuracyScore        float64   `json:"accuracy_score"`
	AlternativeProviders []string  `json:"alternative_providers,omitempty"`
}

// Analytics metrics structures

// CostEfficiencyMetrics represents cost efficiency analytics
type CostEfficiencyMetrics struct {
	TotalCost           float64            `json:"total_cost"`
	TotalTokens         int64              `json:"total_tokens"`
	TotalRequests       int64              `json:"total_requests"`
	AvgCostPerRequest   float64            `json:"avg_cost_per_request"`
	AvgCostPerToken     float64            `json:"avg_cost_per_token"`
	TokensPerRequest    float64            `json:"tokens_per_request"`
	CostByProvider      map[string]float64 `json:"cost_by_provider"`
	CostByModel         map[string]float64 `json:"cost_by_model"`
	EfficiencyScore     float64            `json:"efficiency_score"`
}

// ProviderMetrics represents provider performance metrics
type ProviderMetrics struct {
	Provider      string  `json:"provider"`
	TotalRequests int64   `json:"total_requests"`
	AvgLatency    float64 `json:"avg_latency"`
	AvgCost       float64 `json:"avg_cost"`
	AvgQuality    float64 `json:"avg_quality"`
	SuccessRate   float64 `json:"success_rate"`
	ErrorRate     float64 `json:"error_rate"`
	P95Latency    float64 `json:"p95_latency"`
	TotalCost     float64 `json:"total_cost"`
}

// UsageMetrics represents usage analytics over time
type UsageMetrics struct {
	TimeWindow          TimeWindow `json:"time_window"`
	TotalRequests       int64      `json:"total_requests"`
	TotalTokens         int64      `json:"total_tokens"`
	TotalCost           float64    `json:"total_cost"`
	AvgTokensPerRequest float64    `json:"avg_tokens_per_request"`
	AvgCostPerRequest   float64    `json:"avg_cost_per_request"`
	RequestsPerHour     float64    `json:"requests_per_hour"`
	PeakHour            int        `json:"peak_hour"`
	PeakDay             string     `json:"peak_day"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	TotalRequests   int64   `json:"total_requests"`
	CacheHits       int64   `json:"cache_hits"`
	CacheMisses     int64   `json:"cache_misses"`
	HitRate         float64 `json:"hit_rate"`
	MissRate        float64 `json:"miss_rate"`
	TotalCostSaved  float64 `json:"total_cost_saved"`
	TotalTimeSaved  float64 `json:"total_time_saved"`
	AvgCostSaved    float64 `json:"avg_cost_saved"`
	AvgTimeSaved    float64 `json:"avg_time_saved"`
}

// QualityMetrics represents AI response quality analytics
type QualityMetrics struct {
	OverallAverage     float64            `json:"overall_average"`
	OverallMedian      float64            `json:"overall_median"`
	StandardDeviation  float64            `json:"standard_deviation"`
	MinScore           float64            `json:"min_score"`
	MaxScore           float64            `json:"max_score"`
	P25Score           float64            `json:"p25_score"`
	P75Score           float64            `json:"p75_score"`
	P90Score           float64            `json:"p90_score"`
	P95Score           float64            `json:"p95_score"`
	QualityByProvider  map[string]float64 `json:"quality_by_provider"`
	QualityByModel     map[string]float64 `json:"quality_by_model"`
}

// RoutingMetrics represents routing decision analytics
type RoutingMetrics struct {
	TotalDecisions       int64             `json:"total_decisions"`
	AvgDecisionLatency   float64           `json:"avg_decision_latency"`
	AvgRoutedCost        float64           `json:"avg_routed_cost"`
	AvgAccuracy          float64           `json:"avg_accuracy"`
	RoutingReasons       map[string]int64  `json:"routing_reasons"`
	ProviderSelections   map[string]int64  `json:"provider_selections"`
	EfficiencyScore      float64           `json:"efficiency_score"`
}

// Business Intelligence structures

// GrowthMetrics represents growth analytics between periods
type GrowthMetrics struct {
	RequestsGrowth      float64       `json:"requests_growth"`
	TokensGrowth        float64       `json:"tokens_growth"`
	CostGrowth          float64       `json:"cost_growth"`
	QualityImprovement  float64       `json:"quality_improvement"`
	LatencyImprovement  float64       `json:"latency_improvement"`
	CurrentPeriod       PeriodMetrics `json:"current_period"`
	PreviousPeriod      PeriodMetrics `json:"previous_period"`
}

// PeriodMetrics represents metrics for a specific time period
type PeriodMetrics struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	AvgLatency    float64 `json:"avg_latency"`
	AvgQuality    float64 `json:"avg_quality"`
}

// Anomaly represents a detected anomaly in metrics
type Anomaly struct {
	Index       int     `json:"index"`
	Value       float64 `json:"value"`
	Expected    float64 `json:"expected"`
	Deviation   float64 `json:"deviation"`
	Severity    string  `json:"severity"`
	ZScore      float64 `json:"z_score"`
}

// TrendAnalysis represents trend analysis of time series data
type TrendAnalysis struct {
	Direction   string  `json:"direction"`
	Slope       float64 `json:"slope"`
	Intercept   float64 `json:"intercept"`
	Correlation float64 `json:"correlation"`
	Strength    string  `json:"strength"`
	DataPoints  int     `json:"data_points"`
}

// Advanced Analytics structures

// SegmentationResult represents user/usage segmentation results
type SegmentationResult struct {
	SegmentName   string                 `json:"segment_name"`
	Size          int64                  `json:"size"`
	Percentage    float64                `json:"percentage"`
	Metrics       map[string]interface{} `json:"metrics"`
	Characteristics []string             `json:"characteristics"`
}

// PredictionResult represents prediction model results
type PredictionResult struct {
	Metric          string    `json:"metric"`
	PredictedValue  float64   `json:"predicted_value"`
	Confidence      float64   `json:"confidence"`
	PredictionDate  time.Time `json:"prediction_date"`
	ModelAccuracy   float64   `json:"model_accuracy"`
	Factors         []string  `json:"factors"`
}

// ComparisonResult represents comparison between time periods or segments
type ComparisonResult struct {
	Metric        string  `json:"metric"`
	BaselineValue float64 `json:"baseline_value"`
	CompareValue  float64 `json:"compare_value"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Significance  string  `json:"significance"`
}

// Dashboard and Reporting structures

// DashboardMetrics represents metrics for dashboard display
type DashboardMetrics struct {
	Summary        *SummaryMetrics      `json:"summary"`
	Trends         []*TrendMetric       `json:"trends"`
	TopProviders   []*ProviderMetrics   `json:"top_providers"`
	RecentAlerts   []*Alert             `json:"recent_alerts"`
	UsageBreakdown *UsageBreakdown      `json:"usage_breakdown"`
	CostAnalysis   *CostAnalysis        `json:"cost_analysis"`
	LastUpdated    time.Time            `json:"last_updated"`
}

// SummaryMetrics represents high-level summary metrics
type SummaryMetrics struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalCost         float64 `json:"total_cost"`
	AvgLatency        float64 `json:"avg_latency"`
	SuccessRate       float64 `json:"success_rate"`
	CacheHitRate      float64 `json:"cache_hit_rate"`
	CostSavings       float64 `json:"cost_savings"`
	RequestsChange    float64 `json:"requests_change"`
	CostChange        float64 `json:"cost_change"`
	PerformanceChange float64 `json:"performance_change"`
}

// TrendMetric represents a trending metric
type TrendMetric struct {
	Name        string        `json:"name"`
	Value       float64       `json:"value"`
	Change      float64       `json:"change"`
	Trend       string        `json:"trend"`
	Sparkline   []float64     `json:"sparkline"`
	Unit        string        `json:"unit"`
	TimePeriod  string        `json:"time_period"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UsageBreakdown represents usage breakdown by different dimensions
type UsageBreakdown struct {
	ByProvider    map[string]int64 `json:"by_provider"`
	ByModel       map[string]int64 `json:"by_model"`
	ByEnvironment map[string]int64 `json:"by_environment"`
	ByHour        map[int]int64    `json:"by_hour"`
	ByDay         map[string]int64 `json:"by_day"`
	ByUser        map[string]int64 `json:"by_user,omitempty"`
}

// CostAnalysis represents detailed cost analysis
type CostAnalysis struct {
	TotalCost        float64            `json:"total_cost"`
	CostByProvider   map[string]float64 `json:"cost_by_provider"`
	CostByModel      map[string]float64 `json:"cost_by_model"`
	CostTrend        []float64          `json:"cost_trend"`
	PredictedCost    float64            `json:"predicted_cost"`
	CostOptimization *CostOptimization  `json:"cost_optimization"`
}

// CostOptimization represents cost optimization suggestions
type CostOptimization struct {
	PotentialSavings   float64              `json:"potential_savings"`
	Recommendations    []*Recommendation    `json:"recommendations"`
	ProviderMigration  []*MigrationSuggestion `json:"provider_migration"`
	CacheOptimization  *CacheOptimization   `json:"cache_optimization"`
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Impact          string                 `json:"impact"`
	Effort          string                 `json:"effort"`
	PotentialSaving float64                `json:"potential_saving"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// MigrationSuggestion represents a provider migration suggestion
type MigrationSuggestion struct {
	FromProvider    string  `json:"from_provider"`
	ToProvider      string  `json:"to_provider"`
	Model           string  `json:"model,omitempty"`
	PotentialSaving float64 `json:"potential_saving"`
	QualityImpact   float64 `json:"quality_impact"`
	Confidence      float64 `json:"confidence"`
	Reason          string  `json:"reason"`
}

// CacheOptimization represents cache optimization analysis
type CacheOptimization struct {
	CurrentHitRate      float64 `json:"current_hit_rate"`
	OptimalHitRate      float64 `json:"optimal_hit_rate"`
	AdditionalSavings   float64 `json:"additional_savings"`
	RecommendedTTL      int     `json:"recommended_ttl"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

// Export and reporting structures

// ReportConfig represents configuration for generating reports
type ReportConfig struct {
	Type        string                 `json:"type"`
	TimeRange   TimeWindow             `json:"time_range"`
	Metrics     []string               `json:"metrics"`
	GroupBy     []string               `json:"group_by"`
	Filters     map[string]interface{} `json:"filters"`
	Format      string                 `json:"format"`
	Recipients  []string               `json:"recipients,omitempty"`
	Schedule    string                 `json:"schedule,omitempty"`
}

// Report represents a generated analytics report
type Report struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	TimeRange   TimeWindow             `json:"time_range"`
	Data        interface{}            `json:"data"`
	Summary     string                 `json:"summary"`
	Insights    []string               `json:"insights"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Query builder structures

// AnalyticsQuery represents a structured analytics query
type AnalyticsQuery struct {
	Select    []string               `json:"select"`
	From      string                 `json:"from"`
	Where     map[string]interface{} `json:"where,omitempty"`
	GroupBy   []string               `json:"group_by,omitempty"`
	OrderBy   []string               `json:"order_by,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	TimeRange *TimeWindow            `json:"time_range,omitempty"`
}

// QueryResult represents the result of an analytics query
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int64           `json:"count"`
	Query   string          `json:"query,omitempty"`
	Took    time.Duration   `json:"took"`
}

// Helper types for common patterns

// MetricValue represents a single metric value with metadata
type MetricValue struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Threshold represents a metric threshold for alerting
type Threshold struct {
	Metric       string  `json:"metric"`
	Operator     string  `json:"operator"` // >, <, >=, <=, ==, !=
	Value        float64 `json:"value"`
	Window       string  `json:"window"`
	Severity     string  `json:"severity"`
	Description  string  `json:"description,omitempty"`
}

// Filter represents a data filter
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SortOrder represents sort ordering
type SortOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // asc, desc
}