package dashboard

import (
	"context"

	"brokle/pkg/ulid"
)

// DashboardRepository defines the interface for dashboard data access.
type DashboardRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, dashboard *Dashboard) error
	GetByID(ctx context.Context, id ulid.ULID) (*Dashboard, error)
	Update(ctx context.Context, dashboard *Dashboard) error
	Delete(ctx context.Context, id ulid.ULID) error

	// Project-scoped queries
	GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *DashboardFilter) (*DashboardListResponse, error)
	GetByNameAndProject(ctx context.Context, projectID ulid.ULID, name string) (*Dashboard, error)

	// Soft delete operations
	SoftDelete(ctx context.Context, id ulid.ULID) error

	// Count operations
	CountByProject(ctx context.Context, projectID ulid.ULID) (int64, error)
}

// TemplateRepository defines the interface for dashboard template data access.
type TemplateRepository interface {
	// List retrieves all active templates with optional filtering.
	List(ctx context.Context, filter *TemplateFilter) ([]*Template, error)

	// GetByID retrieves a template by its ID.
	GetByID(ctx context.Context, id ulid.ULID) (*Template, error)

	// GetByName retrieves a template by its name.
	GetByName(ctx context.Context, name string) (*Template, error)

	// GetByCategory retrieves a template by its category.
	GetByCategory(ctx context.Context, category TemplateCategory) (*Template, error)

	// Create creates a new template (used for seeding).
	Create(ctx context.Context, template *Template) error

	// Update updates an existing template.
	Update(ctx context.Context, template *Template) error

	// Delete removes a template by its ID.
	Delete(ctx context.Context, id ulid.ULID) error

	// Upsert creates or updates a template by name (used for seeding).
	Upsert(ctx context.Context, template *Template) error
}
