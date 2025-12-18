package credentials

import (
	"context"

	"brokle/pkg/ulid"
)

// LLMProviderCredentialRepository defines the repository interface for LLM provider credentials.
type LLMProviderCredentialRepository interface {
	// Create creates a new LLM provider credential.
	// Returns ErrCredentialExists if a credential already exists for this project/provider.
	Create(ctx context.Context, credential *LLMProviderCredential) error

	// GetByID retrieves a credential by its ID.
	// Returns ErrCredentialNotFound if not found.
	GetByID(ctx context.Context, id ulid.ULID) (*LLMProviderCredential, error)

	// GetByProjectAndProvider retrieves the credential for a specific project and provider.
	// Returns ErrCredentialNotFound if not found.
	GetByProjectAndProvider(ctx context.Context, projectID ulid.ULID, provider LLMProvider) (*LLMProviderCredential, error)

	// ListByProject retrieves all credentials for a project.
	// Returns empty slice if no credentials configured.
	ListByProject(ctx context.Context, projectID ulid.ULID) ([]*LLMProviderCredential, error)

	// Update updates an existing credential.
	// The ID field must be set.
	Update(ctx context.Context, credential *LLMProviderCredential) error

	// Delete removes a credential by ID.
	// Returns ErrCredentialNotFound if the credential doesn't exist.
	Delete(ctx context.Context, id ulid.ULID) error

	// DeleteByProjectAndProvider removes a credential by project and provider.
	// Returns ErrCredentialNotFound if the credential doesn't exist.
	DeleteByProjectAndProvider(ctx context.Context, projectID ulid.ULID, provider LLMProvider) error

	// ExistsByProjectAndProvider checks if a credential exists for a project/provider.
	Exists(ctx context.Context, projectID ulid.ULID, provider LLMProvider) (bool, error)
}
