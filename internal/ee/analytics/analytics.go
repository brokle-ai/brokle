package analytics

import (
	"context"
	"errors"
	"time"
)

// EnterpriseAnalytics interface for advanced analytics features
type EnterpriseAnalytics interface {
	GeneratePredictiveInsights(ctx context.Context, timeRange string) (*PredictiveReport, error)
	CreateCustomDashboard(ctx context.Context, dashboard *Dashboard) error
	UpdateCustomDashboard(ctx context.Context, dashboardID string, dashboard *Dashboard) error
	GetCustomDashboard(ctx context.Context, dashboardID string) (*Dashboard, error)
	ListCustomDashboards(ctx context.Context) ([]*Dashboard, error)
	DeleteCustomDashboard(ctx context.Context, dashboardID string) error
	GenerateAdvancedReport(ctx context.Context, req *ReportRequest) (*Report, error)
	ExportData(ctx context.Context, format string, query *ExportQuery) ([]byte, error)
	RunMLModel(ctx context.Context, modelName string, data interface{}) (interface{}, error)
}

// PredictiveReport represents ML-powered insights
type PredictiveReport struct {
	TimeRange       string            `json:"time_range"`
	CostForecast    *CostForecast     `json:"cost_forecast,omitempty"`
	UsageTrends     []*UsageTrend     `json:"usage_trends,omitempty"`
	Anomalies       []*Anomaly        `json:"anomalies,omitempty"`
	Recommendations []*Recommendation `json:"recommendations,omitempty"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// Dashboard represents a custom dashboard configuration
type Dashboard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Widgets     []*Widget `json:"widgets"`
	Layout      *Layout   `json:"layout,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // chart, metric, table, etc.
	Title    string                 `json:"title"`
	Config   map[string]interface{} `json:"config"`
	Position *Position              `json:"position,omitempty"`
}

// Layout represents dashboard layout configuration
type Layout struct {
	Columns int    `json:"columns"`
	Rows    int    `json:"rows"`
	Theme   string `json:"theme,omitempty"`
}

// Position represents widget position in dashboard
type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ReportRequest represents advanced report parameters
type ReportRequest struct {
	Type        string                 `json:"type"`
	TimeRange   string                 `json:"time_range"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	GroupBy     []string               `json:"group_by,omitempty"`
	Aggregation string                 `json:"aggregation,omitempty"`
}

// Report represents generated report
type Report struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   time.Time              `json:"expires_at,omitempty"`
}

// ExportQuery represents data export parameters
type ExportQuery struct {
	Table     string                 `json:"table"`
	TimeRange string                 `json:"time_range"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Columns   []string               `json:"columns,omitempty"`
}

// Supporting types
type CostForecast struct {
	NextMonth   float64 `json:"next_month"`
	NextQuarter float64 `json:"next_quarter"`
	Confidence  float64 `json:"confidence"`
	Trend       string  `json:"trend"` // increasing, decreasing, stable
}

type UsageTrend struct {
	Metric     string  `json:"metric"`
	Trend      string  `json:"trend"`
	Change     float64 `json:"change"` // percentage
	Confidence float64 `json:"confidence"`
	Period     string  `json:"period"`
}

type Anomaly struct {
	Metric      string    `json:"metric"`
	Timestamp   time.Time `json:"timestamp"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Severity    string    `json:"severity"` // low, medium, high
	Description string    `json:"description"`
}

type Recommendation struct {
	Type        string  `json:"type"` // cost, performance, security
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // low, medium, high
	Effort      string  `json:"effort"` // low, medium, high
	Savings     float64 `json:"savings,omitempty"`
}

// StubEnterpriseAnalytics provides stub implementation for OSS version
type StubEnterpriseAnalytics struct{}

// New returns the enterprise analytics implementation (stub or real based on build tags)
func New() EnterpriseAnalytics {
	return &StubEnterpriseAnalytics{}
}

func (s *StubEnterpriseAnalytics) GeneratePredictiveInsights(ctx context.Context, timeRange string) (*PredictiveReport, error) {
	return nil, errors.New("predictive insights require Enterprise license")
}

func (s *StubEnterpriseAnalytics) CreateCustomDashboard(ctx context.Context, dashboard *Dashboard) error {
	return errors.New("custom dashboards require Enterprise license")
}

func (s *StubEnterpriseAnalytics) UpdateCustomDashboard(ctx context.Context, dashboardID string, dashboard *Dashboard) error {
	return errors.New("custom dashboards require Enterprise license")
}

func (s *StubEnterpriseAnalytics) GetCustomDashboard(ctx context.Context, dashboardID string) (*Dashboard, error) {
	return nil, errors.New("custom dashboards require Enterprise license")
}

func (s *StubEnterpriseAnalytics) ListCustomDashboards(ctx context.Context) ([]*Dashboard, error) {
	return []*Dashboard{}, errors.New("custom dashboards require Enterprise license")
}

func (s *StubEnterpriseAnalytics) DeleteCustomDashboard(ctx context.Context, dashboardID string) error {
	return errors.New("custom dashboards require Enterprise license")
}

func (s *StubEnterpriseAnalytics) GenerateAdvancedReport(ctx context.Context, req *ReportRequest) (*Report, error) {
	return nil, errors.New("advanced reports require Enterprise license")
}

func (s *StubEnterpriseAnalytics) ExportData(ctx context.Context, format string, query *ExportQuery) ([]byte, error) {
	return nil, errors.New("data export requires Enterprise license")
}

func (s *StubEnterpriseAnalytics) RunMLModel(ctx context.Context, modelName string, data interface{}) (interface{}, error) {
	return nil, errors.New("ML models require Enterprise license")
}
