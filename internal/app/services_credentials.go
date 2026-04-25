package app

import (
	"fmt"
	"log/slog"

	"brokle/internal/config"
	credentialsService "brokle/internal/core/services/credentials"
	"brokle/pkg/encryption"
)

// ProvideCredentialsServices builds the credentials-domain services
// for storage of provider API keys (AES-256-GCM encryption at rest)
// and the merged model catalog. Encryption-key validity is enforced
// at config-load; a panic here means a build-time invariant was
// violated and ought to be loud.
func ProvideCredentialsServices(
	credentialsRepos *CredentialsRepositories,
	analyticsRepos *AnalyticsRepositories,
	cfg *config.Config,
	logger *slog.Logger,
) *CredentialsServices {
	encryptor, err := encryption.NewServiceFromBase64(cfg.Encryption.AIKeyEncryptionKey)
	if err != nil {
		panic(fmt.Sprintf("encryption initialization failed after config validation: %v (this is a bug)", err))
	}

	providerSvc := credentialsService.NewProviderCredentialService(
		credentialsRepos.ProviderCredential,
		encryptor,
		logger,
	)

	modelCatalogSvc := credentialsService.NewModelCatalogService(
		credentialsRepos.ProviderCredential,
		analyticsRepos.ProviderModel,
		logger,
	)

	return &CredentialsServices{
		ProviderCredential: providerSvc,
		ModelCatalog:       modelCatalogSvc,
	}
}
