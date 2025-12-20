package credentials

import (
	"context"

	"brokle/pkg/ulid"
)

// ProviderCredentialRepository defines the repository interface for provider credentials.
type ProviderCredentialRepository interface {
	// Create creates a new provider credential.
	// Returns ErrCredentialExists if a credential with the same name already exists for this project.
	Create(ctx context.Context, credential *ProviderCredential) error

	// GetByID retrieves a credential by its ID within a specific project.
	// Returns ErrCredentialNotFound if not found or belongs to different project.
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*ProviderCredential, error)

	// GetByProjectAndName retrieves the credential for a specific project and name.
	// Returns nil if not found.
	GetByProjectAndName(ctx context.Context, projectID ulid.ULID, name string) (*ProviderCredential, error)

	// GetByProjectAndAdapter retrieves all credentials for a specific project and adapter type.
	// Returns empty slice if none found.
	GetByProjectAndAdapter(ctx context.Context, projectID ulid.ULID, adapter Provider) ([]*ProviderCredential, error)

	// ListByProject retrieves all credentials for a project.
	// Returns empty slice if no credentials configured.
	ListByProject(ctx context.Context, projectID ulid.ULID) ([]*ProviderCredential, error)

	// Update updates an existing credential within a specific project.
	// The ID field must be set. Returns ErrCredentialNotFound if not found or belongs to different project.
	Update(ctx context.Context, credential *ProviderCredential, projectID ulid.ULID) error

	// Delete removes a credential by ID within a specific project.
	// Returns ErrCredentialNotFound if the credential doesn't exist or belongs to different project.
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error

	// ExistsByProjectAndName checks if a credential exists for a project/name combination.
	ExistsByProjectAndName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}
