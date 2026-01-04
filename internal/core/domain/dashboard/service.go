package dashboard

import (
	"context"

	"brokle/pkg/ulid"
)

// DashboardService defines the dashboard management service interface.
type DashboardService interface {
	// Dashboard CRUD operations
	CreateDashboard(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *CreateDashboardRequest) (*Dashboard, error)
	GetDashboard(ctx context.Context, id ulid.ULID) (*Dashboard, error)
	GetDashboardByProject(ctx context.Context, projectID, dashboardID ulid.ULID) (*Dashboard, error)
	UpdateDashboard(ctx context.Context, projectID, dashboardID ulid.ULID, req *UpdateDashboardRequest) (*Dashboard, error)
	DeleteDashboard(ctx context.Context, projectID, dashboardID ulid.ULID) error

	// List operations
	ListDashboards(ctx context.Context, projectID ulid.ULID, filter *DashboardFilter) (*DashboardListResponse, error)

	// Widget operations
	AddWidget(ctx context.Context, projectID, dashboardID ulid.ULID, widget *Widget) (*Dashboard, error)
	UpdateWidget(ctx context.Context, projectID, dashboardID ulid.ULID, widgetID string, widget *Widget) (*Dashboard, error)
	RemoveWidget(ctx context.Context, projectID, dashboardID ulid.ULID, widgetID string) (*Dashboard, error)

	// Layout operations
	UpdateLayout(ctx context.Context, projectID, dashboardID ulid.ULID, layout []LayoutItem) (*Dashboard, error)

	// Duplication
	DuplicateDashboard(ctx context.Context, projectID, dashboardID ulid.ULID, req *DuplicateDashboardRequest) (*Dashboard, error)

	// Lock operations
	LockDashboard(ctx context.Context, projectID, dashboardID ulid.ULID) (*Dashboard, error)
	UnlockDashboard(ctx context.Context, projectID, dashboardID ulid.ULID) (*Dashboard, error)

	// Export/Import operations
	ExportDashboard(ctx context.Context, projectID, dashboardID ulid.ULID) (*DashboardExport, error)
	ImportDashboard(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *DashboardImportRequest) (*Dashboard, error)

	// Validation
	ValidateDashboardConfig(config *DashboardConfig) error
	ValidateWidgetQuery(query *WidgetQuery) error
}

// TemplateService defines the template management service interface.
type TemplateService interface {
	// ListTemplates retrieves all active templates.
	ListTemplates(ctx context.Context) ([]*Template, error)

	// GetTemplate retrieves a template by ID.
	GetTemplate(ctx context.Context, id ulid.ULID) (*Template, error)

	// CreateFromTemplate creates a new dashboard from a template.
	CreateFromTemplate(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *CreateFromTemplateRequest) (*Dashboard, error)
}
