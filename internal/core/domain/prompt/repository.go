package prompt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"brokle/pkg/pagination"
)

// PromptRepository defines the interface for prompt data access.
type PromptRepository interface {
	Create(ctx context.Context, prompt *Prompt) error
	GetByID(ctx context.Context, id uuid.UUID) (*Prompt, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Prompt, error)
	Update(ctx context.Context, prompt *Prompt) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	ListByProject(ctx context.Context, projectID uuid.UUID, filters *PromptFilters) ([]*Prompt, int64, error)
}

// VersionRepository defines the interface for prompt version data access.
// GetNextVersionNumber is invoked from the create-version transaction
// (FOR UPDATE locks held until commit serialise concurrent inserts).
type VersionRepository interface {
	Create(ctx context.Context, version *Version) error
	GetByID(ctx context.Context, id uuid.UUID) (*Version, error)
	GetByPromptAndVersion(ctx context.Context, promptID uuid.UUID, version int) (*Version, error)
	GetByIDs(ctx context.Context, versionIDs []uuid.UUID) ([]*Version, error)
	GetLatestByPrompts(ctx context.Context, promptIDs []uuid.UUID) ([]*Version, error)
	Delete(ctx context.Context, id uuid.UUID) error

	ListByPrompt(ctx context.Context, promptID uuid.UUID) ([]*Version, error)
	GetNextVersionNumber(ctx context.Context, promptID uuid.UUID) (int, error)
}

// LabelRepository defines the interface for prompt label data access.
// SetLabel/RemoveLabel are atomic operations that internally read + write
// inside a single transaction; callers think of them as mutations.
type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	GetByID(ctx context.Context, id uuid.UUID) (*Label, error)
	GetByPromptAndName(ctx context.Context, promptID uuid.UUID, name string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id uuid.UUID) error

	ListByPrompt(ctx context.Context, promptID uuid.UUID) ([]*Label, error)
	ListByPrompts(ctx context.Context, promptIDs []uuid.UUID) ([]*Label, error)
	ListByVersion(ctx context.Context, versionID uuid.UUID) ([]*Label, error)
	ListByVersions(ctx context.Context, versionIDs []uuid.UUID) ([]*Label, error)

	SetLabel(ctx context.Context, promptID, versionID uuid.UUID, name string, createdBy *uuid.UUID) error
	RemoveLabel(ctx context.Context, promptID uuid.UUID, name string) error
}

// ProtectedLabelRepository defines the interface for protected-label data
// access. IsProtected is the hot-path check called on every label
// mutation in the prompt service.
type ProtectedLabelRepository interface {
	Create(ctx context.Context, label *ProtectedLabel) error
	Delete(ctx context.Context, id uuid.UUID) error

	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*ProtectedLabel, error)
	IsProtected(ctx context.Context, projectID uuid.UUID, labelName string) (bool, error)
	SetProtectedLabels(ctx context.Context, projectID uuid.UUID, labels []string, createdBy *uuid.UUID) error
}

// CacheRepository defines the interface for prompt caching.
type CacheRepository interface {
	Get(ctx context.Context, key string) (*CachedPrompt, error)
	Set(ctx context.Context, key string, prompt *CachedPrompt, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	BuildKey(projectID uuid.UUID, name string, labelOrVersion string) string
}

// CachedPrompt represents a prompt stored in the cache.
type CachedPrompt struct {
	PromptID      uuid.UUID    `json:"prompt_id"`
	ProjectID     uuid.UUID    `json:"project_id"`
	Name          string       `json:"name"`
	Type          PromptType   `json:"type"`
	Description   string       `json:"description"`
	Tags          []string     `json:"tags"`
	Version       int          `json:"version"`
	VersionID     uuid.UUID    `json:"version_id"`
	Labels        []string     `json:"labels"`
	Template      any          `json:"template"`
	Config        *ModelConfig `json:"config,omitempty"`
	Variables     []string     `json:"variables"`
	CommitMessage string       `json:"commit_message"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CreatedBy     *uuid.UUID   `json:"created_by,omitempty"`
	CachedAt      time.Time    `json:"cached_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
}

// IsExpired reports whether the cached prompt has expired.
func (c *CachedPrompt) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsStale reports whether the cached prompt is stale but still usable.
func (c *CachedPrompt) IsStale(ttl time.Duration) bool {
	staleAt := c.CachedAt.Add(ttl)
	return time.Now().After(staleAt)
}

// PromptFilters represents filters for prompt queries.
type PromptFilters struct {
	Type   *PromptType
	Tags   []string
	Search *string

	pagination.Params
}

// VersionFilters represents filters for version queries.
type VersionFilters struct {
	PromptID *uuid.UUID

	pagination.Params
}
