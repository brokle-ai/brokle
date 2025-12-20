package credentials

import (
	"context"
	"log/slog"

	"brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// ModelCatalogService provides model selection for the playground.
// Combines default models from provider_models table with custom user-defined models.
type ModelCatalogService interface {
	// GetAvailableModels returns all available models for a project based on configured providers.
	// Standard providers (openai, anthropic, etc.): default models from DB + custom_models
	// Custom provider: only custom_models (no defaults exist)
	GetAvailableModels(ctx context.Context, projectID ulid.ULID) ([]*analytics.AvailableModel, error)
}

type modelCatalogServiceImpl struct {
	credentialRepo credentialsDomain.ProviderCredentialRepository
	modelRepo      analytics.ProviderModelRepository
	logger         *slog.Logger
}

func NewModelCatalogService(
	credentialRepo credentialsDomain.ProviderCredentialRepository,
	modelRepo analytics.ProviderModelRepository,
	logger *slog.Logger,
) ModelCatalogService {
	return &modelCatalogServiceImpl{
		credentialRepo: credentialRepo,
		modelRepo:      modelRepo,
		logger:         logger,
	}
}

// GetAvailableModels returns all available models for a project based on configured providers.
// For multiple credentials of the same provider, includes credential info to allow selection.
func (s *modelCatalogServiceImpl) GetAvailableModels(
	ctx context.Context,
	projectID ulid.ULID,
) ([]*analytics.AvailableModel, error) {
	// 1. Get all credentials for this project
	credentials, err := s.credentialRepo.ListByProject(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to list credentials",
			"error", err,
			"project_id", projectID,
		)
		return nil, appErrors.NewInternalError("Failed to list credentials", err)
	}

	if len(credentials) == 0 {
		return []*analytics.AvailableModel{}, nil
	}

	// 2. Group credentials by provider
	providerCredentials := make(map[string][]*credentialsDomain.ProviderCredential)
	for _, cred := range credentials {
		adapter := string(cred.Adapter)
		providerCredentials[adapter] = append(providerCredentials[adapter], cred)
	}

	var result []*analytics.AvailableModel
	var standardProviders []string
	seenStandardProviders := make(map[string]bool)

	// 3. Process each credential for custom models
	for _, cred := range credentials {
		adapter := string(cred.Adapter)
		credIDStr := cred.ID.String()

		// Custom models: ALWAYS include credential info (they belong to specific credential)
		for _, customModel := range cred.CustomModels {
			result = append(result, &analytics.AvailableModel{
				ID:             customModel,
				Name:           customModel,
				Provider:       adapter,
				CredentialID:   &credIDStr,
				CredentialName: &cred.Name,
				IsCustom:       true,
			})
		}

		// Track standard providers (for default model lookup)
		if cred.Adapter != credentialsDomain.ProviderCustom && !seenStandardProviders[adapter] {
			standardProviders = append(standardProviders, adapter)
			seenStandardProviders[adapter] = true
		}
	}

	// 4. Fetch default models for standard providers
	if len(standardProviders) > 0 {
		defaultModels, err := s.modelRepo.ListByProviders(ctx, standardProviders)
		if err != nil {
			s.logger.Error("failed to fetch default models",
				"error", err,
				"providers", standardProviders,
			)
			return nil, appErrors.NewInternalError("Failed to fetch default models", err)
		}

		for _, m := range defaultModels {
			displayName := m.ModelName
			if m.DisplayName != nil && *m.DisplayName != "" {
				displayName = *m.DisplayName
			}

			// Always include credential info for each credential of this provider
			for _, cred := range providerCredentials[m.Provider] {
				credIDStr := cred.ID.String()
				result = append(result, &analytics.AvailableModel{
					ID:             m.ModelName,
					Name:           displayName,
					Provider:       m.Provider,
					CredentialID:   &credIDStr,
					CredentialName: &cred.Name,
					IsCustom:       false,
				})
			}
		}
	}

	return result, nil
}
